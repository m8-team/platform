package role

import (
	"errors"
	"github.com/m8-team/platform/internal/access/domain/shared"
	"time"
)

type EventKind string

const (
	RoleCreated            EventKind = "RoleCreated"
	RolePermissionsChanged EventKind = "RolePermissionsChanged"
	RoleBindingCreated     EventKind = "RoleBindingCreated"
	RoleBindingRevoked     EventKind = "RoleBindingRevoked"
)

var ErrInvalidEvent = errors.New("invalid role domain event")

// Event is a business fact; application-layer outbox wraps it in delivery metadata.
type Event struct {
	Kind        EventKind
	TenantID    shared.TenantID
	AggregateID string
	Revision    int64
	OccurredAt  time.Time
}

func (e Event) Validate() error {
	if e.Kind != RoleCreated && e.Kind != RolePermissionsChanged && e.Kind != RoleBindingCreated && e.Kind != RoleBindingRevoked {
		return ErrInvalidEvent
	}
	if e.TenantID.Validate() != nil || !shared.ValidUUID(e.AggregateID) || e.Revision < 1 || e.OccurredAt.IsZero() {
		return ErrInvalidEvent
	}
	return nil
}
