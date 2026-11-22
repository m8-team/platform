package spicedb

import (
	"encoding/base64"
	"regexp"
	"time"

	"github.com/m8-team/platform/internal/access/domain"
)

// Object types must match the deployed SpiceDB schema. Namespaces are scoped
// by encoded IDs, not by dynamically generated object types.
var graphName = regexp.MustCompile(`^[a-z][a-z0-9_]{1,62}[a-z0-9]$`)

// Prefixing both ends of every tuple prevents subjects and objects from being
// shared across tenants even when their public IDs are identical.
func graphID(tenant domain.TenantID, id string) (string, error) {
	if err := tenant.Validate(); err != nil {
		return "", err
	}
	if id == "" {
		return "", ErrInvalidGraphReference
	}
	encoded := string(tenant) + "/" + base64.RawURLEncoding.EncodeToString([]byte(id))
	if len(encoded) > 1024 {
		return "", ErrInvalidGraphReference
	}
	return encoded, nil
}

type objectReference struct {
	ObjectType string `json:"objectType"`
	ObjectID   string `json:"objectId"`
}

type subjectReference struct {
	Object           objectReference `json:"object"`
	OptionalRelation string          `json:"optionalRelation,omitempty"`
}

func graphObject(tenant domain.TenantID, typ, id string) (objectReference, error) {
	if !graphName.MatchString(typ) {
		return objectReference{}, ErrInvalidGraphReference
	}
	graphID, err := graphID(tenant, id)
	if err != nil {
		return objectReference{}, err
	}
	return objectReference{ObjectType: typ, ObjectID: graphID}, nil
}

func graphSubject(tenant domain.TenantID, s domain.Subject, relation string) (subjectReference, error) {
	if err := s.Validate(); err != nil {
		return subjectReference{}, err
	}
	obj, err := graphObject(tenant, s.Type(), s.ID())
	if err != nil {
		return subjectReference{}, err
	}
	if relation != "" && !graphName.MatchString(relation) {
		return subjectReference{}, ErrInvalidGraphReference
	}
	return subjectReference{Object: obj, OptionalRelation: relation}, nil
}

type caveatReference struct {
	CaveatName string `json:"caveatName"`
}
type graphRelationship struct {
	Resource          objectReference  `json:"resource"`
	Relation          string           `json:"relation"`
	Subject           subjectReference `json:"subject"`
	OptionalCaveat    *caveatReference `json:"optionalCaveat,omitempty"`
	OptionalExpiresAt *time.Time       `json:"optionalExpiresAt,omitempty"`
}

func mapRelationship(r domain.Relationship) (graphRelationship, error) {
	if err := r.Validate(); err != nil {
		return graphRelationship{}, err
	}
	resource, err := graphObject(r.TenantID(), r.Resource().Type(), r.Resource().ID())
	if err != nil {
		return graphRelationship{}, err
	}
	subject, err := graphSubject(r.TenantID(), r.Subject(), r.SubjectRelation())
	if err != nil {
		return graphRelationship{}, err
	}
	if !graphName.MatchString(r.Relation()) {
		return graphRelationship{}, ErrInvalidGraphReference
	}
	rel := graphRelationship{Resource: resource, Subject: subject, Relation: r.Relation()}
	if !r.ExpiresAt().IsZero() {
		expires := r.ExpiresAt().UTC()
		rel.OptionalExpiresAt = &expires
	}
	if r.ConditionRef() != "" {
		if !graphName.MatchString(r.ConditionRef()) {
			return graphRelationship{}, ErrInvalidGraphReference
		}
		// Caveats are declared by trusted M8 policy code, never client input.
		// Runtime caveat context is not accepted by this adapter yet; an
		// unresolved conditional check is always a denial.
		rel.OptionalCaveat = &caveatReference{CaveatName: r.ConditionRef()}
	}
	return rel, nil
}
