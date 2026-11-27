package shared

import "testing"

const testTenant = "11111111-1111-4111-8111-111111111111"

func TestTenantAndRevision(t *testing.T) {
	tenant, err := NewTenantID(testTenant)
	if err != nil || tenant.String() != testTenant {
		t.Fatalf("tenant: %v", err)
	}
	if _, err := NewTenantID("not-a-tenant"); err == nil {
		t.Fatal("invalid tenant accepted")
	}
	r, err := NewRevision(1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Next(0); err != ErrRevisionConflict {
		t.Fatalf("expected CAS conflict, got %v", err)
	}
	next, err := r.Next(1)
	if err != nil || next != 2 || r != 1 {
		t.Fatalf("invalid immutable revision: %v", err)
	}
}

func TestCanonicalTenantID(t *testing.T) {
	const nilUUID = "00000000-0000-0000-0000-000000000000"
	id, err := NewTenantID(" " + nilUUID + " ")
	if err != nil || id.String() != nilUUID || id.Validate() != nil {
		t.Fatalf("explicit nil UUID: %v, %v", id, err)
	}
	for _, raw := range []string{"", "AAAAAAAA-AAAA-4AAA-8AAA-AAAAAAAAAAAA", "{" + testTenant + "}", "urn:uuid:" + testTenant} {
		id, err := NewTenantID(raw)
		if err != ErrInvalidTenantID || id.Validate() != ErrInvalidTenantID || ValidUUID(raw) {
			t.Errorf("expected invalid UUID for %q", raw)
		}
	}
}
