package authz

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
	eventsv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/events/v1"
	"github.com/m8platform/platform/iam/internal/config"
	"github.com/m8platform/platform/iam/internal/core"
	redisstore "github.com/m8platform/platform/iam/internal/storage/redis"
	"github.com/m8platform/platform/iam/internal/storage/ydb"
	"go.uber.org/zap"
)

type Service struct {
	authzv1.UnimplementedAuthorizationFacadeServiceServer

	store         core.DocumentStore
	cache         core.Cache
	publisher     core.EventPublisher
	runtime       core.AuthorizationRuntime
	logger        *zap.Logger
	now           func() time.Time
	policyVersion string
	topics        config.TopicsConfig
}

func NewService(store core.DocumentStore, cache core.Cache, publisher core.EventPublisher, runtime core.AuthorizationRuntime, logger *zap.Logger, cfg config.Config) *Service {
	return &Service{
		store:         store,
		cache:         cache,
		publisher:     publisher,
		runtime:       runtime,
		logger:        logger,
		now:           time.Now,
		policyVersion: cfg.Redis.PolicyVersion,
		topics:        cfg.Topics,
	}
}

func (s *Service) GetRole(_ context.Context, req *authzv1.GetRoleRequest) (*authzv1.Role, error) {
	role, ok := ResolveRole(req.GetRoleId())
	if !ok {
		return nil, fmt.Errorf("role %s not found", req.GetRoleId())
	}
	return role, nil
}

func (s *Service) ListRoles(context.Context, *authzv1.ListRolesRequest) (*authzv1.ListRolesResponse, error) {
	return &authzv1.ListRolesResponse{Roles: DefaultRoles()}, nil
}

func (s *Service) SetAccessBindings(ctx context.Context, req *authzv1.SetAccessBindingsRequest) (*authzv1.SetAccessBindingsResponse, error) {
	now := s.now()
	for _, binding := range req.GetDesiredBindings() {
		if err := core.SaveProto(ctx, s.store, ydb.TableBindingOperations, binding.GetBindingId(), binding.GetResource().GetTenantId(), binding, now); err != nil {
			return nil, err
		}
	}
	if s.runtime != nil {
		if err := s.runtime.WriteBindings(ctx, req.GetDesiredBindings()); err != nil {
			s.logger.Warn("spicedb write skipped", zap.Error(err))
		}
	}
	operation := core.NewOperation(now, req.GetResource().GetTenantId(), "set_access_bindings", req.GetResource().GetType().String(), req.GetResource().GetId())
	if err := core.PersistOperation(ctx, s.store, operation, now); err != nil {
		return nil, err
	}
	audit := core.NewAuditEvent(now, req.GetResource().GetTenantId(), "access_bindings.set", req.GetPerformedBy(), operation.GetOperationId(), req.GetReason())
	audit.Resource = req.GetResource()
	if err := core.PersistAuditEvent(ctx, s.store, audit, now); err != nil {
		return nil, err
	}
	event := &eventsv1.AccessIntentChanged{
		EventId:               operation.GetOperationId(),
		OccurredAt:            core.Timestamp(now),
		TenantId:              req.GetResource().GetTenantId(),
		Resource:              req.GetResource(),
		Actor:                 req.GetPerformedBy(),
		Reason:                req.GetReason(),
		DesiredBindings:       req.GetDesiredBindings(),
		RelationshipMutations: toAddMutations(req.GetDesiredBindings()),
		CorrelationId:         operation.GetOperationId(),
	}
	if err := s.publisher.PublishProto(ctx, s.topics.Relationships, event); err != nil {
		return nil, err
	}
	return &authzv1.SetAccessBindingsResponse{
		Bindings:    req.GetDesiredBindings(),
		OperationId: operation.GetOperationId(),
	}, nil
}

