package usecase_test

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/platform/types"
	"github.com/m8-team/platform/internal/resourcemanager/adapter/memory"
	"github.com/m8-team/platform/internal/resourcemanager/app/command"
	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/app/query"
	"github.com/m8-team/platform/internal/resourcemanager/app/usecase"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
)

const testPageTokenKey = "0123456789abcdef0123456789abcdef"

var testStartTime = time.Date(2026, time.July, 19, 10, 30, 0, 0, time.FixedZone("CEST", 2*60*60))

func TestOrganizationServiceCreateAndGet(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	labels := map[string]string{"environment": "production"}

	created, err := h.service.Create(context.Background(), command.CreateOrganization{
		Name:        "Acme",
		Description: "Primary organization",
		Labels:      labels,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if got, want := created.ID(), fixtureID(1); !got.Equal(want) {
		t.Errorf("created ID = %s, want %s", got, want)
	}
	if got, want := created.State(), organization.StateActive; got != want {
		t.Errorf("created state = %s, want %s", got, want)
	}
	if got, want := created.Version(), types.InitialVersion; !got.Equal(want) {
		t.Errorf("created version = %s, want %s", got, want)
	}
	if got, want := created.CreateTime(), testStartTime.UTC(); !got.Equal(want) {
		t.Errorf("created create time = %s, want %s", got, want)
	}
	if created.CreateTime().Location() != time.UTC || created.UpdateTime().Location() != time.UTC {
		t.Errorf("created timestamps must be normalized to UTC")
	}

	// Neither the request nor the returned aggregate may alias repository state.
	labels["environment"] = "development"
	returnedLabels := created.Labels()
	returnedLabels["environment"] = "staging"

	got, err := h.service.Get(context.Background(), query.GetOrganization{ID: created.ID()})
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Name() != "Acme" || got.Description() != "Primary organization" {
		t.Errorf("Get() = (%q, %q), want original fields", got.Name(), got.Description())
	}
	if got.Labels()["environment"] != "production" {
		t.Errorf("stored labels = %v, want detached production label", got.Labels())
	}

	wantRequests := []ports.AuthorizationRequest{
		{Action: ports.ActionCreateOrganization},
		{Action: ports.ActionGetOrganization, OrganizationID: created.ID()},
	}
	assertAuthorizationRequests(t, h.authorizer.requests, wantRequests)
}

func TestOrganizationServiceListUsesSignedKeysetTokens(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	for _, item := range []struct {
		name   string
		labels map[string]string
	}{
		{name: "Alpha", labels: map[string]string{"environment": "prod"}},
		{name: "Bravo", labels: map[string]string{"environment": "prod"}},
		{name: "Charlie", labels: map[string]string{"environment": "dev"}},
	} {
		if _, err := h.service.Create(context.Background(), command.CreateOrganization{
			Name:   item.name,
			Labels: item.labels,
		}); err != nil {
			t.Fatalf("Create(%q) error = %v", item.name, err)
		}
	}

	first, err := h.service.List(context.Background(), query.ListOrganizations{
		PageSize: 2,
		OrderBy:  "name asc",
	})
	if err != nil {
		t.Fatalf("List(first page) error = %v", err)
	}
	assertOrganizationNames(t, first.Organizations, "Alpha", "Bravo")
	if first.NextPageToken == "" {
		t.Fatal("List(first page) next page token is empty")
	}
	if first.TotalSize != 3 {
		t.Errorf("List(first page) total size = %d, want 3", first.TotalSize)
	}

	// Inserting before the cursor must not shift or duplicate the second page.
	if _, err := h.service.Create(context.Background(), command.CreateOrganization{Name: "Aardvark"}); err != nil {
		t.Fatalf("Create(Aardvark) error = %v", err)
	}
	second, err := h.service.List(context.Background(), query.ListOrganizations{
		PageSize:  2,
		PageToken: first.NextPageToken,
		OrderBy:   "name asc",
	})
	if err != nil {
		t.Fatalf("List(second page) error = %v", err)
	}
	assertOrganizationNames(t, second.Organizations, "Charlie")
	if second.NextPageToken != "" {
		t.Errorf("List(second page) next page token = %q, want empty", second.NextPageToken)
	}
	if second.TotalSize != 4 {
		t.Errorf("List(second page) total size = %d, want 4", second.TotalSize)
	}

	t.Run("tampered signature", func(t *testing.T) {
		_, err := h.service.List(context.Background(), query.ListOrganizations{
			PageSize:  2,
			PageToken: tamperPageToken(t, first.NextPageToken),
			OrderBy:   "name asc",
		})
		if !errors.Is(err, usecase.ErrInvalidOrganizationPageToken) {
			t.Fatalf("List() error = %v, want %v", err, usecase.ErrInvalidOrganizationPageToken)
		}
	})

	t.Run("different signing key", func(t *testing.T) {
		other := newService(t, h.repository, h.authorizer, h.clock, h.ids, h.children, usecase.OrganizationServiceConfig{
			SoftDeleteRetention: 48 * time.Hour,
			PageTokenKey:        []byte("abcdef0123456789abcdef0123456789"),
		})
		_, err := other.List(context.Background(), query.ListOrganizations{
			PageSize:  2,
			PageToken: first.NextPageToken,
			OrderBy:   "name asc",
		})
		if !errors.Is(err, usecase.ErrInvalidOrganizationPageToken) {
			t.Fatalf("List() error = %v, want %v", err, usecase.ErrInvalidOrganizationPageToken)
		}
	})

	t.Run("different authorization scope", func(t *testing.T) {
		otherAuthorizer := &fakeAuthorizer{allow: true, scopeKey: "test:other-caller"}
		other := newService(t, h.repository, otherAuthorizer, h.clock, h.ids, h.children, defaultServiceConfig())
		_, err := other.List(context.Background(), query.ListOrganizations{
			PageSize:  2,
			PageToken: first.NextPageToken,
			OrderBy:   "name asc",
		})
		if !errors.Is(err, usecase.ErrInvalidOrganizationPageToken) {
			t.Fatalf("List() error = %v, want %v", err, usecase.ErrInvalidOrganizationPageToken)
		}
	})

	for name, changedRequest := range map[string]query.ListOrganizations{
		"page size": {
			PageSize: 3, PageToken: first.NextPageToken, OrderBy: "name asc",
		},
		"order": {
			PageSize: 2, PageToken: first.NextPageToken, OrderBy: "name desc",
		},
		"filter": {
			PageSize: 2, PageToken: first.NextPageToken, OrderBy: "name asc", Filter: `labels.environment = "prod"`,
		},
		"show deleted": {
			PageSize: 2, PageToken: first.NextPageToken, OrderBy: "name asc", ShowDeleted: true,
		},
	} {
		t.Run("request mismatch "+name, func(t *testing.T) {
			_, err := h.service.List(context.Background(), changedRequest)
			if !errors.Is(err, usecase.ErrInvalidOrganizationPageToken) {
				t.Fatalf("List() error = %v, want %v", err, usecase.ErrInvalidOrganizationPageToken)
			}
		})
	}
}

func TestOrganizationServiceListFiltersAndDeletedVisibility(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	prod, err := h.service.Create(context.Background(), command.CreateOrganization{
		Name:   "Production",
		Labels: map[string]string{"environment": "prod", "team": "platform"},
	})
	if err != nil {
		t.Fatalf("Create(Production) error = %v", err)
	}
	if _, err := h.service.Create(context.Background(), command.CreateOrganization{
		Name:   "Development",
		Labels: map[string]string{"environment": "dev", "team": "platform"},
	}); err != nil {
		t.Fatalf("Create(Development) error = %v", err)
	}

	h.clock.now = testStartTime.Add(time.Hour)
	if _, err := h.service.Delete(context.Background(), command.DeleteOrganization{ID: prod.ID()}); err != nil {
		t.Fatalf("Delete(Production) error = %v", err)
	}

	visible, err := h.service.List(context.Background(), query.ListOrganizations{})
	if err != nil {
		t.Fatalf("List(default) error = %v", err)
	}
	assertOrganizationNames(t, visible.Organizations, "Development")

	deleted, err := h.service.List(context.Background(), query.ListOrganizations{
		Filter:      `state = "DELETED" AND labels.environment = "prod" AND labels.team = "platform"`,
		ShowDeleted: true,
	})
	if err != nil {
		t.Fatalf("List(deleted filter) error = %v", err)
	}
	assertOrganizationNames(t, deleted.Organizations, "Production")

	_, err = h.service.List(context.Background(), query.ListOrganizations{Filter: `description = "unsupported"`})
	if !errors.Is(err, usecase.ErrInvalidOrganizationFilter) {
		t.Fatalf("List(unsupported filter) error = %v, want %v", err, usecase.ErrInvalidOrganizationFilter)
	}
	_, err = h.service.List(context.Background(), query.ListOrganizations{OrderBy: "state asc"})
	if !errors.Is(err, usecase.ErrInvalidOrganizationOrderBy) {
		t.Fatalf("List(unsupported order) error = %v, want %v", err, usecase.ErrInvalidOrganizationOrderBy)
	}
}

func TestOrganizationServiceUpdateUsesPointersAndVersionPreconditions(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	created, err := h.service.Create(context.Background(), command.CreateOrganization{
		Name:        "Original",
		Description: "Old description",
		Labels:      map[string]string{"old": "value"},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	h.clock.now = testStartTime.Add(time.Minute)
	description := "New description"
	labels := map[string]string{}
	updated, err := h.service.Update(context.Background(), command.UpdateOrganization{
		ID:              created.ID(),
		ExpectedVersion: types.InitialVersion,
		Description:     &description,
		Labels:          &labels,
	})
	if err != nil {
		t.Fatalf("Update(selected fields) error = %v", err)
	}
	if updated.Name() != "Original" {
		t.Errorf("Update() name = %q, want preserved Original", updated.Name())
	}
	if updated.Description() != description {
		t.Errorf("Update() description = %q, want %q", updated.Description(), description)
	}
	if len(updated.Labels()) != 0 {
		t.Errorf("Update() labels = %v, want cleared map", updated.Labels())
	}
	if got, want := updated.Version(), types.Version(2); !got.Equal(want) {
		t.Errorf("Update() version = %s, want %s", got, want)
	}

	// A pointer to an empty value clears the selected field; nil preserves it.
	h.clock.now = testStartTime.Add(2 * time.Minute)
	emptyName := ""
	updated, err = h.service.Update(context.Background(), command.UpdateOrganization{
		ID:   created.ID(),
		Name: &emptyName,
	})
	if err != nil {
		t.Fatalf("Update(clear name without version precondition) error = %v", err)
	}
	if updated.Name() != "" || updated.Description() != description {
		t.Errorf("Update(clear name) = (%q, %q), want (empty, %q)", updated.Name(), updated.Description(), description)
	}
	if got, want := updated.Version(), types.Version(3); !got.Equal(want) {
		t.Errorf("Update(clear name) version = %s, want %s", got, want)
	}

	h.clock.now = testStartTime.Add(3 * time.Minute)
	_, err = h.service.Update(context.Background(), command.UpdateOrganization{
		ID:              created.ID(),
		ExpectedVersion: types.InitialVersion,
		Description:     ptr("stale write"),
	})
	if !errors.Is(err, organization.ErrVersionMismatch) {
		t.Fatalf("Update(stale version) error = %v, want %v", err, organization.ErrVersionMismatch)
	}
	stored, err := h.repository.Get(context.Background(), created.ID())
	if err != nil {
		t.Fatalf("repository.Get() error = %v", err)
	}
	if stored.Description() != description || !stored.Version().Equal(types.Version(3)) {
		t.Errorf("stored organization after stale write = description %q version %s", stored.Description(), stored.Version())
	}

	_, err = h.service.Update(context.Background(), command.UpdateOrganization{ID: created.ID()})
	if !errors.Is(err, organization.ErrNoOrganizationUpdates) {
		t.Fatalf("Update(no selected fields) error = %v, want %v", err, organization.ErrNoOrganizationUpdates)
	}
}

func TestOrganizationServiceUpdateSurfacesRepositoryCASConflict(t *testing.T) {
	t.Parallel()

	base := memory.NewOrganizationRepository()
	repository := &conflictingRepository{OrganizationRepository: base}
	authorizer := &fakeAuthorizer{allow: true}
	clock := &fakeClock{now: testStartTime}
	ids := &fakeIDGenerator{ids: fixtureIDs(4)}
	children := &fakeWorkspaceChildren{}
	service := newService(t, repository, authorizer, clock, ids, children, defaultServiceConfig())

	created, err := service.Create(context.Background(), command.CreateOrganization{Name: "Original"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	clock.now = testStartTime.Add(time.Minute)
	repository.conflictOnNextUpdate = true

	_, err = service.Update(context.Background(), command.UpdateOrganization{
		ID:          created.ID(),
		Description: ptr("losing update"),
	})
	if !errors.Is(err, ports.ErrOrganizationVersionConflict) {
		t.Fatalf("Update(concurrent write) error = %v, want %v", err, ports.ErrOrganizationVersionConflict)
	}

	stored, err := base.Get(context.Background(), created.ID())
	if err != nil {
		t.Fatalf("repository.Get() error = %v", err)
	}
	if stored.Description() != "concurrent update" || !stored.Version().Equal(types.Version(2)) {
		t.Errorf("winning organization = description %q version %s, want concurrent update version 2", stored.Description(), stored.Version())
	}
}

func TestOrganizationServiceDeleteAllowMissingChildBlockerAndPurgeTimes(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	missingID := fixtureID(99)

	deleted, err := h.service.Delete(context.Background(), command.DeleteOrganization{
		ID:           missingID,
		AllowMissing: true,
	})
	if err != nil || deleted != nil {
		t.Fatalf("Delete(missing, allow_missing) = (%v, %v), want (nil, nil)", deleted, err)
	}
	_, err = h.service.Delete(context.Background(), command.DeleteOrganization{ID: missingID})
	if !errors.Is(err, ports.ErrOrganizationNotFound) {
		t.Fatalf("Delete(missing) error = %v, want %v", err, ports.ErrOrganizationNotFound)
	}

	created, err := h.service.Create(context.Background(), command.CreateOrganization{Name: "Has workspaces"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	h.children.hasNonDeleted = true
	h.clock.now = testStartTime.Add(time.Hour)
	_, err = h.service.Delete(context.Background(), command.DeleteOrganization{
		ID:              created.ID(),
		ExpectedVersion: types.InitialVersion,
	})
	if !errors.Is(err, usecase.ErrOrganizationHasWorkspaces) {
		t.Fatalf("Delete(with children) error = %v, want %v", err, usecase.ErrOrganizationHasWorkspaces)
	}
	stored, err := h.repository.Get(context.Background(), created.ID())
	if err != nil {
		t.Fatalf("repository.Get() error = %v", err)
	}
	if stored.State() != organization.StateActive || stored.DeleteTime() != nil || stored.PurgeTime() != nil {
		t.Errorf("organization mutated despite child blocker: %+v", stored.Snapshot())
	}

	h.children.hasNonDeleted = false
	deleteTime := testStartTime.Add(2 * time.Hour)
	h.clock.now = deleteTime
	deleted, err = h.service.Delete(context.Background(), command.DeleteOrganization{
		ID:              created.ID(),
		ExpectedVersion: types.InitialVersion,
	})
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if deleted.State() != organization.StateDeleted {
		t.Errorf("Delete() state = %s, want DELETED", deleted.State())
	}
	if got := deleted.DeleteTime(); got == nil || !got.Equal(deleteTime.UTC()) {
		t.Errorf("Delete() delete time = %v, want %s", got, deleteTime.UTC())
	}
	wantPurgeTime := deleteTime.Add(h.retention).UTC()
	if got := deleted.PurgeTime(); got == nil || !got.Equal(wantPurgeTime) {
		t.Errorf("Delete() purge time = %v, want %s", got, wantPurgeTime)
	}
	if !deleted.UpdateTime().Equal(deleteTime.UTC()) || !deleted.Version().Equal(types.Version(2)) {
		t.Errorf("Delete() update time/version = %s/%s", deleted.UpdateTime(), deleted.Version())
	}

	unchanged, err := h.service.Delete(context.Background(), command.DeleteOrganization{
		ID:           created.ID(),
		AllowMissing: true,
	})
	if err != nil {
		t.Fatalf("Delete(already deleted, allow_missing) error = %v", err)
	}
	if unchanged == nil || !unchanged.Version().Equal(types.Version(2)) {
		t.Fatalf("Delete(already deleted, allow_missing) version = %v, want 2", unchanged)
	}
	_, err = h.service.Delete(context.Background(), command.DeleteOrganization{ID: created.ID()})
	if !errors.Is(err, organization.ErrOrganizationAlreadyDeleted) {
		t.Fatalf("Delete(already deleted) error = %v, want %v", err, organization.ErrOrganizationAlreadyDeleted)
	}
}

func TestOrganizationServiceUndeleteRestoresRetainedOrganization(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	created, err := h.service.Create(context.Background(), command.CreateOrganization{Name: "Recoverable"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	h.clock.now = testStartTime.Add(time.Hour)
	deleted, err := h.service.Delete(context.Background(), command.DeleteOrganization{ID: created.ID()})
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	undeleteTime := testStartTime.Add(2 * time.Hour)
	h.clock.now = undeleteTime
	restored, err := h.service.Undelete(context.Background(), command.UndeleteOrganization{ID: created.ID()})
	if err != nil {
		t.Fatalf("Undelete() error = %v", err)
	}
	if restored.State() != organization.StateActive {
		t.Errorf("Undelete() state = %s, want ACTIVE", restored.State())
	}
	if restored.DeleteTime() != nil || restored.PurgeTime() != nil {
		t.Errorf("Undelete() deletion timestamps = (%v, %v), want nil", restored.DeleteTime(), restored.PurgeTime())
	}
	if !restored.UpdateTime().Equal(undeleteTime.UTC()) || !restored.Version().Equal(types.Version(3)) {
		t.Errorf("Undelete() update time/version = %s/%s, want %s/3", restored.UpdateTime(), restored.Version(), undeleteTime.UTC())
	}
	if !restored.CreateTime().Equal(deleted.CreateTime()) {
		t.Errorf("Undelete() create time = %s, want preserved %s", restored.CreateTime(), deleted.CreateTime())
	}

	_, err = h.service.Undelete(context.Background(), command.UndeleteOrganization{ID: created.ID()})
	if !errors.Is(err, organization.ErrOrganizationNotDeleted) {
		t.Fatalf("Undelete(active) error = %v, want %v", err, organization.ErrOrganizationNotDeleted)
	}

	second, err := h.service.Create(context.Background(), command.CreateOrganization{Name: "Expired"})
	if err != nil {
		t.Fatalf("Create(Expired) error = %v", err)
	}
	h.clock.now = testStartTime.Add(3 * time.Hour)
	expired, err := h.service.Delete(context.Background(), command.DeleteOrganization{ID: second.ID()})
	if err != nil {
		t.Fatalf("Delete(Expired) error = %v", err)
	}
	h.clock.now = *expired.PurgeTime()
	_, err = h.service.Undelete(context.Background(), command.UndeleteOrganization{ID: second.ID()})
	if !errors.Is(err, organization.ErrPurgeTimePassed) {
		t.Fatalf("Undelete(at purge time) error = %v, want %v", err, organization.ErrPurgeTimePassed)
	}
}

func TestOrganizationServiceAuthorizationDeniesEveryActionBeforeWork(t *testing.T) {
	t.Parallel()

	id := fixtureID(77)
	tests := []struct {
		name   string
		action ports.AuthorizationAction
		id     organization.ID
		call   func(*usecase.OrganizationService) error
	}{
		{
			name: "create", action: ports.ActionCreateOrganization,
			call: func(service *usecase.OrganizationService) error {
				_, err := service.Create(context.Background(), command.CreateOrganization{Name: "Denied"})
				return err
			},
		},
		{
			name: "get", action: ports.ActionGetOrganization, id: id,
			call: func(service *usecase.OrganizationService) error {
				_, err := service.Get(context.Background(), query.GetOrganization{ID: id})
				return err
			},
		},
		{
			name: "list", action: ports.ActionListOrganizations,
			call: func(service *usecase.OrganizationService) error {
				_, err := service.List(context.Background(), query.ListOrganizations{})
				return err
			},
		},
		{
			name: "update", action: ports.ActionUpdateOrganization, id: id,
			call: func(service *usecase.OrganizationService) error {
				_, err := service.Update(context.Background(), command.UpdateOrganization{ID: id, Name: ptr("Denied")})
				return err
			},
		},
		{
			name: "delete", action: ports.ActionDeleteOrganization, id: id,
			call: func(service *usecase.OrganizationService) error {
				_, err := service.Delete(context.Background(), command.DeleteOrganization{ID: id, AllowMissing: true})
				return err
			},
		},
		{
			name: "undelete", action: ports.ActionUndeleteOrganization, id: id,
			call: func(service *usecase.OrganizationService) error {
				_, err := service.Undelete(context.Background(), command.UndeleteOrganization{ID: id})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &countingRepository{OrganizationRepository: memory.NewOrganizationRepository()}
			authorizer := &fakeAuthorizer{}
			ids := &fakeIDGenerator{ids: fixtureIDs(2)}
			children := &fakeWorkspaceChildren{}
			service := newService(
				t,
				repository,
				authorizer,
				&fakeClock{now: testStartTime},
				ids,
				children,
				defaultServiceConfig(),
			)

			err := tt.call(service)
			if !errors.Is(err, ports.ErrPermissionDenied) {
				t.Fatalf("call error = %v, want %v", err, ports.ErrPermissionDenied)
			}
			assertAuthorizationRequests(t, authorizer.requests, []ports.AuthorizationRequest{{
				Action:         tt.action,
				OrganizationID: tt.id,
			}})
			if repository.calls != 0 {
				t.Errorf("repository calls = %d, want 0", repository.calls)
			}
			if children.calls != 0 {
				t.Errorf("workspace child checks = %d, want 0", children.calls)
			}
			if ids.calls != 0 {
				t.Errorf("ID generator calls = %d, want 0", ids.calls)
			}
		})
	}
}

func TestOrganizationServiceAuthorizesAllActionsWithCorrectScope(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	created, err := h.service.Create(context.Background(), command.CreateOrganization{Name: "Lifecycle"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := h.service.Get(context.Background(), query.GetOrganization{ID: created.ID()}); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if _, err := h.service.List(context.Background(), query.ListOrganizations{}); err != nil {
		t.Fatalf("List() error = %v", err)
	}
	h.clock.now = testStartTime.Add(time.Minute)
	if _, err := h.service.Update(context.Background(), command.UpdateOrganization{
		ID: created.ID(), Name: ptr("Updated"),
	}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	h.clock.now = testStartTime.Add(2 * time.Minute)
	if _, err := h.service.Delete(context.Background(), command.DeleteOrganization{ID: created.ID()}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	h.clock.now = testStartTime.Add(3 * time.Minute)
	if _, err := h.service.Undelete(context.Background(), command.UndeleteOrganization{ID: created.ID()}); err != nil {
		t.Fatalf("Undelete() error = %v", err)
	}

	assertAuthorizationRequests(t, h.authorizer.requests, []ports.AuthorizationRequest{
		{Action: ports.ActionCreateOrganization},
		{Action: ports.ActionGetOrganization, OrganizationID: created.ID()},
		{Action: ports.ActionListOrganizations},
		{Action: ports.ActionUpdateOrganization, OrganizationID: created.ID()},
		{Action: ports.ActionDeleteOrganization, OrganizationID: created.ID()},
		{Action: ports.ActionUndeleteOrganization, OrganizationID: created.ID()},
	})
}

type harness struct {
	service    *usecase.OrganizationService
	repository *memory.OrganizationRepository
	authorizer *fakeAuthorizer
	clock      *fakeClock
	ids        *fakeIDGenerator
	children   *fakeWorkspaceChildren
	retention  time.Duration
}

func newHarness(t *testing.T) *harness {
	t.Helper()

	repository := memory.NewOrganizationRepository()
	authorizer := &fakeAuthorizer{allow: true}
	clock := &fakeClock{now: testStartTime}
	ids := &fakeIDGenerator{ids: fixtureIDs(32)}
	children := &fakeWorkspaceChildren{}
	config := defaultServiceConfig()

	return &harness{
		service:    newService(t, repository, authorizer, clock, ids, children, config),
		repository: repository,
		authorizer: authorizer,
		clock:      clock,
		ids:        ids,
		children:   children,
		retention:  config.SoftDeleteRetention,
	}
}

func defaultServiceConfig() usecase.OrganizationServiceConfig {
	return usecase.OrganizationServiceConfig{
		SoftDeleteRetention: 48 * time.Hour,
		PageTokenKey:        []byte(testPageTokenKey),
	}
}

func newService(
	t *testing.T,
	repository ports.OrganizationRepository,
	authorizer ports.Authorizer,
	clock ports.Clock,
	ids ports.IDGenerator,
	children ports.WorkspaceChildren,
	config usecase.OrganizationServiceConfig,
) *usecase.OrganizationService {
	t.Helper()

	service, err := usecase.NewOrganizationService(repository, authorizer, clock, ids, children, config)
	if err != nil {
		t.Fatalf("NewOrganizationService() error = %v", err)
	}
	return service
}

type fakeClock struct {
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	return c.now
}

type fakeIDGenerator struct {
	ids   []organization.ID
	calls int
}

func (g *fakeIDGenerator) NewID() organization.ID {
	g.calls++
	if len(g.ids) == 0 {
		return organization.ID{}
	}
	id := g.ids[0]
	g.ids = g.ids[1:]
	return id
}

type fakeAuthorizer struct {
	allow    bool
	scopeKey string
	requests []ports.AuthorizationRequest
}

func (a *fakeAuthorizer) Authorize(_ context.Context, request ports.AuthorizationRequest) error {
	a.requests = append(a.requests, request)
	if a.allow {
		return nil
	}
	return ports.ErrPermissionDenied
}

func (a *fakeAuthorizer) ScopeKey(context.Context) (string, error) {
	if !a.allow {
		return "", ports.ErrPermissionDenied
	}
	if a.scopeKey == "" {
		return "test:allow-all", nil
	}
	return a.scopeKey, nil
}

type fakeWorkspaceChildren struct {
	hasNonDeleted bool
	err           error
	calls         int
}

func (w *fakeWorkspaceChildren) HasNonDeleted(_ context.Context, _ organization.ID) (bool, error) {
	w.calls++
	return w.hasNonDeleted, w.err
}

type conflictingRepository struct {
	*memory.OrganizationRepository
	conflictOnNextUpdate bool
}

func (r *conflictingRepository) Update(
	ctx context.Context,
	value *organization.Organization,
	expectedVersion types.Version,
) error {
	if r.conflictOnNextUpdate {
		r.conflictOnNextUpdate = false
		current, err := r.OrganizationRepository.Get(ctx, value.ID())
		if err != nil {
			return err
		}
		winningDescription := "concurrent update"
		if err := current.Update(organization.UpdateParams{
			Description:     &winningDescription,
			Now:             current.UpdateTime().Add(time.Second),
			ExpectedVersion: current.Version(),
		}); err != nil {
			return err
		}
		if err := r.OrganizationRepository.Update(ctx, current, expectedVersion); err != nil {
			return err
		}
	}
	return r.OrganizationRepository.Update(ctx, value, expectedVersion)
}

type countingRepository struct {
	*memory.OrganizationRepository
	calls int
}

func (r *countingRepository) Create(ctx context.Context, value *organization.Organization) error {
	r.calls++
	return r.OrganizationRepository.Create(ctx, value)
}

func (r *countingRepository) Get(ctx context.Context, id organization.ID) (*organization.Organization, error) {
	r.calls++
	return r.OrganizationRepository.Get(ctx, id)
}

func (r *countingRepository) Update(
	ctx context.Context,
	value *organization.Organization,
	expectedVersion types.Version,
) error {
	r.calls++
	return r.OrganizationRepository.Update(ctx, value, expectedVersion)
}

func (r *countingRepository) List(
	ctx context.Context,
	options ports.ListOrganizationsOptions,
) (ports.ListOrganizationsResult, error) {
	r.calls++
	return r.OrganizationRepository.List(ctx, options)
}

func fixtureIDs(count int) []organization.ID {
	ids := make([]organization.ID, count)
	for i := range ids {
		ids[i] = fixtureID(i + 1)
	}
	return ids
}

func fixtureID(value int) organization.ID {
	return organization.MustParseID(fmt.Sprintf("00000000-0000-4000-8000-%012x", value))
}

func ptr[T any](value T) *T {
	return &value
}

func assertOrganizationNames(t *testing.T, values []*organization.Organization, want ...string) {
	t.Helper()
	got := make([]string, len(values))
	for i, value := range values {
		got[i] = value.Name()
	}
	if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Errorf("organization names = %v, want %v", got, want)
	}
}

func assertAuthorizationRequests(t *testing.T, got, want []ports.AuthorizationRequest) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("authorization requests = %v, want %v", got, want)
	}
	for i := range want {
		if got[i].Action != want[i].Action || !got[i].OrganizationID.Equal(want[i].OrganizationID) {
			t.Errorf("authorization request[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func tamperPageToken(t *testing.T, token string) string {
	t.Helper()
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		t.Fatalf("page token %q does not have payload and signature", token)
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode page token signature: %v", err)
	}
	if len(signature) == 0 {
		t.Fatal("page token signature is empty")
	}
	signature[0] ^= 0xff
	return parts[0] + "." + base64.RawURLEncoding.EncodeToString(signature)
}
