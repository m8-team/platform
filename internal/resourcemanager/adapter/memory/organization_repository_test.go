package memory_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/resourcemanager/adapter/memory"
	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
)

type organizationRepositoryFactory func() ports.OrganizationRepository

func TestOrganizationRepositoryContract(t *testing.T) {
	runOrganizationRepositoryContract(t, func() ports.OrganizationRepository {
		return memory.NewOrganizationRepository()
	})
}

func runOrganizationRepositoryContract(t *testing.T, factory organizationRepositoryFactory) {
	t.Helper()

	t.Run("create and get use defensive copies", func(t *testing.T) {
		repository := factory()
		now := time.Date(2026, 7, 19, 8, 0, 0, 0, time.UTC)
		original := newOrganization(t, organizationID(1), "original", now, map[string]string{"team": "platform"})

		if err := repository.Create(context.Background(), original); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if err := repository.Create(context.Background(), original); !errors.Is(err, ports.ErrOrganizationAlreadyExists) {
			t.Fatalf("duplicate Create() error = %v, want %v", err, ports.ErrOrganizationAlreadyExists)
		}

		renameOrganization(t, original, "changed outside repository", now.Add(time.Minute))
		stored := getOrganization(t, repository, original.ID())
		if stored.Name() != "original" {
			t.Fatalf("stored Name() = %q, want original", stored.Name())
		}
		if stored == original {
			t.Fatal("Get() returned the caller-owned aggregate pointer")
		}

		renameOrganization(t, stored, "changed detached copy", now.Add(2*time.Minute))
		again := getOrganization(t, repository, original.ID())
		if again.Name() != "original" {
			t.Fatalf("Name() after detached mutation = %q, want original", again.Name())
		}
		if again == stored {
			t.Fatal("consecutive Get() calls returned the same aggregate pointer")
		}

		labels := again.Labels()
		labels["team"] = "other"
		if got := getOrganization(t, repository, original.ID()).Labels()["team"]; got != "platform" {
			t.Fatalf("stored label = %q, want platform", got)
		}
	})

	t.Run("update is an atomic compare and swap", func(t *testing.T) {
		repository := factory()
		now := time.Date(2026, 7, 19, 9, 0, 0, 0, time.UTC)
		created := newOrganization(t, organizationID(2), "before", now, nil)
		mustCreate(t, repository, created)

		first := getOrganization(t, repository, created.ID())
		second := getOrganization(t, repository, created.ID())
		expectedVersion := first.Version()
		renameOrganization(t, first, "first", now.Add(time.Minute))
		renameOrganization(t, second, "second", now.Add(time.Minute))

		if err := repository.Update(context.Background(), first, expectedVersion); err != nil {
			t.Fatalf("first Update() error = %v", err)
		}
		err := repository.Update(context.Background(), second, expectedVersion)
		if !errors.Is(err, ports.ErrOrganizationVersionConflict) {
			t.Fatalf("stale Update() error = %v, want %v", err, ports.ErrOrganizationVersionConflict)
		}
		var conflict *ports.OrganizationVersionConflictError
		if !errors.As(err, &conflict) {
			t.Fatalf("stale Update() error type = %T, want *OrganizationVersionConflictError", err)
		}
		if !conflict.Expected.Equal(expectedVersion) || !conflict.Actual.Equal(first.Version()) {
			t.Fatalf("conflict versions = expected %s actual %s", conflict.Expected, conflict.Actual)
		}

		stored := getOrganization(t, repository, created.ID())
		if stored.Name() != "first" || !stored.Version().Equal(first.Version()) {
			t.Fatalf("stored organization = name %q version %s", stored.Name(), stored.Version())
		}
	})

	t.Run("concurrent updates allow exactly one winner", func(t *testing.T) {
		repository := factory()
		now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
		created := newOrganization(t, organizationID(3), "before", now, nil)
		mustCreate(t, repository, created)

		left := getOrganization(t, repository, created.ID())
		right := getOrganization(t, repository, created.ID())
		expectedVersion := left.Version()
		renameOrganization(t, left, "left", now.Add(time.Minute))
		renameOrganization(t, right, "right", now.Add(time.Minute))

		start := make(chan struct{})
		results := make(chan error, 2)
		var workers sync.WaitGroup
		workers.Add(2)
		for _, candidate := range []*organization.Organization{left, right} {
			candidate := candidate
			go func() {
				defer workers.Done()
				<-start
				results <- repository.Update(context.Background(), candidate, expectedVersion)
			}()
		}
		close(start)
		workers.Wait()
		close(results)

		successes := 0
		conflicts := 0
		for err := range results {
			switch {
			case err == nil:
				successes++
			case errors.Is(err, ports.ErrOrganizationVersionConflict):
				conflicts++
			default:
				t.Fatalf("concurrent Update() unexpected error = %v", err)
			}
		}
		if successes != 1 || conflicts != 1 {
			t.Fatalf("concurrent results = %d successes, %d conflicts; want 1 and 1", successes, conflicts)
		}

		stored := getOrganization(t, repository, created.ID())
		if !stored.Version().Equal(left.Version()) {
			t.Fatalf("stored Version() = %s, want %s", stored.Version(), left.Version())
		}
		if stored.Name() != "left" && stored.Name() != "right" {
			t.Fatalf("stored Name() = %q, want left or right", stored.Name())
		}
	})

	t.Run("update requires the next aggregate version", func(t *testing.T) {
		repository := factory()
		now := time.Date(2026, 7, 19, 11, 0, 0, 0, time.UTC)
		created := newOrganization(t, organizationID(4), "unchanged", now, nil)
		mustCreate(t, repository, created)

		detached := getOrganization(t, repository, created.ID())
		err := repository.Update(context.Background(), detached, detached.Version())
		if !errors.Is(err, ports.ErrInvalidOrganizationVersion) {
			t.Fatalf("Update() error = %v, want %v", err, ports.ErrInvalidOrganizationVersion)
		}
	})

	t.Run("list filters orders and paginates with a keyset cursor", func(t *testing.T) {
		repository := factory()
		baseTime := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
		organizations := []*organization.Organization{
			newOrganization(t, organizationID(11), "beta", baseTime, map[string]string{"team": "platform"}),
			newOrganization(t, organizationID(12), "alpha", baseTime.Add(time.Minute), map[string]string{"team": "platform"}),
			newOrganization(t, organizationID(13), "alpha", baseTime.Add(2*time.Minute), map[string]string{"team": "platform"}),
			newOrganization(t, organizationID(14), "gamma", baseTime.Add(3*time.Minute), map[string]string{"team": "other"}),
		}
		for _, value := range organizations {
			mustCreate(t, repository, value)
		}

		deleted := getOrganization(t, repository, organizations[3].ID())
		deletedVersion := deleted.Version()
		if err := deleted.Delete(organization.DeleteParams{
			Now:             baseTime.Add(4 * time.Minute),
			PurgeTime:       baseTime.Add(24 * time.Hour),
			ExpectedVersion: deletedVersion,
		}); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}
		if err := repository.Update(context.Background(), deleted, deletedVersion); err != nil {
			t.Fatalf("persist deleted organization: %v", err)
		}

		options := ports.ListOrganizationsOptions{
			Filter: ports.OrganizationFilter{
				LabelsEqual: map[string]string{"team": "platform"},
			},
			Order: ports.OrganizationOrder{
				Field:     ports.OrganizationOrderFieldName,
				Direction: ports.SortDirectionAscending,
			},
			PageSize: 2,
		}
		firstPage, err := repository.List(context.Background(), options)
		if err != nil {
			t.Fatalf("first List() error = %v", err)
		}
		assertOrganizationIDs(t, firstPage.Organizations, organizationID(12), organizationID(13))
		if firstPage.TotalSize != 3 {
			t.Fatalf("first TotalSize = %d, want 3", firstPage.TotalSize)
		}
		if firstPage.Next == nil {
			t.Fatal("first Next = nil, want cursor")
		}

		options.After = firstPage.Next
		secondPage, err := repository.List(context.Background(), options)
		if err != nil {
			t.Fatalf("second List() error = %v", err)
		}
		assertOrganizationIDs(t, secondPage.Organizations, organizationID(11))
		if secondPage.TotalSize != 3 {
			t.Fatalf("second TotalSize = %d, want 3", secondPage.TotalSize)
		}
		if secondPage.Next != nil {
			t.Fatalf("second Next = %+v, want nil", secondPage.Next)
		}

		renameOrganization(t, firstPage.Organizations[0], "detached", baseTime.Add(5*time.Minute))
		if got := getOrganization(t, repository, organizationID(12)).Name(); got != "alpha" {
			t.Fatalf("stored Name() after List result mutation = %q, want alpha", got)
		}

		deletedOnly, err := repository.List(context.Background(), ports.ListOrganizationsOptions{
			Filter: ports.OrganizationFilter{
				States:      []organization.State{organization.StateDeleted},
				ShowDeleted: true,
			},
		})
		if err != nil {
			t.Fatalf("deleted List() error = %v", err)
		}
		assertOrganizationIDs(t, deletedOnly.Organizations, organizationID(14))
	})

	t.Run("not found invalid input and cancellation are explicit", func(t *testing.T) {
		repository := factory()
		missingID := organizationID(99)
		if _, err := repository.Get(context.Background(), missingID); !errors.Is(err, ports.ErrOrganizationNotFound) {
			t.Fatalf("Get() error = %v, want %v", err, ports.ErrOrganizationNotFound)
		}
		if err := repository.Create(context.Background(), nil); !errors.Is(err, ports.ErrNilOrganization) {
			t.Fatalf("Create(nil) error = %v, want %v", err, ports.ErrNilOrganization)
		}
		if _, err := repository.List(context.Background(), ports.ListOrganizationsOptions{PageSize: -1}); !errors.Is(err, ports.ErrInvalidListOrganizationsOptions) {
			t.Fatalf("List(invalid page size) error = %v, want %v", err, ports.ErrInvalidListOrganizationsOptions)
		}
		if _, err := repository.List(context.Background(), ports.ListOrganizationsOptions{
			Order: ports.OrganizationOrder{Field: "unknown"},
		}); !errors.Is(err, ports.ErrInvalidListOrganizationsOptions) {
			t.Fatalf("List(invalid order) error = %v, want %v", err, ports.ErrInvalidListOrganizationsOptions)
		}

		cancelled, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := repository.List(cancelled, ports.ListOrganizationsOptions{}); !errors.Is(err, context.Canceled) {
			t.Fatalf("List(cancelled) error = %v, want %v", err, context.Canceled)
		}
	})
}

