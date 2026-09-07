package health

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
)

func TestProbeSemantics(t *testing.T) {
	h := New()
	if err := h.Register("kafka", func(context.Context) error { return errors.New("offline") }); err != nil {
		t.Fatal(err)
	}
	status := func(path string) int {
		w := httptest.NewRecorder()
		h.Handler().ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		return w.Code
	}
	if status("/health/live") != 200 || status("/health/startup") != 503 {
		t.Fatal("bad startup probes")
	}
	h.Started()
	if status("/health/live") != 200 || status("/health/startup") != 200 || status("/health/ready") != 503 {
		t.Fatal("dependency failure affected liveness")
	}
	h.Drain()
	if h.Ready(context.Background()) {
		t.Fatal("ready while draining")
	}
}
