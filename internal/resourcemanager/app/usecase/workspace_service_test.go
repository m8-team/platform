package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/platform/types"
	"github.com/m8-team/platform/internal/resourcemanager/adapter/authz"
	"github.com/m8-team/platform/internal/resourcemanager/adapter/memory"
	"github.com/m8-team/platform/internal/resourcemanager/app/command"
	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/app/query"
	"github.com/m8-team/platform/internal/resourcemanager/app/usecase"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
)

func TestWorkspaceServiceCreateUpdateDeleteAndUndelete(t *testing.T) {
	h := newWorkspaceHarness(t)
	labels := map[string]string{"environment": "prod"}
	created, err := h.service.Create(context.Background(), command.CreateWorkspace{
		OrganizationID: h.organizationID,
		Name:           "Production",
		Labels:         labels,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !created.ID().Equal(workspaceFixtureID(1)) || !created.OrganizationID().Equal(h.organizationID) {
		t.Fatalf("Create() IDs = %s/%s", created.ID(), created.OrganizationID())
	}
	labels["environment"] = "mutated"
	if created.Labels()["environment"] != "prod" {
		t.Fatal("Create() retained caller-owned labels")
	}

	h.clock.now = h.clock.now.Add(time.Minute)
	name := "Primary"
	updated, err := h.service.Update(context.Background(), command.UpdateWorkspace{
		ID:              created.ID(),
		ExpectedVersion: created.Version(),
		Name:            &name,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Name() != name || updated.Version().Int64() != 2 {
		t.Fatalf("Update() = %q/v%s", updated.Name(), updated.Version())
	}
	if _, err := h.service.Update(context.Background(), command.UpdateWorkspace{
		ID: created.ID(), ExpectedVersion: types.InitialVersion, Name: &name,
	}); !errors.Is(err, workspace.ErrVersionMismatch) {
		t.Fatalf("stale Update() error = %v, want %v", err, workspace.ErrVersionMismatch)
	}

	h.clock.now = h.clock.now.Add(time.Minute)
	deleted, err := h.service.Delete(context.Background(), command.DeleteWorkspace{
		ID: created.ID(), ExpectedVersion: updated.Version(),
	})
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if !deleted.IsDeleted() || deleted.DeleteTime() == nil || deleted.PurgeTime() == nil {
		t.Fatal("Delete() did not create a tombstone")
	}

	h.clock.now = h.clock.now.Add(time.Minute)
	restored, err := h.service.Undelete(context.Background(), command.UndeleteWorkspace{ID: created.ID()})
	if err != nil {
		t.Fatalf("Undelete() error = %v", err)
	}
	if restored.IsDeleted() || restored.Version().Int64() != 4 {
		t.Fatalf("Undelete() state/version = %s/%s", restored.State(), restored.Version())
	}
}

func TestWorkspaceServiceListUsesParentBoundSignedKeysetTokens(t *testing.T) {
	h := newWorkspaceHarness(t)
	secondOrganization := createParentOrganization(t, h.organizations, fixtureID(102), h.clock.now)
	for _, item := range []struct {
		parent            organization.ID
		name, environment string
	}{
		{h.organizationID, "Alpha", "prod"},
		{h.organizationID, "Bravo", "prod"},
		{h.organizationID, "Charlie", "dev"},
		{secondOrganization.ID(), "Other", "prod"},
	} {
		if _, err := h.service.Create(context.Background(), command.CreateWorkspace{
			OrganizationID: item.parent,
			Name:           item.name,
			Labels:         map[string]string{"environment": item.environment},
		}); err != nil {
			t.Fatalf("Create(%q) error = %v", item.name, err)
		}
	}

	first, err := h.service.List(context.Background(), query.ListWorkspaces{
		OrganizationID: h.organizationID,
		PageSize:       2,
		OrderBy:        "name asc",
		Filter:         `labels.environment == "prod"`,
	})
	if err != nil {
		t.Fatalf("List(first) error = %v", err)
	}
	assertWorkspaceNames(t, first.Workspaces, "Alpha", "Bravo")
	if first.NextPageToken != "" {
		t.Fatal("filtered two-item result unexpectedly has a next page")
	}

	page, err := h.service.List(context.Background(), query.ListWorkspaces{
		OrganizationID: h.organizationID, PageSize: 1, OrderBy: "name asc",
	})
	if err != nil || page.NextPageToken == "" {
		t.Fatalf("List(page) = token %q, error %v", page.NextPageToken, err)
	}
	second, err := h.service.List(context.Background(), query.ListWorkspaces{
		OrganizationID: h.organizationID, PageSize: 1, PageToken: page.NextPageToken, OrderBy: "name asc",
	})
	if err != nil {
		t.Fatalf("List(second) error = %v", err)
	}
	assertWorkspaceNames(t, second.Workspaces, "Bravo")

	_, err = h.service.List(context.Background(), query.ListWorkspaces{
		OrganizationID: secondOrganization.ID(), PageSize: 1, PageToken: page.NextPageToken, OrderBy: "name asc",
	})
	if !errors.Is(err, usecase.ErrInvalidWorkspacePageToken) {
		t.Fatalf("cross-parent token error = %v, want %v", err, usecase.ErrInvalidWorkspacePageToken)
	}
	_, err = h.service.List(context.Background(), query.ListWorkspaces{
		OrganizationID: h.organizationID, PageSize: 1, PageToken: tamperPageToken(t, page.NextPageToken), OrderBy: "name asc",
	})
	if !errors.Is(err, usecase.ErrInvalidWorkspacePageToken) {
		t.Fatalf("tampered token error = %v, want %v", err, usecase.ErrInvalidWorkspacePageToken)
	}
	_, err = h.service.List(context.Background(), query.ListWorkspaces{
		OrganizationID: h.organizationID, PageSize: 2, PageToken: page.NextPageToken, OrderBy: "name asc",
	})
	if !errors.Is(err, usecase.ErrInvalidWorkspacePageToken) {
		t.Fatalf("changed page-size token error = %v, want %v", err, usecase.ErrInvalidWorkspacePageToken)
	}
}

func TestWorkspaceServiceRejectsInvalidGeneratedID(t *testing.T) {
	h := newWorkspaceHarness(t)
	h.ids.ids = nil
	_, err := h.service.Create(context.Background(), command.CreateWorkspace{OrganizationID: h.organizationID})
	if !errors.Is(err, usecase.ErrGeneratedWorkspaceID) {
		t.Fatalf("Create() error = %v, want %v", err, usecase.ErrGeneratedWorkspaceID)
	}
}

func TestWorkspaceServiceAuthorizesBeforeMissingDeleteLookup(t *testing.T) {
	h := newWorkspaceHarness(t)
	spy := &countingWorkspaceRepository{WorkspaceRepository: h.repository}
	service := newWorkspaceService(t, spy, h.organizations, h.authorizer, h.clock, h.ids)
	h.authorizer.allow = false

	_, err := service.Delete(context.Background(), command.DeleteWorkspace{
		ID: workspaceFixtureID(99), AllowMissing: true,
	})
	if !errors.Is(err, ports.ErrPermissionDenied) {
		t.Fatalf("Delete() error = %v, want %v", err, ports.ErrPermissionDenied)
	}
	if spy.getCalls != 0 {
		t.Fatalf("repository Get calls = %d, want 0", spy.getCalls)
	}
}

func TestWorkspacePreventsDeletingParentAndRestoringUnderDeletedParent(t *testing.T) {
	h := newWorkspaceHarness(t)
	created, err := h.service.Create(context.Background(), command.CreateWorkspace{OrganizationID: h.organizationID})
	if err != nil {
		t.Fatal(err)
	}
	organizationService := newService(
		t,
		h.organizations,
		h.authorizer,
		h.clock,
		&fakeIDGenerator{ids: fixtureIDs(2)},
		h.repository,
		defaultServiceConfig(),
	)
	if _, err := organizationService.Delete(context.Background(), command.DeleteOrganization{ID: h.organizationID}); !errors.Is(err, usecase.ErrOrganizationHasWorkspaces) {
		t.Fatalf("Delete organization error = %v, want %v", err, usecase.ErrOrganizationHasWorkspaces)
	}

	h.clock.now = h.clock.now.Add(time.Minute)
	deleted, err := h.service.Delete(context.Background(), command.DeleteWorkspace{ID: created.ID()})
	if err != nil {
		t.Fatal(err)
	}
	h.clock.now = h.clock.now.Add(time.Minute)
	if _, err := organizationService.Delete(context.Background(), command.DeleteOrganization{ID: h.organizationID}); err != nil {
		t.Fatalf("Delete empty organization error = %v", err)
	}
	h.clock.now = h.clock.now.Add(time.Minute)
	if _, err := h.service.Undelete(context.Background(), command.UndeleteWorkspace{ID: deleted.ID()}); !errors.Is(err, organization.ErrOrganizationNotActive) {
		t.Fatalf("Undelete under deleted parent error = %v, want %v", err, organization.ErrOrganizationNotActive)
	}
}

func TestWorkspaceCreateAndOrganizationDeletePreserveHierarchyInvariant(t *testing.T) {
	for iteration := 0; iteration < 20; iteration++ {
		organizations := memory.NewOrganizationRepository()
		workspaces := memory.NewWorkspaceRepository()
		clock := &fakeClock{now: testStartTime}
		parent := createParentOrganization(t, organizations, fixtureID(200+iteration), clock.now)
		authorizer := authz.AllowAll()
		workspaceService := newWorkspaceService(
			t,
			workspaces,
			organizations,
			authorizer,
			clock,
			&fakeWorkspaceIDGenerator{ids: []workspace.ID{workspaceFixtureID(100 + iteration)}},
		)
		organizationService := newService(
			t,
			organizations,
			authorizer,
			clock,
			&fakeIDGenerator{ids: fixtureIDs(1)},
			workspaces,
			defaultServiceConfig(),
		)

		start := make(chan struct{})
		done := make(chan struct{}, 2)
		go func() {
			<-start
			_, _ = workspaceService.Create(context.Background(), command.CreateWorkspace{OrganizationID: parent.ID()})
			done <- struct{}{}
		}()
		go func() {
			<-start
			_, _ = organizationService.Delete(context.Background(), command.DeleteOrganization{ID: parent.ID()})
			done <- struct{}{}
		}()
		close(start)
		<-done
		<-done

		storedParent, err := organizations.Get(context.Background(), parent.ID())
		if err != nil {
			t.Fatal(err)
		}
		children, err := workspaces.List(context.Background(), ports.ListWorkspacesOptions{
			OrganizationID: parent.ID(),
			Filter:         ports.WorkspaceFilter{ShowDeleted: true},
		})
		if err != nil {
			t.Fatal(err)
		}
		if storedParent.IsDeleted() && len(children.Workspaces) != 0 {
			t.Fatalf("iteration %d: deleted parent has %d workspaces", iteration, len(children.Workspaces))
		}
	}
}

type workspaceHarness struct {
	service        *usecase.WorkspaceService
	repository     *memory.WorkspaceRepository
	organizations  *memory.OrganizationRepository
	authorizer     *fakeAuthorizer
	clock          *fakeClock
	ids            *fakeWorkspaceIDGenerator
	organizationID organization.ID
}

func newWorkspaceHarness(t *testing.T) *workspaceHarness {
	t.Helper()
	organizations := memory.NewOrganizationRepository()
	clock := &fakeClock{now: testStartTime}
	parent := createParentOrganization(t, organizations, fixtureID(101), clock.now)
	repository := memory.NewWorkspaceRepository()
	authorizer := &fakeAuthorizer{allow: true}
	ids := &fakeWorkspaceIDGenerator{ids: workspaceFixtureIDs(16)}
	return &workspaceHarness{
		service:        newWorkspaceService(t, repository, organizations, authorizer, clock, ids),
		repository:     repository,
		organizations:  organizations,
		authorizer:     authorizer,
		clock:          clock,
		ids:            ids,
		organizationID: parent.ID(),
	}
}

func newWorkspaceService(
	t *testing.T,
	repository ports.WorkspaceRepository,
	organizations ports.OrganizationLookup,
	authorizer ports.Authorizer,
	clock ports.Clock,
	ids ports.WorkspaceIDGenerator,
) *usecase.WorkspaceService {
	t.Helper()
	service, err := usecase.NewWorkspaceService(repository, organizations, authorizer, clock, ids, usecase.WorkspaceServiceConfig{
		SoftDeleteRetention: 48 * time.Hour,
		PageTokenKey:        []byte(testPageTokenKey),
	})
	if err != nil {
		t.Fatalf("NewWorkspaceService() error = %v", err)
	}
	return service
}

func createParentOrganization(t *testing.T, repository ports.OrganizationRepository, id organization.ID, now time.Time) *organization.Organization {
	t.Helper()
	value, err := organization.New(organization.CreateParams{ID: id, Name: "Parent", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Create(context.Background(), value); err != nil {
		t.Fatal(err)
	}
	return value
}

type fakeWorkspaceIDGenerator struct{ ids []workspace.ID }

func (g *fakeWorkspaceIDGenerator) NewWorkspaceID() workspace.ID {
	if len(g.ids) == 0 {
		return workspace.ID{}
	}
	id := g.ids[0]
	g.ids = g.ids[1:]
	return id
}

type countingWorkspaceRepository struct {
	*memory.WorkspaceRepository
	getCalls int
}

func (r *countingWorkspaceRepository) Get(ctx context.Context, id workspace.ID) (*workspace.Workspace, error) {
	r.getCalls++
	return r.WorkspaceRepository.Get(ctx, id)
}

func workspaceFixtureIDs(count int) []workspace.ID {
	ids := make([]workspace.ID, count)
	for index := range ids {
		ids[index] = workspaceFixtureID(index + 1)
	}
	return ids
}

func workspaceFixtureID(value int) workspace.ID {
	return workspace.MustParseID(fmt.Sprintf("10000000-0000-4000-8000-%012x", value))
}

func assertWorkspaceNames(t *testing.T, values []*workspace.Workspace, want ...string) {
	t.Helper()
	if len(values) != len(want) {
		t.Fatalf("workspace count = %d, want %d", len(values), len(want))
	}
	for index := range want {
		if values[index].Name() != want[index] {
			t.Fatalf("workspace[%d].name = %q, want %q", index, values[index].Name(), want[index])
		}
	}
}
