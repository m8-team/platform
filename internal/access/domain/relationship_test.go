package domain

import (
	"errors"
	"testing"
)

const testTenantID = "11111111-1111-1111-1111-111111111111"
const testOtherTenantID = "22222222-2222-2222-2222-222222222222"
const testOrganizationID = "33333333-3333-3333-3333-333333333333"
const testGrantID = "44444444-4444-4444-4444-444444444444"
const testRoleID = "55555555-5555-5555-5555-555555555555"
const testBindingID = "66666666-6666-6666-6666-666666666666"

func domainSubject(t *testing.T) Subject {
	t.Helper()
	s, err := NewSubject("user", "u1")
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func domainResource(t *testing.T) Resource {
	t.Helper()
	r, err := NewResource("workspace", "w1")
	if err != nil {
		t.Fatal(err)
	}
	return r
}
func domainPermission(t *testing.T, name string) Permission {
	t.Helper()
	p, err := NewPermission(name)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRelationshipRequiresTenantAndValidReferences(t *testing.T) {
	valid := RelationshipInput{TenantID: TenantID(testTenantID), Resource: domainResource(t), Relation: "viewer", Subject: domainSubject(t)}
	got, err := NewRelationship(valid)
	if err != nil {
		t.Fatal(err)
	}
	if got.Key().TenantID != TenantID(testTenantID) || got.Key().Relation != "viewer" {
		t.Fatalf("unexpected tuple key: %+v", got.Key())
	}
	invalid := valid
	invalid.TenantID = ""
	if _, err := NewRelationship(invalid); !errors.Is(err, ErrInvalidTenantID) {
		t.Fatalf("missing tenant: %v", err)
	}
	invalid = valid
	invalid.Relation = "Viewer/Write"
	if _, err := NewRelationship(invalid); !errors.Is(err, ErrInvalidRelationName) {
		t.Fatalf("invalid relation: %v", err)
	}
	invalid = valid
	invalid.Subject = Subject{}
	if _, err := NewRelationship(invalid); !errors.Is(err, ErrSubjectTypeRequired) {
		t.Fatalf("missing subject: %v", err)
	}
}

func TestRelationshipKeyExcludesCondition(t *testing.T) {
	input := RelationshipInput{TenantID: TenantID(testTenantID), Resource: domainResource(t), Relation: "viewer", Subject: domainSubject(t)}
	plain, err := NewRelationship(input)
	if err != nil {
		t.Fatal(err)
	}
	input.ConditionRef = "business_hours"
	conditional, err := NewRelationship(input)
	if err != nil {
		t.Fatal(err)
	}
	if conditional.Key() != plain.Key() {
		t.Fatal("caveat must not alter SpiceDB relationship identity")
	}
}

func TestTenantIDMustBeCanonicalUUID(t *testing.T) {
	for _, id := range []string{"", "other", "11111111-1111-1111-1111-11111111111G", "11111111-1111-1111-1111-11111111111A"} {
		if _, err := NewTenantID(id); !errors.Is(err, ErrInvalidTenantID) {
			t.Fatalf("tenant %q: %v", id, err)
		}
	}
	if _, err := NewTenantID(testOtherTenantID); err != nil {
		t.Fatal(err)
	}
}
