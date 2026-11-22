package spicedb

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func classifyRPCFailure(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(ctx.Err(), context.Canceled) || errors.Is(err, context.Canceled) {
		return context.Canceled
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) || status.Code(err) == codes.DeadlineExceeded {
		return fmt.Errorf("%w: gRPC deadline", ports.ErrPermissionEngineTimeout)
	}
	switch status.Code(err) {
	case codes.Unavailable, codes.ResourceExhausted:
		return fmt.Errorf("%w: SpiceDB unavailable", ports.ErrPermissionEngineUnavailable)
	default:
		return fmt.Errorf("%w: %s", ErrRPCRejected, status.Code(err))
	}
}

// CheckPermission resolves an M8 action via the versioned registry and invokes
// the official SDK's gRPC API. Conditional/unknown results NEVER allow access.
func (c *Client) CheckPermission(ctx context.Context, req domain.CheckPermissionRequest) (domain.PermissionDecision, error) {
	if err := ctx.Err(); err != nil {
		return domain.PermissionDecision{}, err
	}
	if err := req.Validate(); err != nil {
		return domain.PermissionDecision{}, err
	}
	if !req.ModelRevision().Equal(c.revision) {
		return domain.PermissionDecision{}, ErrSchemaRevisionMismatch
	}
	definition, err := c.catalog.ResolvePermission(ctx, req.TenantID(), req.Permission(), req.Resource().Type())
	if err != nil {
		return domain.PermissionDecision{}, err
	}
	if err := definition.Validate(); err != nil {
		return domain.PermissionDecision{}, err
	}
	if definition.Permission() != req.Permission() || definition.ResourceType() != req.Resource().Type() || !graphName.MatchString(definition.GraphPermission()) {
		return domain.PermissionDecision{}, ErrInvalidGraphReference
	}
	resource, err := graphObject(req.TenantID(), req.Resource().Type(), req.Resource().ID())
	if err != nil {
		return domain.PermissionDecision{}, err
	}
	subject, err := graphSubject(req.TenantID(), req.Subject(), "")
	if err != nil {
		return domain.PermissionDecision{}, err
	}
	consistency := &pb.Consistency{Requirement: &pb.Consistency_FullyConsistent{FullyConsistent: true}}
	if req.MinZedToken() != "" {
		consistency = &pb.Consistency{Requirement: &pb.Consistency_AtLeastAsFresh{AtLeastAsFresh: &pb.ZedToken{Token: req.MinZedToken()}}}
	}
	resp, err := c.permissions.CheckPermission(ctx, &pb.CheckPermissionRequest{
		Consistency: consistency, Resource: resource, Subject: subject, Permission: definition.GraphPermission(),
	})
	if err != nil {
		return domain.PermissionDecision{}, classifyRPCFailure(ctx, err)
	}
	if resp == nil || resp.CheckedAt == nil || resp.CheckedAt.Token == "" {
		return domain.PermissionDecision{}, ErrInvalidResponse
	}
	var decision domain.Decision
	var reason string
	switch resp.Permissionship {
	case pb.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION:
		decision, reason = domain.DecisionAllow, "spicedb_permission"
	case pb.CheckPermissionResponse_PERMISSIONSHIP_NO_PERMISSION:
		decision, reason = domain.DecisionDeny, "spicedb_no_permission"
	case pb.CheckPermissionResponse_PERMISSIONSHIP_CONDITIONAL_PERMISSION:
		decision, reason = domain.DecisionDeny, "spicedb_missing_caveat_context"
	default:
		return domain.PermissionDecision{}, fmt.Errorf("%w: unknown permissionship", ErrInvalidResponse)
	}
	result, err := domain.NewPermissionDecision(decision, c.revision, reason)
	if err != nil {
		return domain.PermissionDecision{}, err
	}
	return result.WithZedToken(resp.CheckedAt.Token)
}
