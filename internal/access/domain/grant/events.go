package grant

import (
	"errors"
	"github.com/m8-team/platform/internal/access/domain/shared"
	"time"
)

type EventKind string

const (
	GrantRequested EventKind = "AccessGrantRequested"
	GrantSubmitted EventKind = "AccessGrantSubmitted"
	GrantApproved  EventKind = "AccessGrantApproved"
	GrantActivated EventKind = "AccessGrantActivated"
	GrantRejected  EventKind = "AccessGrantRejected"
	GrantRevoked   EventKind = "AccessGrantRevoked"
	GrantExpired   EventKind = "AccessGrantExpired"
)

var ErrInvalidEvent = errors.New("invalid grant domain event")

type Event struct {
	Kind        EventKind
	TenantID    shared.TenantID
	AggregateID string
	Revision    int64
	OccurredAt  time.Time
}

func (e Event) Validate() error {
	switch e.Kind {
	case GrantRequested, GrantSubmitted, GrantApproved, GrantActivated, GrantRejected, GrantRevoked, GrantExpired:
	default:
		return ErrInvalidEvent
	}
	if e.TenantID.Validate() != nil || !shared.ValidUUID(e.AggregateID) || e.Revision < 1 || e.OccurredAt.IsZero() {
		return ErrInvalidEvent
	}
	return nil
}
