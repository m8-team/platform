package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidRelationship = errors.New("invalid relationship")
	ErrInvalidRelationName = errors.New("invalid relation name")
)

var relationNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,127}$`)

// Relationship is a tenant-bound authorization tuple, independent of SpiceDB.
// The adapter must encode the tenant into *both* graph resource and subject
// namespaces where appropriate; a tuple's TenantID alone is not a graph fence.
type Relationship struct {
	tenantID        TenantID
	resource        Resource
	relation        string
	subject         Subject
	subjectRelation string
	conditionRef    string
	expiresAt       time.Time
}

type RelationshipInput struct {
	TenantID        TenantID
	Resource        Resource
	Relation        string
	Subject         Subject
	SubjectRelation string    // e.g. "member" for an indirect group subject
	ConditionRef    string    // trusted, versioned policy name; never executable input
	ExpiresAt       time.Time // optional SpiceDB relationship expiration (UTC)
}

func NewRelationship(input RelationshipInput) (Relationship, error) {
	r := Relationship{
		tenantID:        input.TenantID,
		resource:        input.Resource,
		relation:        strings.TrimSpace(input.Relation),
		subject:         input.Subject,
		subjectRelation: strings.TrimSpace(input.SubjectRelation),
		conditionRef:    strings.TrimSpace(input.ConditionRef),
		expiresAt:       input.ExpiresAt.UTC(),
	}
	return r, r.Validate()
}

func (r Relationship) Validate() error {
	if err := r.tenantID.Validate(); err != nil {
		return err
	}
	if err := r.resource.Validate(); err != nil {
		return err
	}
	if err := r.subject.Validate(); err != nil {
		return err
	}
	if !relationNamePattern.MatchString(r.relation) {
		return ErrInvalidRelationName
	}
	if r.subjectRelation != "" && !relationNamePattern.MatchString(r.subjectRelation) {
		return ErrInvalidRelationName
	}
	if r.conditionRef != "" && !relationNamePattern.MatchString(r.conditionRef) {
		return ErrInvalidRelationship
	}
	return nil
}

// RelationshipKey represents SpiceDB tuple identity. Caveat conditions are not
// part of tuple identity and cannot make two otherwise equal tuples distinct.
type RelationshipKey struct {
	TenantID        TenantID
	ResourceType    string
	ResourceID      string
	Relation        string
	SubjectType     string
	SubjectID       string
	SubjectRelation string
}

func (r Relationship) Key() RelationshipKey {
	return RelationshipKey{r.tenantID, r.resource.Type(), r.resource.ID(), r.relation,
		r.subject.Type(), r.subject.ID(), r.subjectRelation}
}

func (r Relationship) TenantID() TenantID      { return r.tenantID }
func (r Relationship) Resource() Resource      { return r.resource }
func (r Relationship) Relation() string        { return r.relation }
func (r Relationship) Subject() Subject        { return r.subject }
func (r Relationship) SubjectRelation() string { return r.subjectRelation }
func (r Relationship) ConditionRef() string    { return r.conditionRef }
func (r Relationship) ExpiresAt() time.Time    { return r.expiresAt }
