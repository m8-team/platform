package core

import "testing"

func TestIdempotencyKeyStable(t *testing.T) {
	left := IdempotencyKey("create_user", "request-1")
	right := IdempotencyKey("create_user", "request-1")
	if left != right {
		t.Fatalf("expected stable idempotency key, got %q and %q", left, right)
	}
	if left == IdempotencyKey("create_user", "request-2") {
		t.Fatal("expected different request ids to produce different keys")
	}
}
