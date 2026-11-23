package policyjournal

import (
	"errors"

	"github.com/m8-team/platform/internal/access/app/projection"
)

var ErrInvalidPolicyDelivery = errors.New("invalid policy projection event or delivery")

// Event is emitted by the SAME authoritative Access/YDB transaction as the
// policy Head and all fanout Jobs. A role change has one policyRevision and
// one per-binding projection Event with its own per-key revision.
type Event struct {
	PolicyRevision uint64
	Projection     projection.Event
}

func (e Event) Validate() error {
	if e.PolicyRevision == 0 || e.Projection.Validate() != nil || !isPolicyTarget(e.Projection.Key.Kind) {
		return ErrInvalidPolicyDelivery
	}
	return nil
}
