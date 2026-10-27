package orgaccess

import (
	"context"
	"errors"
	"testing"

	accesscmd "github.com/m8-team/platform/internal/access/app/command"
	accessquery "github.com/m8-team/platform/internal/access/app/query"
	"github.com/m8-team/platform/internal/resourcemanager/app/orgstructure"
	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/orgunit"
)

type accessFake struct {
	result *accessquery.CheckPermissionResult
	err    error
	calls  []accesscmd.CheckPermissionCommand
}

func (f *accessFake) Handle(_ context.Context, c accesscmd.CheckPermissionCommand) (*accessquery.CheckPermissionResult, error) {
	f.calls = append(f.calls, c)
	return f.result, f.err
}

type approvedImpact struct{}

func (approvedImpact) CheckImpact(context.Context, orgstructure.Actor, orgunit.Structure, orgunit.Structure, orgunit.ChangeSet) error {
	return nil
}
func TestOrgAccessChecksExplicitActorScopeAndRevision(t *testing.T) {
	actor := orgstructure.Actor{TenantID: "tenant", PrincipalID: "person"}
	f := &accessFake{result: &accessquery.CheckPermissionResult{Allowed: true, ModelRevision: "rev-12"}}
	g := &Gate{Checker: f, ModelRevision: "rev-12", Impact: approvedImpact{}}
	if err := g.AuthorizeOrganization(context.Background(), actor, "alpha"); err != nil {
		t.Fatal(err)
	}
	if len(f.calls) != 1 {
		t.Fatalf("calls: %d", len(f.calls))
	}
	c := f.calls[0]
	if c.SubjectID != "person" || c.Permission != "orgunit.list" || c.ResourceID != "tenant/alpha" || c.FailMode != "" || !c.Critical {
		t.Fatalf("bad permissions %#v", c)
	}
	allow, err := g.CanReadUnit(context.Background(), actor, "alpha", "dept")
	if err != nil || !allow {
		t.Fatalf("unit permission: %v %v", allow, err)
	}
	if f.calls[1].ResourceID != "tenant/alpha/dept" {
		t.Fatal("wrong unit scope")
	}
}
func TestOrgAccessDeniedOnFailureDegradationOrRevisionMismatch(t *testing.T) {
	actor := orgstructure.Actor{TenantID: "tenant", PrincipalID: "person"}
	for _, v := range []struct {
		name   string
		result *accessquery.CheckPermissionResult
		err    error
	}{
		{"deny", &accessquery.CheckPermissionResult{Allowed: false, ModelRevision: "rev"}, nil},
		{"degraded", &accessquery.CheckPermissionResult{Allowed: true, Degraded: true, ModelRevision: "rev"}, nil},
		{"stale", &accessquery.CheckPermissionResult{Allowed: true, ModelRevision: "old"}, nil},
		{"unavailable", nil, errors.New("unavailable")},
		{"nil", nil, nil},
	} {
		t.Run(v.name, func(t *testing.T) {
			g := &Gate{Checker: &accessFake{result: v.result, err: v.err}, ModelRevision: "rev"}
			if err := g.AuthorizeOrganization(context.Background(), actor, "org"); !errors.Is(err, ports.ErrPermissionDenied) {
				t.Fatalf("expected deny: %v", err)
			}
		})
	}
	g := &Gate{ModelRevision: "rev"}
	if err := g.AuthorizeOrganization(context.Background(), actor, "org"); !errors.Is(err, ports.ErrPermissionDenied) {
		t.Fatalf("missing engine must deny: %v", err)
	}
}
func TestOrgAccessRequiresIndependentImpactApproval(t *testing.T) {
	g := &Gate{Checker: &accessFake{result: &accessquery.CheckPermissionResult{Allowed: true, ModelRevision: "rev"}}, ModelRevision: "rev"}
	err := g.CheckImpact(context.Background(), orgstructure.Actor{TenantID: "tenant", PrincipalID: "person"}, orgunit.Structure{}, orgunit.Structure{}, orgunit.ChangeSet{})
	if !errors.Is(err, ErrImpactReviewRequired) {
		t.Fatalf("approval bypass: %v", err)
	}
}
