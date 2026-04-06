package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
	eventsv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/events/v1"
	redisstore "github.com/m8platform/platform/iam/internal/adapter/out/redis"
	legacyspicedb "github.com/m8platform/platform/iam/internal/adapter/out/spicedb"
	ydb "github.com/m8platform/platform/iam/internal/adapter/out/ydb"
	foundationconfig "github.com/m8platform/platform/iam/internal/foundation/config"
	foundationcontracts "github.com/m8platform/platform/iam/internal/foundation/contracts"
	foundationprotokit "github.com/m8platform/platform/iam/internal/foundation/protokit"
	foundationstore "github.com/m8platform/platform/iam/internal/foundation/store"
	modulaudit "github.com/m8platform/platform/iam/internal/module/audit"
	authzentity "github.com/m8platform/platform/iam/internal/module/authz/entity"
	authzmodel "github.com/m8platform/platform/iam/internal/module/authz/model"
	authzport "github.com/m8platform/platform/iam/internal/module/authz/port"
	authzuc "github.com/m8platform/platform/iam/internal/module/authz/usecase"
	"github.com/m8platform/platform/iam/internal/shared/principal"
	"github.com/m8platform/platform/iam/internal/shared/resource"
	"go.uber.org/zap"
)

type Server struct {
	authzv1.UnimplementedAuthorizationFacadeServiceServer

	store                foundationstore.DocumentStore
	cache                foundationcontracts.Cache
	publisher            foundationcontracts.EventPublisher
	runtime              foundationcontracts.AuthorizationRuntime
	logger               *zap.Logger
	now                  func() time.Time
	policyVersion        string
	topics               foundationconfig.TopicsConfig
	checkFallbackLogOnce sync.Once
	writeFallbackLogOnce sync.Once

	checkAccess *authzuc.CheckAccessUseCase
	bindings    authzport.AccessBindingRepository
	roles       authzport.RolePermissionResolver
}

func NewServer(
	store foundationstore.DocumentStore,
	cache foundationcontracts.Cache,
	publisher foundationcontracts.EventPublisher,
	runtime foundationcontracts.AuthorizationRuntime,
	logger *zap.Logger,
	policyVersion string,
	topics foundationconfig.TopicsConfig,
	checkAccess *authzuc.CheckAccessUseCase,
	bindings authzport.AccessBindingRepository,
	roles authzport.RolePermissionResolver,
) *Server {
	return &Server{
		store:         store,
		cache:         cache,
		publisher:     publisher,
		runtime:       runtime,
		logger:        logger,
		now:           time.Now,
		policyVersion: policyVersion,
		topics:        topics,
		checkAccess:   checkAccess,
		bindings:      bindings,
		roles:         roles,
	}
}

func (s *Server) GetRole(_ context.Context, req *authzv1.GetRoleRequest) (*authzv1.Role, error) {
	role, ok := authzentity.ResolveRole(req.GetRoleId())
	if !ok {
		return nil, fmt.Errorf("role %s not found", req.GetRoleId())
	}
	return roleToProto(role), nil
}

func (s *Server) ListRoles(context.Context, *authzv1.ListRolesRequest) (*authzv1.ListRolesResponse, error) {
	roles := authzentity.DefaultRoles()
	items := make([]*authzv1.Role, 0, len(roles))
	for _, role := range roles {
		items = append(items, roleToProto(role))
	}
	return &authzv1.ListRolesResponse{Roles: items}, nil
}

