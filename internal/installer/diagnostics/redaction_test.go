package diagnostics

import "testing"

func TestRedactMapRedactsSensitiveKeysAndBearerValues(t *testing.T) {
	redacted := RedactMap(map[string]string{
		"username":      "admin",
		"password":      "secret",
		"Authorization": "Bearer abc",
	})

	if redacted["username"] != "admin" {
		t.Fatalf("username redacted unexpectedly: %q", redacted["username"])
	}
	if redacted["password"] != "[REDACTED]" {
		t.Fatalf("password was not redacted: %q", redacted["password"])
	}
	if redacted["Authorization"] != "[REDACTED]" {
		t.Fatalf("authorization was not redacted: %q", redacted["Authorization"])
	}
}
