package domain

import "strings"

// ModelRevision identifies an M8 authorization model version. A ModelRevision
// is distinct from a SpiceDB ZedToken (the graph data snapshot revision).
type ModelRevision struct {
	value string
}

func NewModelRevision(value string) (ModelRevision, error) {
	revision := ModelRevision{value: strings.TrimSpace(value)}
	if err := revision.Validate(); err != nil {
		return ModelRevision{}, err
	}
	return revision, nil
}

func (r ModelRevision) Validate() error {
	if r.value == "" {
		return ErrModelRevisionRequired
	}
	return nil
}

func (r ModelRevision) String() string {
	return r.value
}

func (r ModelRevision) Equal(other ModelRevision) bool {
	return r.value == other.value
}