func newOrganization(
	t *testing.T,
	id organization.ID,
	name string,
	now time.Time,
	labels map[string]string,
) *organization.Organization {
	t.Helper()

	value, err := organization.New(organization.CreateParams{
		ID:          id,
		Name:        name,
		Description: name + " description",
		Labels:      labels,
		Now:         now,
	})
	if err != nil {
		t.Fatalf("organization.New() error = %v", err)
	}
	return value
}

func renameOrganization(t *testing.T, value *organization.Organization, name string, now time.Time) {
	t.Helper()

	if err := value.Update(organization.UpdateParams{
		Name:            &name,
		Now:             now,
		ExpectedVersion: value.Version(),
	}); err != nil {
		t.Fatalf("Organization.Update() error = %v", err)
	}
}

func mustCreate(t *testing.T, repository ports.OrganizationRepository, value *organization.Organization) {
	t.Helper()
	if err := repository.Create(context.Background(), value); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}

func getOrganization(
	t *testing.T,
	repository ports.OrganizationRepository,
	id organization.ID,
) *organization.Organization {
	t.Helper()
	value, err := repository.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	return value
}

func assertOrganizationIDs(t *testing.T, values []*organization.Organization, expected ...organization.ID) {
	t.Helper()
	if len(values) != len(expected) {
		t.Fatalf("organization count = %d, want %d", len(values), len(expected))
	}
	for i := range expected {
		if !values[i].ID().Equal(expected[i]) {
			t.Fatalf("organization[%d].ID() = %s, want %s", i, values[i].ID(), expected[i])
		}
	}
}

func organizationID(sequence int) organization.ID {
	return organization.MustParseID(fmt.Sprintf("00000000-0000-4000-8000-%012d", sequence))
}
