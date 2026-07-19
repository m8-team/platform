package grpcserver

import (
	"errors"
	"fmt"

	"go.uber.org/fx"
	"google.golang.org/grpc"
)

const (
	// RegistrationGroup is the Fx value group consumed by Module.
	RegistrationGroup = "grpc-server-registrations"
	// RegistrationResultTag can be passed directly to fx.ResultTags when a
	// module contributes a gRPC service registration.
	RegistrationResultTag = `group:"grpc-server-registrations"`
)

var ErrRegistrationRequired = errors.New("gRPC service registration is required")

// Registration registers one or more gRPC services with a registrar. Business
// adapters contribute callbacks through RegistrationGroup; the platform server
// executes every callback before it starts accepting connections.
type Registration func(grpc.ServiceRegistrar) error

type registrationParams struct {
	fx.In

	Registrations []Registration `group:"grpc-server-registrations"`
}

func newGRPCServer(params registrationParams, options ...grpc.ServerOption) (*grpc.Server, error) {
	server := grpc.NewServer(options...)

	for index, register := range params.Registrations {
		if register == nil {
			return nil, fmt.Errorf("%w at index %d", ErrRegistrationRequired, index)
		}
		if err := register(server); err != nil {
			return nil, fmt.Errorf("register gRPC service at index %d: %w", index, err)
		}
	}

	return server, nil
}
