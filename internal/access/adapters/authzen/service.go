// Package authzen exposes a permission-evaluating AccessService backed by
// the trusted Resource Manager consistency barrier and SpiceDB Go SDK.
package authzen

import (
	"context"
	"errors"

	"buf.build/go/protovalidate"
	"github.com/m8-team/platform/internal/access/app/consistency"
	"github.com/m8-team/platform/internal/access/domain"
	accesspb "github.com/m8-team/platform/internal/proto/m8/platform/access/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrUnauthenticated      = errors.New("trusted authorization principal not found")
	ErrInvalidConfiguration = errors.New("invalid AuthZEN service configuration")
)

// Principal is produced only by verified OIDC/mTLS authentication. The
// frontend-supplied subject is NOT an authentication credential.
type Principal struct {
	TenantID domain.TenantID
	Subject  domain.Subject
}

// PrincipalResolver MUST authenticate the caller, or verify its service
// identity and scoped delegation; it must never trust client-supplied tenant
// headers, request properties, or arbitrary bearer strings.
type PrincipalResolver interface {
	AuthenticatedPrincipal(context.Context) (Principal, error)
}

// Service is deliberately deny-by-default: a missing authenticator, source
// barrier, or unregistered action cannot silently become an allow.
type Service struct {
	accesspb.UnimplementedAccessServiceServer
	resolver PrincipalResolver
	barrier  *consistency.PolicyBarrier
	model    domain.ModelRevision
}

func New(resolver PrincipalResolver, barrier *consistency.PolicyBarrier, modelRevision string) (*Service, error) {
	if resolver == nil || barrier == nil {
		return nil, ErrInvalidConfiguration
	}
	revision, err := domain.NewModelRevision(modelRevision)
	if err != nil {
		return nil, ErrInvalidConfiguration
	}
	return &Service{resolver: resolver, barrier: barrier, model: revision}, nil
}
func (s *Service) EvaluateAccess(ctx context.Context, req *accesspb.EvaluateAccessRequest) (*accesspb.EvaluateAccessResponse, error) {
	if s == nil || s.resolver == nil || s.barrier == nil {
		return nil, status.Error(codes.Unavailable, "access policy decision point unavailable")
	}
	if req == nil || protovalidate.Validate(req) != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid access evaluation")
	}
	// No implicit trust in AuthZEN attributes, context or dynamic MFA/risk
	// claims until a separately authenticated contextual evaluator exists.
	if req.Context != nil && len(req.Context.Fields) > 0 || req.Subject.Properties != nil && len(req.Subject.Properties.Fields) > 0 ||
		req.Action.Properties != nil && len(req.Action.Properties.Fields) > 0 || req.Resource.Properties != nil && len(req.Resource.Properties.Fields) > 0 {
		return nil, status.Error(codes.InvalidArgument, "untrusted contextual attributes are unsupported")
	}
	principal, err := s.resolver.AuthenticatedPrincipal(ctx)
	if err != nil || principal.TenantID.Validate() != nil || principal.Subject.Validate() != nil {
		return nil, status.Error(codes.Unauthenticated, "trusted principal required")
	}
	if req.Subject.Type != principal.Subject.Type() || req.Subject.Id != principal.Subject.ID() {
		return nil, status.Error(codes.PermissionDenied, "delegation is not authorized")
	}
	subject, err := domain.NewSubject(req.Subject.Type, req.Subject.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid subject")
	}
	permission, err := domain.NewPermission(req.Action.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid action")
	}
	resource, err := domain.NewResource(req.Resource.Type, req.Resource.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid resource")
	}
	check, err := domain.NewCheckPermissionRequest(domain.CheckPermissionInput{
		TenantID: principal.TenantID, Subject: subject, Permission: permission,
		Resource: resource, ModelRevision: s.model, FailMode: domain.FailModeDeny, Critical: true,
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid authorization request")
	}
	decision, err := s.barrier.CheckPermission(ctx, check)
	if err != nil {
		// Source freshness, unknown ownership, missing checkpoint or SpiceDB
		// failure is never converted into a permit. Sanitize dependency errors.
		return nil, status.Error(codes.Unavailable, "authorization consistency unavailable")
	}
	return &accesspb.EvaluateAccessResponse{Decision: decision.Allowed()}, nil
}

func (s *Service) EvaluateAccessBatch(ctx context.Context, req *accesspb.EvaluateAccessBatchRequest) (*accesspb.EvaluateAccessBatchResponse, error) {
	if req == nil || protovalidate.Validate(req) != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid batch")
	}
	if len(req.Evaluations) > 1000 {
		return nil, status.Error(codes.InvalidArgument, "batch too large")
	}
	if len(req.Evaluations) == 0 {
		result, err := s.EvaluateAccess(ctx, &accesspb.EvaluateAccessRequest{
			Subject: req.Subject, Action: req.Action, Resource: req.Resource, Context: req.Context,
		})
		if err != nil {
			return nil, err
		}
		v := result.Decision
		return &accesspb.EvaluateAccessBatchResponse{Decision: &v, Context: result.Context}, nil
	}
	semantic := "execute_all"
	if req.Options != nil && req.Options.EvaluationsSemantic != nil {
		semantic = req.Options.GetEvaluationsSemantic()
	}
	if semantic != "execute_all" && semantic != "deny_on_first_deny" && semantic != "permit_on_first_permit" {
		return nil, status.Error(codes.InvalidArgument, "invalid batch semantics")
	}
	output := &accesspb.EvaluateAccessBatchResponse{Evaluations: make([]*accesspb.EvaluateAccessResponse, 0, len(req.Evaluations))}
	for _, item := range req.Evaluations {
		if item == nil {
			return nil, status.Error(codes.InvalidArgument, "nil evaluation")
		}
		subject, action, resource, contextData := item.Subject, item.Action, item.Resource, item.Context
		if subject == nil {
			subject = req.Subject
		}
		if action == nil {
			action = req.Action
		}
		if resource == nil {
			resource = req.Resource
		}
		if contextData == nil {
			contextData = req.Context
		}
		result, err := s.EvaluateAccess(ctx, &accesspb.EvaluateAccessRequest{Subject: subject, Action: action, Resource: resource, Context: contextData})
		if err != nil {
			return nil, err
		}
		output.Evaluations = append(output.Evaluations, result)
		if semantic == "deny_on_first_deny" && !result.Decision || semantic == "permit_on_first_permit" && result.Decision {
			break
		}
	}
	return output, nil
}
