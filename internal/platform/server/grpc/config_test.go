package grpcserver

import (
	"errors"
	"testing"
)

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		address string
		wantErr error
	}{
		{name: "IPv4", address: "127.0.0.1:9090"},
		{name: "any host ephemeral port", address: ":0"},
		{name: "IPv6", address: "[::1]:9090"},
		{name: "empty", wantErr: ErrAddressRequired},
		{name: "whitespace", address: "  ", wantErr: ErrAddressRequired},
		{name: "missing port", address: "127.0.0.1", wantErr: ErrInvalidAddress},
		{name: "invalid port", address: "127.0.0.1:not-a-port", wantErr: ErrInvalidAddress},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := (Config{Address: tt.address}).Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate() error = %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
