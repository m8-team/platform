// Package orgaccess binds Resource Manager OrgUnit authorization to M8 Access.
// Unlike a tenant-wide boolean gate, every operation is checked against an
// explicit subject, scope and authorization-model revision.
package orgaccess

import (
    "context"
    "errors"
    "fmt"

    accesscmd "github.com/m8-team/platform/internal/access/app/command"
    accessquery "github.com/m8-team/platform/internal/access/app/query"
    "github.com/m8-team/platform/internal/resourcemanager/app/orgstructure"
    "github.com/m8-team/platform/internal/resourcemanager/app/ports"
    "github.com/m8-team/platform/internal/resourcemanager/domain/orgunit"
)

var ErrImpactReviewRequired = errors.New("org structure impact assessment is not configured")
var ErrAccessConfigurationRequired = errors.New("M8 Access permission checker and model revision required")

// PermissionChecker is implemented by M8 Access's CheckPermissionHandler.
type PermissionChecker interface {
    Handle(context.Context, accesscmd.CheckPermissionCommand) (*accessquery.CheckPermissionResult, error)
}

// ImpactReviewer is a separate, authoritative owner/employee/access-diff
// evaluator. An Access ALLOW decision alone cannot approve consequences of
// reparenting or closing a department.
type ImpactReviewer interface {
    CheckImpact(context.Context, orgstructure.Actor, orgunit.Structure, orgunit.Structure, orgunit.ChangeSet) error
}

type Gate struct {
    Checker PermissionChecker
    ModelRevision string
    Impact ImpactReviewer
}

func (g *Gate) check(ctx context.Context, actor orgstructure.Actor, permission, resourceType, id string) error {
    if g == nil || g.Checker == nil || g.ModelRevision == "" {
        return ErrAccessConfigurationRequired
    }
    if ctx == nil || actor.TenantID == "" || actor.PrincipalID == "" || id == "" {
        return ports.ErrUnauthenticated
    }
    result, err := g.Checker.Handle(ctx, accesscmd.CheckPermissionCommand{
        SubjectType: "person", SubjectID: actor.PrincipalID,
        Permission: permission, ResourceType: resourceType,
        ResourceID: actor.TenantID + "/" + id,
        ModelRevision: g.ModelRevision,
        Critical: true,
        // FailMode must never be ALLOW for organization mutations or reads.
    })
    if err != nil {
        // Do not leak whether the authorization engine/revision failed.
        return fmt.Errorf("%w: access evaluation unavailable", ports.ErrPermissionDenied)
    }
    if result == nil || !result.Allowed || result.Degraded || result.ModelRevision != g.ModelRevision {
        return ports.ErrPermissionDenied
    }
    return nil
}

func (g *Gate) AuthorizeOrganization(ctx context.Context, actor orgstructure.Actor, organizationID string) error {
    return g.check(ctx, actor, "orgunit.list", "organization", organizationID)
}
func (g *Gate) CanReadUnit(ctx context.Context, actor orgstructure.Actor, organizationID, unitID string) (bool, error) {
    if organizationID == "" || unitID == "" {
        return false, ports.ErrPermissionDenied
    }
    if err := g.check(ctx, actor, "orgunit.read", "orgunit", organizationID + "/" + unitID); err != nil {
        if errors.Is(err, ports.ErrPermissionDenied) {
            return false, nil
        }
        return false, err
    }
    return true, nil
}
func (g *Gate) Authorize(ctx context.Context, actor orgstructure.Actor, plan orgunit.ChangeSet) error {
    return g.check(ctx, actor, "orgunit.changeset.apply", "organization", plan.OrganizationID)
}
func (g *Gate) CheckImpact(ctx context.Context, actor orgstructure.Actor, before, after orgunit.Structure, plan orgunit.ChangeSet) error {
    if g == nil || g.Impact == nil {
        return ErrImpactReviewRequired
    }
    return g.Impact.CheckImpact(ctx, actor, before, after, plan)
}

var _ orgstructure.HierarchyReadGate = (*Gate)(nil)
var _ orgstructure.DecisionGate = (*Gate)(nil)