func (s *Service) UpdateAccessBindings(ctx context.Context, req *authzv1.UpdateAccessBindingsRequest) (*authzv1.UpdateAccessBindingsResponse, error) {
	now := s.now()
	current, err := s.listBindingsForResource(ctx, req.GetResource())
	if err != nil {
		return nil, err
	}
	index := make(map[string]*authzv1.AccessBinding, len(current))
	for _, binding := range current {
		index[binding.GetBindingId()] = binding
	}
	for _, mutation := range req.GetDelta().GetMutations() {
		switch mutation.GetKind() {
		case authzv1.BindingMutationKind_BINDING_MUTATION_KIND_ADD:
			index[mutation.GetBinding().GetBindingId()] = mutation.GetBinding()
			if err := core.SaveProto(ctx, s.store, ydb.TableBindingOperations, mutation.GetBinding().GetBindingId(), mutation.GetBinding().GetResource().GetTenantId(), mutation.GetBinding(), now); err != nil {
				return nil, err
			}
		case authzv1.BindingMutationKind_BINDING_MUTATION_KIND_REMOVE:
			delete(index, mutation.GetBinding().GetBindingId())
			if err := s.store.DeleteDocument(ctx, ydb.TableBindingOperations, mutation.GetBinding().GetBindingId()); err != nil && err != core.ErrNotFound {
				return nil, err
			}
		}
	}
	bindings := make([]*authzv1.AccessBinding, 0, len(index))
	for _, binding := range index {
		bindings = append(bindings, binding)
	}
	operation := core.NewOperation(now, req.GetResource().GetTenantId(), "update_access_bindings", req.GetResource().GetType().String(), req.GetResource().GetId())
	if err := core.PersistOperation(ctx, s.store, operation, now); err != nil {
		return nil, err
	}
	audit := core.NewAuditEvent(now, req.GetResource().GetTenantId(), "access_bindings.updated", req.GetPerformedBy(), operation.GetOperationId(), req.GetReason())
	audit.Resource = req.GetResource()
	if err := core.PersistAuditEvent(ctx, s.store, audit, now); err != nil {
		return nil, err
	}
	event := &eventsv1.AccessIntentChanged{
		EventId:               operation.GetOperationId(),
		OccurredAt:            core.Timestamp(now),
		TenantId:              req.GetResource().GetTenantId(),
		Resource:              req.GetResource(),
		Actor:                 req.GetPerformedBy(),
		Reason:                req.GetReason(),
		DesiredBindings:       bindings,
		RelationshipMutations: req.GetDelta().GetMutations(),
		CorrelationId:         operation.GetOperationId(),
	}
	if err := s.publisher.PublishProto(ctx, s.topics.Relationships, event); err != nil {
		return nil, err
	}
	if s.runtime != nil {
		if err := s.runtime.WriteBindings(ctx, bindings); err != nil {
			s.logger.Warn("spicedb update skipped", zap.Error(err))
		}
	}
	return &authzv1.UpdateAccessBindingsResponse{Bindings: bindings, OperationId: operation.GetOperationId()}, nil
}

func (s *Service) CheckAccess(ctx context.Context, req *authzv1.CheckAccessRequest) (*authzv1.AccessCheckResult, error) {
	cacheKey := redisstore.BuildCheckAccessCacheKey(req.GetSubject(), req.GetResource(), req.GetPermission(), s.policyVersion)
	if s.cache != nil {
		if payload, ok, err := s.cache.Get(ctx, cacheKey); err == nil && ok {
			var result authzv1.AccessCheckResult
			if unmarshalErr := json.Unmarshal([]byte(payload), &result); unmarshalErr == nil {
				result.CacheHit = true
				return &result, nil
			}
		}
	}

	result, err := s.checkWithRuntimeOrFallback(ctx, req)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		if payload, marshalErr := json.Marshal(result); marshalErr == nil {
			_ = s.cache.Set(ctx, cacheKey, string(payload), 30*time.Second)
		}
	}
	return result, nil
}

