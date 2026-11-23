package shared

import (
 "errors"
 "strings"
)

var (
 ErrSubjectTypeRequired = errors.New("subject type is required")
 ErrSubjectIDRequired = errors.New("subject id is required")
)

// Subject is a reference to a principal owned by Identity, not a credential.
type Subject struct { subjectType, id string }

func NewSubject(subjectType, id string) (Subject, error) {
 s := Subject{subjectType: strings.TrimSpace(subjectType), id: strings.TrimSpace(id)}
 return s, s.Validate()
}
func (s Subject) Validate() error {
 if s.subjectType == "" { return ErrSubjectTypeRequired }
 if s.id == "" { return ErrSubjectIDRequired }
 return nil
}
func (s Subject) Type() string { return s.subjectType }
func (s Subject) ID() string { return s.id }
