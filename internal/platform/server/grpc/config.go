package grpcserver

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

var (
	ErrAddressRequired = errors.New("gRPC server address is required")
	ErrInvalidAddress  = errors.New("invalid gRPC server address")
)

// Config contains process-level configuration for a gRPC listener.
// Configuration is supplied by the composition root; this package does not
// read environment variables itself.
type Config struct {
	Address string
}

func (c Config) Validate() error {
	address := strings.TrimSpace(c.Address)
	if address == "" {
		return ErrAddressRequired
	}

	if _, err := net.ResolveTCPAddr("tcp", address); err != nil {
		return fmt.Errorf("%w %q: %v", ErrInvalidAddress, address, err)
	}

	return nil
}

func (c Config) normalized() Config {
	c.Address = strings.TrimSpace(c.Address)
	return c
}
