package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestRedaction(t *testing.T) {
	var output bytes.Buffer
	l := New(&output, Info{Name: "test"}, Config{})
	l.Info("request", "authorization", "Bearer hidden", "nested", map[string]any{"password": "hidden", "name": "public"})
	if strings.Contains(output.String(), "hidden") || !strings.Contains(output.String(), "REDACTED") {
		t.Fatal(output.String())
	}
}
