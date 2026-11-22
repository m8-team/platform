package spicedb

import (
	"encoding/base64"
	"regexp"

	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/m8-team/platform/internal/access/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var graphName = regexp.MustCompile(`^[a-z][a-z0-9_]{1,62}[a-z0-9]$`)

// Namespaces remain global in SpiceDB. All object identifiers are partitioned
// by tenant; the authoritative owner must be verified by the calling service.
func graphID(tenant domain.TenantID, id string) (string, error) {
	if err := tenant.Validate(); err != nil {
		return "", err
	}
	if id == "" {
		return "", ErrInvalidGraphReference
	}
	result := string(tenant) + "/" + base64.RawURLEncoding.EncodeToString([]byte(id))
	if len(result) > 1024 {
		return "", ErrInvalidGraphReference
	}
	return result, nil
}

func graphObject(tenant domain.TenantID, typ, id string) (*pb.ObjectReference, error) {
	if !graphName.MatchString(typ) {
		return nil, ErrInvalidGraphReference
	}
	encoded, err := graphID(tenant, id)
	if err != nil {
		return nil, err
	}
	return &pb.ObjectReference{ObjectType: typ, ObjectId: encoded}, nil
}

func graphSubject(tenant domain.TenantID, s domain.Subject, relation string) (*pb.SubjectReference, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	obj, err := graphObject(tenant, s.Type(), s.ID())
	if err != nil {
		return nil, err
	}
	if relation != "" && !graphName.MatchString(relation) {
		return nil, ErrInvalidGraphReference
	}
	return &pb.SubjectReference{Object: obj, OptionalRelation: relation}, nil
}

func mapRelationship(r domain.Relationship) (*pb.Relationship, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	resource, err := graphObject(r.TenantID(), r.Resource().Type(), r.Resource().ID())
	if err != nil {
		return nil, err
	}
	subject, err := graphSubject(r.TenantID(), r.Subject(), r.SubjectRelation())
	if err != nil {
		return nil, err
	}
	if !graphName.MatchString(r.Relation()) {
		return nil, ErrInvalidGraphReference
	}
	out := &pb.Relationship{Resource: resource, Relation: r.Relation(), Subject: subject}
	if r.ConditionRef() != "" {
		if !graphName.MatchString(r.ConditionRef()) {
			return nil, ErrInvalidGraphReference
		}
		out.OptionalCaveat = &pb.ContextualizedCaveat{CaveatName: r.ConditionRef()}
	}
	if !r.ExpiresAt().IsZero() {
		out.OptionalExpiresAt = timestamppb.New(r.ExpiresAt())
		if err := out.OptionalExpiresAt.CheckValid(); err != nil {
			return nil, ErrInvalidGraphReference
		}
	}
	return out, nil
}
