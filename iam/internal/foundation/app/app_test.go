package app

import (
	"context"
	"errors"
	"testing"
)

type stubServer struct {
	serveCalls    int
	shutdownCalls int
	serveErr      error
}

func (s *stubServer) Serve() error {
	s.serveCalls++
	return s.serveErr
}

func (s *stubServer) Shutdown(_ context.Context) error {
	s.shutdownCalls++
	return nil
}

func TestApplicationCloseJoinsErrors(t *testing.T) {
	app := New(nil,
		func(context.Context) error { return errors.New("close one") },
		func(context.Context) error { return errors.New("close two") },
	)

	err := app.Close(context.Background())
	if err == nil {
		t.Fatal("expected close error")
	}
	if !stringsContain(err.Error(), "close one") || !stringsContain(err.Error(), "close two") {
		t.Fatalf("unexpected close error: %v", err)
	}
}

func TestApplicationServeDelegatesToServer(t *testing.T) {
	server := &stubServer{}
	app := New(server)

	if err := app.Serve(context.Background()); err != nil {
		t.Fatalf("serve returned error: %v", err)
	}
	if server.serveCalls != 1 {
		t.Fatalf("serve calls = %d, want 1", server.serveCalls)
	}
}

func stringsContain(value string, want string) bool {
	return len(value) >= len(want) && (value == want || contains(value, want))
}

func contains(value string, want string) bool {
	for i := 0; i+len(want) <= len(value); i++ {
		if value[i:i+len(want)] == want {
			return true
		}
	}
	return false
}
