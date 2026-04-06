package modulekit

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type GRPCService struct {
	Name            string
	Register        func(grpc.ServiceRegistrar)
	RegisterGateway func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error
}

type GRPCRegistrar interface {
	RegisterGRPCService(service GRPCService)
}

type GRPCRegistry struct {
	services []GRPCService
}

func NewGRPCRegistry() *GRPCRegistry {
	return &GRPCRegistry{services: make([]GRPCService, 0)}
}

func (r *GRPCRegistry) RegisterGRPCService(service GRPCService) {
	if r == nil {
		return
	}
	r.services = append(r.services, service)
}

func (r *GRPCRegistry) Services() []GRPCService {
	if r == nil {
		return nil
	}
	services := make([]GRPCService, len(r.services))
	copy(services, r.services)
	return services
}