func (s *Server) SetAccessBindings(ctx context.Context, req *authzv1.SetAccessBindingsRequest) (*authzv1.SetAccessBindingsResponse, error) {
	now := s.now()
	current, err := s.listBindingsForResource(ctx, req.GetResource())
	if err != nil {
		return nil, err
	}
	if err := s.applyBindingsSnapshot(ctx, current, req.GetDesiredBindings(), now); err != nil {
		return nil, err
	}
	if err := s.syncResourceWrite(ctx, req.GetResource(), req.GetDesiredBindings(), current, now); err != nil {
		return nil, err
	}
	operation := modulaudit.NewOperation(now, req.GetResource().GetTenantId(), "set_access_bindings", req.GetResource().GetType().String(), req.GetResource().GetId())
	if err := modulaudit.PersistOperation(ctx, s.store, operation, now); err != nil {
		return nil, err
	}
	audit := modulaudit.NewEvent(now, req.GetResource().GetTenantId(), "access_bindings.set", req.GetPerformedBy(), operation.GetOperationId(), req.GetReason())
	audit.Resource = req.GetResource()
	if err := modulaudit.PersistEvent(ctx, s.store, audit, now); err != nil {
		return nil, err
	}
	event := &eventsv1.AccessIntentChanged{
		EventId:               operation.GetOperationId(),
		OccurredAt:            foundationprotokit.Timestamp(now),
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

func (s *Server) UpdateAccessBindings(ctx context.Context, req *authzv1.UpdateAccessBindingsRequest) (*authzv1.UpdateAccessBindingsResponse, error) {
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
		case authzv1.BindingMutationKind_BINDING_MUTATION_KIND_REMOVE:
			delete(index, mutation.GetBinding().GetBindingId())
		}
	}
	bindings := make([]*authzv1.AccessBinding, 0, len(index))
	for _, binding := range index {
		bindings = append(bindings, binding)
	}
	if err := s.applyBindingsSnapshot(ctx, current, bindings, now); err != nil {
		return nil, err
	}
	operation := modulaudit.NewOperation(now, req.GetResource().GetTenantId(), "update_access_bindings", req.GetResource().GetType().String(), req.GetResource().GetId())
	if err := modulaudit.PersistOperation(ctx, s.store, operation, now); err != nil {
		return nil, err
	}
	audit := modulaudit.NewEvent(now, req.GetResource().GetTenantId(), "access_bindings.updated", req.GetPerformedBy(), operation.GetOperationId(), req.GetReason())
	audit.Resource = req.GetResource()
	if err := modulaudit.PersistEvent(ctx, s.store, audit, now); err != nil {
		return nil, err
	}
	event := &eventsv1.AccessIntentChanged{
		EventId:               operation.GetOperationId(),
		OccurredAt:            foundationprotokit.Timestamp(now),
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
	if err := s.syncResourceWrite(ctx, req.GetResource(), bindings, current, now); err != nil {
		return nil, err
	}
	return &authzv1.UpdateAccessBindingsResponse{Bindings: bindings, OperationId: operation.GetOperationId()}, nil
}

func (s *Server) CheckAccess(ctx context.Context, req *authzv1.CheckAccessRequest) (*authzv1.AccessCheckResult, error) {
	if s.checkAccess != nil && req.GetSubject() != nil && req.GetResource() != nil {
		result, err := s.checkAccess.Execute(ctx, accessCheckQueryFromProto(req))
		if err == nil {
			return &authzv1.AccessCheckResult{
				Decision:          permissionDecisionToProto(result.Decision),
				Permission:        result.Permission,
				CacheHit:          result.CacheHit,
				ZedToken:          result.ZedToken,
				CaveatExpressions: result.CaveatExpressions,
			}, nil
		}
	}

	cacheKey := redisstore.BuildCheckAccessCacheKey(req.GetSubject(), req.GetResource(), req.GetPermission(), s.policyVersion)
	useCache := s.cache != nil && s.runtime == nil
	if useCache {
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
	if useCache {
		if payload, marshalErr := json.Marshal(result); marshalErr == nil {
			_ = s.cache.Set(ctx, cacheKey, string(payload), 30*time.Second)
		}
	}
	return result, nil
}

func (s *Server) BatchCheckAccess(ctx context.Context, req *authzv1.BatchCheckAccessRequest) (*authzv1.BatchCheckAccessResponse, error) {
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

func (s *Server) ExplainAccess(ctx context.Context, req *authzv1.ExplainAccessRequest) (*authzv1.ExplainAccessResponse, error) {
	if s.checkAccess != nil && s.bindings != nil && s.roles != nil && req.GetSubject() != nil && req.GetResource() != nil {
		result, err := s.checkAccess.Execute(ctx, authzmodel.AccessCheckQuery{
			Subject: principal.Principal{
				TenantID: req.GetSubject().GetTenantId(),
				Type:     req.GetSubject().GetType().String(),
				ID:       req.GetSubject().GetId(),
			},
			Resource: resource.Ref{
				TenantID: req.GetResource().GetTenantId(),
				Type:     req.GetResource().GetType().String(),
				ID:       req.GetResource().GetId(),
			},
			Permission: req.GetPermission(),
		})
		if err == nil {
			bindings, repoErr := s.bindings.ListByResource(ctx, resource.Ref{
				TenantID: req.GetResource().GetTenantId(),
				Type:     req.GetResource().GetType().String(),
				ID:       req.GetResource().GetId(),
			})
			if repoErr != nil {
				return nil, repoErr
			}
			pathIDs := make([]string, 0, len(bindings))
			summary := "no matching access path found"
			subject := principal.Principal{
				TenantID: req.GetSubject().GetTenantId(),
				Type:     req.GetSubject().GetType().String(),
				ID:       req.GetSubject().GetId(),
			}
			for _, binding := range bindings {
				if binding.Subject.Equals(subject) && slices.Contains(s.roles.Permissions(binding.RoleID), req.GetPermission()) {
					pathIDs = append(pathIDs, binding.ID)
					summary = fmt.Sprintf("subject %s has %s via role %s", subject.ID, req.GetPermission(), binding.RoleID)
				}
			}
			return &authzv1.ExplainAccessResponse{
				Decision: permissionDecisionToProto(result.Decision),
				Summary:  summary,
				PathIds:  pathIDs,
				Revision: fmt.Sprintf("local-%d", time.Now().Unix()),
			}, nil
		}
	}

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
		if sameSubject(binding.GetSubject(), req.GetSubject()) && slices.Contains(authzentity.PermissionsForRole(binding.GetRoleId()), req.GetPermission()) {
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

func ListBindingsForSubject(ctx context.Context, store foundationstore.DocumentStore, subject *authzv1.SubjectRef) ([]*authzv1.AccessBinding, error) {
	documents, _, err := store.ListDocuments(ctx, ydb.TableBindingOperations, subject.GetTenantId(), 0, 1000)
	if err != nil {
		return nil, err
	}
	bindings := make([]*authzv1.AccessBinding, 0, len(documents))
	for _, document := range documents {
		binding := &authzv1.AccessBinding{}
		if err := foundationprotokit.Unmarshal(document.Payload, binding); err != nil {
			return nil, err
		}
		if sameSubject(binding.GetSubject(), subject) {
			bindings = append(bindings, binding)
		}
	}
	return bindings, nil
}

func ListBindingsForResource(ctx context.Context, store foundationstore.DocumentStore, resource *authzv1.ResourceRef) ([]*authzv1.AccessBinding, error) {
	service := &Server{store: store}
	return service.listBindingsForResource(ctx, resource)
}

func (s *Server) listBindingsForResource(ctx context.Context, resource *authzv1.ResourceRef) ([]*authzv1.AccessBinding, error) {
	documents, _, err := s.store.ListDocuments(ctx, ydb.TableBindingOperations, resource.GetTenantId(), 0, 1000)
	if err != nil {
		return nil, err
	}
	bindings := make([]*authzv1.AccessBinding, 0, len(documents))
	for _, document := range documents {
		binding := &authzv1.AccessBinding{}
		if err := foundationprotokit.Unmarshal(document.Payload, binding); err != nil {
			return nil, err
		}
		if sameResource(binding.GetResource(), resource) {
			bindings = append(bindings, binding)
		}
	}
	return bindings, nil
}

func (s *Server) applyBindingsSnapshot(ctx context.Context, current []*authzv1.AccessBinding, desired []*authzv1.AccessBinding, now time.Time) error {
	desiredIDs := make(map[string]struct{}, len(desired))
	for _, binding := range desired {
		if binding == nil {
			continue
		}
		desiredIDs[binding.GetBindingId()] = struct{}{}
		if err := foundationstore.SaveProto(ctx, s.store, ydb.TableBindingOperations, binding.GetBindingId(), binding.GetResource().GetTenantId(), binding, now); err != nil {
			return err
		}
	}
	for _, binding := range current {
		if binding == nil {
			continue
		}
		if _, ok := desiredIDs[binding.GetBindingId()]; ok {
			continue
		}
		if err := s.store.DeleteDocument(ctx, ydb.TableBindingOperations, binding.GetBindingId()); err != nil && err != foundationstore.ErrNotFound {
			return err
		}
	}
	return nil
}

func (s *Server) syncResourceWrite(ctx context.Context, resource *authzv1.ResourceRef, desired []*authzv1.AccessBinding, previous []*authzv1.AccessBinding, now time.Time) error {
	if s.runtime == nil {
		return nil
	}
	if err := s.runtime.SyncResource(ctx, resource, desired); err != nil {
		if isExpectedRuntimeFallback(err) {
			s.logRuntimeWriteFallback(err)
			return nil
		}
		if rollbackErr := s.applyBindingsSnapshot(ctx, desired, previous, now); rollbackErr != nil {
			if s.logger != nil {
				s.logger.Warn("spicedb write rollback failed", zap.Error(rollbackErr))
			}
			return fmt.Errorf("spicedb sync failed: %w (rollback failed: %v)", err, rollbackErr)
		}
		s.logRuntimeWriteFallback(err)
		return err
	}
	return nil
}

func (s *Server) checkWithRuntimeOrFallback(ctx context.Context, req *authzv1.CheckAccessRequest) (*authzv1.AccessCheckResult, error) {
	if s.runtime != nil {
		result, err := s.runtime.Check(ctx, req)
		if err == nil && result != nil {
			return result, nil
		}
		s.logRuntimeCheckFallback(err)
	}
	bindings, err := s.listBindingsForResource(ctx, req.GetResource())
	if err != nil {
		return nil, err
	}
	decision := authzv1.PermissionDecision_PERMISSION_DECISION_DENY
	for _, binding := range bindings {
		if sameSubject(binding.GetSubject(), req.GetSubject()) && slices.Contains(authzentity.PermissionsForRole(binding.GetRoleId()), req.GetPermission()) {
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

func accessCheckQueryFromProto(req *authzv1.CheckAccessRequest) authzmodel.AccessCheckQuery {
	query := authzmodel.AccessCheckQuery{
		Subject: principal.Principal{
			TenantID: req.GetSubject().GetTenantId(),
			Type:     req.GetSubject().GetType().String(),
			ID:       req.GetSubject().GetId(),
		},
		Resource: resource.Ref{
			TenantID: req.GetResource().GetTenantId(),
			Type:     req.GetResource().GetType().String(),
			ID:       req.GetResource().GetId(),
		},
		Permission: req.GetPermission(),
	}
	if req.GetCaveatContext() != nil {
		query.CaveatContext = req.GetCaveatContext().AsMap()
	}
	return query
}

func roleToProto(role authzentity.Role) *authzv1.Role {
	permissions := make([]*authzv1.Permission, 0, len(role.Permissions))
	for _, permission := range role.Permissions {
		permissions = append(permissions, &authzv1.Permission{
			Id:          permission.ID,
			DisplayName: permission.DisplayName,
		})
	}
	return &authzv1.Role{
		RoleId:        role.ID,
		DisplayName:   role.DisplayName,
		Description:   role.Description,
		Permissions:   permissions,
		RelationNames: append([]string(nil), role.RelationNames...),
	}
}

func permissionDecisionToProto(value authzentity.PermissionDecision) authzv1.PermissionDecision {
	switch value {
	case authzentity.PermissionDecisionAllow:
		return authzv1.PermissionDecision_PERMISSION_DECISION_ALLOW
	case authzentity.PermissionDecisionConditional:
		return authzv1.PermissionDecision_PERMISSION_DECISION_CONDITIONAL
	default:
		return authzv1.PermissionDecision_PERMISSION_DECISION_DENY
	}
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

func (s *Server) logRuntimeCheckFallback(err error) {
	if s == nil || err == nil {
		return
	}
	if isExpectedRuntimeFallback(err) {
		s.checkFallbackLogOnce.Do(func() {
			if s.logger != nil {
				s.logger.Info("spicedb runtime unavailable; using fallback evaluator", zap.Error(err))
			}
		})
		return
	}
	if s.logger != nil {
		s.logger.Warn("spicedb check failed, using fallback evaluator", zap.Error(err))
	}
}

func (s *Server) logRuntimeWriteFallback(err error) {
	if s == nil || err == nil {
		return
	}
	if isExpectedRuntimeFallback(err) {
		s.writeFallbackLogOnce.Do(func() {
			if s.logger != nil {
				s.logger.Info("spicedb runtime unavailable; write sync skipped", zap.Error(err))
			}
		})
		return
	}
	if s.logger != nil {
		s.logger.Warn("spicedb write skipped", zap.Error(err))
	}
}

func isExpectedRuntimeFallback(err error) bool {
	return errors.Is(err, legacyspicedb.ErrNotConfigured) || errors.Is(err, legacyspicedb.ErrNotImplemented)
}
