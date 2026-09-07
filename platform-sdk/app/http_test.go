package app

import (
	"context"
	"net"
	"net/http"
	"testing"
)

func TestHTTPRollbackClosesUnservedListener(t *testing.T) {
	reservation, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := reservation.Addr().String()
	if err = reservation.Close(); err != nil {
		t.Fatal(err)
	}
	component, err := HTTP("http", HTTPConfig{Address: address, InsecureLocal: true}, http.NewServeMux())
	if err != nil {
		t.Fatal(err)
	}
	if err = component.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	// Rollback before Run/Serve must release the port, not only server connections.
	if err = component.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		t.Fatal("listener leaked:", err)
	}
	listener.Close()
}
func TestHTTPRequiresTLS(t *testing.T) {
	if _, err := HTTP("http", HTTPConfig{Address: ":8080"}, http.NewServeMux()); err == nil {
		t.Fatal("insecure default")
	}
}
