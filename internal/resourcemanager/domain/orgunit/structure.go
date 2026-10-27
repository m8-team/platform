// Package orgunit contains pure organizational hierarchy invariants. It never
// stores employee memberships: those are temporal relationships owned by Identity.
package orgunit

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

var (
	ErrInvalidUnit       = errors.New("invalid org unit")
	ErrDuplicateUnit     = errors.New("duplicate org unit")
	ErrMissingParent     = errors.New("org unit parent not found")
	ErrCrossOrganization = errors.New("cross-organization org unit parent")
	ErrHierarchyCycle    = errors.New("org unit hierarchy cycle")
	ErrUnitNotFound      = errors.New("org unit not found")
	ErrParentInactive    = errors.New("org unit parent is not active")
	ErrUnitInactive      = errors.New("org unit is not active")
	ErrRevisionConflict  = errors.New("org structure revision conflict")
	ErrRevisionOverflow  = errors.New("org structure revision overflow")
)

type State string

const (
	StatePlanned State = "PLANNED"
	StateActive  State = "ACTIVE"
	StateClosed  State = "CLOSED"
)

// Unit is the adjacency-list input. ID is stable across renames/reparenting.
// There is no PersonID or employee assignment in this structure.
type Unit struct {
	ID             string
	OrganizationID string
	ParentID       string // empty for an organizational root
	State          State
}

// ClosureRow is an immutable materialized hierarchy read model. Self edges
// always have Depth == 0. OrganizationID scopes each ancestry relationship.
type ClosureRow struct {
	OrganizationID string
	AncestorID     string
	DescendantID   string
	Depth          uint32
}

// Structure is a versioned, pure snapshot of an organization's hierarchy.
type Structure struct {
	OrganizationID string
	Revision       uint64
	Units          []Unit
	Closure        []ClosureRow
}

// Build validates an adjacency list and materializes all ancestor paths.
// It rejects orphans, cycles, duplicate identifiers and cross-org references.
// The output is stable independent of input ordering and never aliases inputs.
func Build(organizationID string, revision uint64, units []Unit) (Structure, error) {
	if organizationID == "" || revision == 0 {
		return Structure{}, ErrInvalidUnit
	}
	byID := make(map[string]Unit, len(units))
	for _, u := range units {
		if u.ID == "" || u.OrganizationID != organizationID || !validState(u.State) || u.ID == u.ParentID {
			return Structure{}, fmt.Errorf("%w: %q", ErrInvalidUnit, u.ID)
		}
		if _, exists := byID[u.ID]; exists {
			return Structure{}, fmt.Errorf("%w: %s", ErrDuplicateUnit, u.ID)
		}
		byID[u.ID] = u
	}
	ids := make([]string, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	rows := make([]ClosureRow, 0, len(units))
	for _, id := range ids {
		visited := map[string]struct{}{}
		current := id
		for depth := uint32(0); ; depth++ {
			if _, exists := visited[current]; exists {
				return Structure{}, fmt.Errorf("%w: %s", ErrHierarchyCycle, id)
			}
			visited[current] = struct{}{}
			ancestor, ok := byID[current]
			if !ok {
				return Structure{}, fmt.Errorf("%w: %s -> %s", ErrMissingParent, id, current)
			}
			rows = append(rows, ClosureRow{OrganizationID: organizationID, AncestorID: current, DescendantID: id, Depth: depth})
			if ancestor.ParentID == "" {
				break
			}
			current = ancestor.ParentID
		}
	}
	result := make([]Unit, 0, len(ids))
	for _, id := range ids {
		result = append(result, byID[id])
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].DescendantID != rows[j].DescendantID {
			return rows[i].DescendantID < rows[j].DescendantID
		}
		return rows[i].Depth < rows[j].Depth
	})
	return Structure{OrganizationID: organizationID, Revision: revision, Units: result, Closure: rows}, nil
}

// Reparent returns a new revision without mutating the supplied snapshot.
// It is the pure domain part of a ChangeSet, not a persistence/API operation.
func Reparent(s Structure, unitID, newParentID string, expectedRevision uint64) (Structure, error) {
	if expectedRevision == 0 || s.Revision != expectedRevision {
		return Structure{}, ErrRevisionConflict
	}
	if s.Revision == math.MaxUint64 {
		return Structure{}, ErrRevisionOverflow
	}
	copyUnits := append([]Unit(nil), s.Units...)
	index, parentIndex := -1, -1
	for i, u := range copyUnits {
		if u.ID == unitID {
			index = i
		}
		if u.ID == newParentID {
			parentIndex = i
		}
	}
	if index == -1 {
		return Structure{}, ErrUnitNotFound
	}
	if copyUnits[index].State != StateActive {
		return Structure{}, ErrUnitInactive
	}
	if copyUnits[index].ParentID == newParentID {
		return Build(s.OrganizationID, s.Revision, copyUnits)
	}
	if newParentID != "" {
		if parentIndex == -1 {
			return Structure{}, ErrMissingParent
		}
		if copyUnits[parentIndex].OrganizationID != copyUnits[index].OrganizationID {
			return Structure{}, ErrCrossOrganization
		}
		if copyUnits[parentIndex].State != StateActive {
			return Structure{}, ErrParentInactive
		}
	}
	copyUnits[index].ParentID = newParentID
	return Build(s.OrganizationID, s.Revision+1, copyUnits)
}

func validState(s State) bool { return s == StatePlanned || s == StateActive || s == StateClosed }
