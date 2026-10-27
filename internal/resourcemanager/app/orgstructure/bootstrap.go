package orgstructure

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/m8-team/platform/internal/resourcemanager/domain/orgunit"
)

// BootstrapRepository contains an idempotent initializer and a snapshot reader.
// Its YDB implementation owns the active head and immutable revision-1 snapshot.
type BootstrapRepository interface {
    HierarchyReader
    Initialize(ctx context.Context, tenantID string, structure orgunit.Structure, now time.Time) error
}

// EnsureOrganizationInitialized provisions an empty OrgUnit graph. Organization
// itself is the container; we do not invent an artificial root OrgUnit.
// Retrying creation after a lost acknowledgement must not reset existing units.
func EnsureOrganizationInitialized(ctx context.Context, store BootstrapRepository, tenantID, organizationID string, at time.Time) error {
    if store == nil || tenantID == "" || organizationID == "" || at.IsZero() {
        return orgunit.ErrInvalidUnit
    }
    if err := ctx.Err(); err != nil { return err }
    existing, err := store.Load(ctx, tenantID, organizationID)
    if err == nil {
        if existing.OrganizationID != organizationID || existing.Revision == 0 {
            return fmt.Errorf("invalid existing structure for %s", organizationID)
        }
        return nil
    }
    if !errors.Is(err, ErrStoreNotFound) { return err }
    initial, err := orgunit.Build(organizationID, 1, nil)
    if err != nil { return err }
    if err = store.Initialize(ctx, tenantID, initial, at.UTC()); err == nil { return nil }
    // Another creator may have published the same empty initial structure.
    // Do not hide arbitrary errors: only resolve racing duplicate creates
    // when the current head is valid and readable.
    observed, readErr := store.Load(ctx, tenantID, organizationID)
    if readErr == nil && observed.OrganizationID == organizationID && observed.Revision >= 1 {
        return nil
    }
    return err
}
