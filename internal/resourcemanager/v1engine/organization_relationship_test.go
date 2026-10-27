package v1engine

import (
	"errors"
	"testing"
	"time"
)

func TestTemporalOwnershipGraph(t *testing.T) {
	newRel := func(from, to string, start, end int64) OrganizationRelationship {
		return OrganizationRelationship{
			TenantID: "tenant-a", FromOrganizationID: from, ToOrganizationID: to,
			Type: RelationshipParentSubsidiary, ValidFrom: start, ValidTo: end,
		}
	}
	existing := []OrganizationRelationship{
		newRel("a", "b", 100, 200),
		newRel("b", "c", 150, 250),
	}
	if err := ValidateRelationshipGraph(existing, newRel("c", "a", 170, 190)); !errors.Is(err, ErrRelationshipCycle) {
		t.Fatalf("overlapping indirect cycle must fail: %v", err)
	}
	if err := ValidateRelationshipGraph(existing, newRel("c", "a", 210, 240)); err != nil {
		t.Fatalf("non-overlapping graph must be valid: %v", err)
	}
	if err := ValidateRelationshipGraph(existing, newRel("a", "b", 180, 220)); !errors.Is(err, ErrRelationshipOverlap) {
		t.Fatalf("duplicate overlapping type must fail: %v", err)
	}
	if err := ValidateRelationshipGraph(existing, newRel("a", "b", 200, 220)); err != nil {
		t.Fatalf("adjacent validity periods must be allowed: %v", err)
	}
	if err := ValidateRelationshipGraph(existing, newRel("x", "x", 100, 0)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("self-parent must fail: %v", err)
	}
}

func TestRelationshipOwnershipRules(t *testing.T) {
	r := OrganizationRelationship{
		TenantID: "tenant", FromOrganizationID: "a", ToOrganizationID: "b",
		Type: RelationshipPartner, ValidFrom: 1234, OwnershipBasisPoints: 10,
	}
	if err := r.Validate(); !errors.Is(err, ErrInvalid) {
		t.Fatalf("partner stake cannot be implied: %v", err)
	}
	r.Type = RelationshipParentSubsidiary
	r.OwnershipBasisPoints = 10001
	if err := r.Validate(); !errors.Is(err, ErrInvalid) {
		t.Fatalf("stake > 100%%: %v", err)
	}
	r.OwnershipBasisPoints = 5100
	if err := r.Validate(); err != nil {
		t.Fatalf("valid stake: %v", err)
	}
}

func TestRelationshipTenantIsolationAndRevision(t *testing.T) {
	s := NewStore()
	create := func(kind Kind, name string, parent Owner) Record {
		t.Helper()
		r, err := s.Create(Record{Kind: kind, Name: name, Owner: parent}, name)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	tenantA := create(Tenant, "Tenant A", Owner{})
	tenantB := create(Tenant, "Tenant B", Owner{})
	a := create(Organization, "Parent", Owner{Kind: Tenant, ID: tenantA.ID})
	b := create(Organization, "Child", Owner{Kind: Tenant, ID: tenantA.ID})
	other := create(Organization, "Outside", Owner{Kind: Tenant, ID: tenantB.ID})
	r := OrganizationRelationship{TenantID: tenantA.ID, FromOrganizationID: a.ID, ToOrganizationID: b.ID,
		Type: RelationshipParentSubsidiary, OwnershipBasisPoints: 5100, ValidFrom: time.Now().Add(-time.Hour).Unix()}
	result, err := s.CreateRelationship(r, "create-1")
	if err != nil {
		t.Fatal(err)
	}
	replay, err := s.CreateRelationship(r, "create-1")
	if err != nil || replay.ID != result.ID {
		t.Fatalf("idempotency replay: %#v %v", replay, err)
	}
	if _, err = s.CreateRelationship(OrganizationRelationship{TenantID: tenantA.ID,
		FromOrganizationID: a.ID, ToOrganizationID: other.ID, Type: RelationshipPartner, ValidFrom: r.ValidFrom}, "create-2"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("cross-tenant association: %v", err)
	}
	if _, err = s.GetRelationship(tenantB.ID, result.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-tenant read: %v", err)
	}
	if list, err := s.ListRelationships(tenantB.ID, "", true); err != nil || len(list) != 0 {
		t.Fatalf("tenant list isolation: %#v %v", list, err)
	}
	u := RelationshipUpdate{TenantID: tenantA.ID, ID: result.ID, ExpectedRevision: 1, OwnershipBasisPoints: 5500}
	updated, err := s.UpdateRelationship(u, "update-1")
	if err != nil || updated.Revision != 2 {
		t.Fatalf("update: %#v %v", updated, err)
	}
	if _, err = s.UpdateRelationship(u, "update-2"); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale revision: %v", err)
	}
	now := time.Now().Unix()
	closed, err := s.CloseRelationship(tenantA.ID, result.ID, 2, now, "close-1")
	if err != nil || closed.State != "CLOSED" || closed.Revision != 3 {
		t.Fatalf("close: %#v %v", closed, err)
	}
	closedReplay, err := s.CloseRelationship(tenantA.ID, result.ID, 2, now, "close-1")
	if err != nil || closedReplay.ID != result.ID {
		t.Fatalf("close replay: %#v %v", closedReplay, err)
	}
	if result, err := s.ListRelationships(tenantA.ID, "", false); err != nil || len(result) != 0 {
		t.Fatalf("closed hidden: %#v %v", result, err)
	}
	if result, err := s.ListRelationships(tenantA.ID, "", true); err != nil || len(result) != 1 {
		t.Fatalf("closed history: %#v %v", result, err)
	}
}
