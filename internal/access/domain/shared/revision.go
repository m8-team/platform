package shared

import (
	"errors"
	"math"
	"strings"
)

var (
	ErrInvalidRevision       = errors.New("invalid aggregate revision")
	ErrRevisionConflict      = errors.New("aggregate revision conflict")
	ErrModelRevisionRequired = errors.New("authorization model revision is required")
)

// Revision is an optimistic-concurrency sequence for one aggregate, NOT a SpiceDB ZedToken.
type Revision int64

func NewRevision(value int64) (Revision, error) {
	r := Revision(value)
	return r, r.Validate()
}
func (r Revision) Validate() error {
	if r < 1 {
		return ErrInvalidRevision
	}
	return nil
}
func (r Revision) Int64() int64 { return int64(r) }
func (r Revision) Next(expected int64) (Revision, error) {
	if r.Validate() != nil {
		return 0, ErrInvalidRevision
	}
	if int64(r) != expected {
		return 0, ErrRevisionConflict
	}
	if r == Revision(math.MaxInt64) {
		return 0, ErrInvalidRevision
	}
	return r + 1, nil
}

// ModelRevision identifies a published authorization schema, distinct from graph/data revisions.
type ModelRevision struct{ value string }

func NewModelRevision(raw string) (ModelRevision, error) {
	m := ModelRevision{value: strings.TrimSpace(raw)}
	return m, m.Validate()
}
func (m ModelRevision) Validate() error {
	if m.value == "" {
		return ErrModelRevisionRequired
	}
	return nil
}
func (m ModelRevision) String() string                 { return m.value }
func (m ModelRevision) Equal(other ModelRevision) bool { return m.value == other.value }