func (s *Service) BatchCheckAccess(ctx context.Context, req *authzv1.BatchCheckAccessRequest) (*authzv1.BatchCheckAccessResponse, error) {
	results := make([]*authzv1.AccessCheckResult, 0, len(req.GetItems()))
	for _, item := range req.GetItems() {
		result, err := s.CheckAccess(ctx, item)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return &authzv1.BatchCheckAccessResponse{Results: results}, nil
}

func (s *Service) ExplainAccess(ctx context.Context, req *authzv1.ExplainAccessRequest) (*authzv1.ExplainAccessResponse, error) {
	result, err := s.checkWithRuntimeOrFallback(ctx, &authzv1.CheckAccessRequest{
		Subject:    req.GetSubject(),
		Resource:   req.GetResource(),
		Permission: req.GetPermission(),
	})
	if err != nil {
		return nil, err
	}
	bindings, err := s.listBindingsForResource(ctx, req.GetResource())
	if err != nil {
		return nil, err
	}
	pathIDs := make([]string, 0, len(bindings))
	summary := "no matching access path found"
	for _, binding := range bindings {
		if sameSubject(binding.GetSubject(), req.GetSubject()) && slices.Contains(PermissionsForRole(binding.GetRoleId()), req.GetPermission()) {
			pathIDs = append(pathIDs, binding.GetBindingId())
			summary = fmt.Sprintf("subject %s has %s via role %s", req.GetSubject().GetId(), req.GetPermission(), binding.GetRoleId())
		}
	}
	return &authzv1.ExplainAccessResponse{
		Decision: result.GetDecision(),
		Summary:  summary,
		PathIds:  pathIDs,
		Revision: fmt.Sprintf("local-%d", s.now().Unix()),
	}, nil
}

func (s *Service) checkWithRuntimeOrFallback(ctx context.Context, req *authzv1.CheckAccessRequest) (*authzv1.AccessCheckResult, error) {
	if s.runtime != nil {
		result, err := s.runtime.Check(ctx, req)
		if err == nil && result != nil {
			return result, nil
		}
		s.logger.Warn("spicedb check failed, using fallback evaluator", zap.Error(err))
	}
	bindings, err := s.listBindingsForResource(ctx, req.GetResource())
	if err != nil {
		return nil, err
	}
	decision := authzv1.PermissionDecision_PERMISSION_DECISION_DENY
	for _, binding := range bindings {
		if sameSubject(binding.GetSubject(), req.GetSubject()) && slices.Contains(PermissionsForRole(binding.GetRoleId()), req.GetPermission()) {
			decision = authzv1.PermissionDecision_PERMISSION_DECISION_ALLOW
			break
		}
	}
	return &authzv1.AccessCheckResult{
		Decision:   decision,
		Permission: req.GetPermission(),
		CacheHit:   false,
		ZedToken:   "fallback",
	}, nil
}

func (s *Service) listBindingsForResource(ctx context.Context, resource *authzv1.ResourceRef) ([]*authzv1.AccessBinding, error) {
	documents, _, err := s.store.ListDocuments(ctx, ydb.TableBindingOperations, resource.GetTenantId(), 0, 1000)
	if err != nil {
		return nil, err
	}
	bindings := make([]*authzv1.AccessBinding, 0, len(documents))
	for _, document := range documents {
		binding := &authzv1.AccessBinding{}
		if err := core.UnmarshalProto(document.Payload, binding); err != nil {
			return nil, err
		}
		if sameResource(binding.GetResource(), resource) {
			bindings = append(bindings, binding)
		}
	}
	return bindings, nil
}

func ListBindingsForSubject(ctx context.Context, store core.DocumentStore, subject *authzv1.SubjectRef) ([]*authzv1.AccessBinding, error) {
	documents, _, err := store.ListDocuments(ctx, ydb.TableBindingOperations, subject.GetTenantId(), 0, 1000)
	if err != nil {
		return nil, err
	}
	bindings := make([]*authzv1.AccessBinding, 0, len(documents))
	for _, document := range documents {
		binding := &authzv1.AccessBinding{}
		if err := core.UnmarshalProto(document.Payload, binding); err != nil {
			return nil, err
		}
		if sameSubject(binding.GetSubject(), subject) {
			bindings = append(bindings, binding)
		}
	}
	return bindings, nil
}

func ListBindingsForResource(ctx context.Context, store core.DocumentStore, resource *authzv1.ResourceRef) ([]*authzv1.AccessBinding, error) {
	service := &Service{store: store}
	return service.listBindingsForResource(ctx, resource)
}

func toAddMutations(bindings []*authzv1.AccessBinding) []*authzv1.AccessBindingMutation {
	mutations := make([]*authzv1.AccessBindingMutation, 0, len(bindings))
	for _, binding := range bindings {
		mutations = append(mutations, &authzv1.AccessBindingMutation{
			Kind:    authzv1.BindingMutationKind_BINDING_MUTATION_KIND_ADD,
			Binding: binding,
		})
	}
	return mutations
}

func sameSubject(left *authzv1.SubjectRef, right *authzv1.SubjectRef) bool {
	return left.GetType() == right.GetType() && left.GetId() == right.GetId()
}

func sameResource(left *authzv1.ResourceRef, right *authzv1.ResourceRef) bool {
	return left.GetType() == right.GetType() && left.GetId() == right.GetId()
}
