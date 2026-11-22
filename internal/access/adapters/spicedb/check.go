package spicedb

import (
	"context"
	"fmt"

	"github.com/m8-team/platform/internal/access/domain"
)

// CheckPermission reads a fully consistent graph snapshot unless the caller
// provides a trusted minimum ZedToken. Model revision is independently pinned
// in the adapter configuration. A conditional result is denied, never allowed.
func (c *Client) CheckPermission(ctx context.Context, req domain.CheckPermissionRequest) (domain.PermissionDecision, error) {
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
	if definition.Permission() != req.Permission() || definition.ResourceType() != req.Resource().Type() {
		return domain.PermissionDecision{}, ErrInvalidGraphReference
	}
	if !graphName.MatchString(definition.GraphPermission()) {
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
	consistency := map[string]any{"fullyConsistent": true}
	if req.MinZedToken() != "" {
		consistency = map[string]any{"atLeastAsFresh": map[string]string{"token": req.MinZedToken()}}
	}
	request := struct {
		Consistency map[string]any   `json:"consistency"`
		Resource    objectReference  `json:"resource"`
		Permission  string           `json:"permission"`
		Subject     subjectReference `json:"subject"`
	}{Consistency: consistency, Resource: resource, Permission: definition.GraphPermission(), Subject: subject}
	var response struct {
		CheckedAt struct {
			Token string `json:"token"`
		} `json:"checkedAt"`
		Permissionship string `json:"permissionship"`
	}
	if err := c.post(ctx, "/v1/permissions/check", request, &response); err != nil {
		return domain.PermissionDecision{}, err
	}
	var decision domain.Decision
	var reason string
	switch response.Permissionship {
	case "PERMISSIONSHIP_HAS_PERMISSION":
		decision = domain.DecisionAllow
		reason = "spicedb_permission"
	case "PERMISSIONSHIP_NO_PERMISSION":
		decision = domain.DecisionDeny
		reason = "spicedb_no_permission"
	case "PERMISSIONSHIP_CONDITIONAL_PERMISSION":
		decision = domain.DecisionDeny
		reason = "spicedb_missing_caveat_context"
	default:
		return domain.PermissionDecision{}, fmt.Errorf("%w: unknown permissionship", ErrInvalidResponse)
	}
	if response.CheckedAt.Token == "" {
		return domain.PermissionDecision{}, fmt.Errorf("%w: missing checkedAt", ErrInvalidResponse)
	}
	result, err := domain.NewPermissionDecision(decision, c.revision, reason)
	if err != nil {
		return domain.PermissionDecision{}, err
	}
	return result.WithZedToken(response.CheckedAt.Token)
}
