package idempotency

import "testing"

func TestKeyStable(t *testing.T) {
	left := Key("create_user", "request-1")
	right := Key("create_user", "request-1")
	if left != right {
		t.Fatalf("expected stable idempotency key, got %q and %q", left, right)
	}
	if left == Key("create_user", "request-2") {
		t.Fatal("expected different request ids to produce different keys")
	}
}
