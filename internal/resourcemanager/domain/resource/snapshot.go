// Package resource defines Resource Manager-owned identity and revision facts.
// It deliberately has no dependency on protobuf, storage or Access.
package resource

import (
	"errors"
	"strings"
)

var ErrInvalidSnapshot = errors.New("invalid authoritative resource snapshot")

type Kind string

const (
	Organization     Kind = "organization"
	Workspace        Kind = "workspace"
	Environment      Kind = "environment"
	WorkloadService  Kind = "workload_service"
)

func (k Kind) Valid() bool {
	switch k {
	case Organization, Workspace, Environment, WorkloadService:
		return true
	default:
		return false
	}
}

type State string

const (
	StateCreating  State = "CREATING"
	StateActive    State = "ACTIVE"
	StateSuspended State = "SUSPENDED"
	StateArchived  State = "ARCHIVED"
	StateDeleting  State = "DELETING"
	StateDeleted   State = "DELETED"
	StateFailed    State = "FAILED"
)

func (s State) Valid() bool {
	switch s {
	case StateCreating, StateActive, StateSuspended, StateArchived, StateDeleting, StateDeleted, StateFailed:
		return true
	default:
		return false
	}
}

// Snapshot is an authoritative Resource Manager query result, not an Access
// authorization decision. Tenant, ownership, lifecycle and revision are read
// from the single writer. Deletion tombstones remain visible for projections.
type Snapshot struct {
	Kind           Kind
	ID             string
	TenantID       string
	OwnerKind      string
	OwnerID        string
	OrganizationID string
	State          State
	Revision       int64
}

func (s Snapshot) Validate() error {
	if !s.Kind.Valid() || strings.TrimSpace(s.ID) == "" ||
		strings.TrimSpace(s.TenantID) == "" || !s.State.Valid() || s.Revision <= 0 ||
		(strings.TrimSpace(s.OwnerKind) == "") != (strings.TrimSpace(s.OwnerID) == "") {
		return ErrInvalidSnapshot
	}
	return nil
}

func (s Snapshot) Active() bool { return s.State == StateActive }
