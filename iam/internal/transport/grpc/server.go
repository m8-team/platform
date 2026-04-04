package grpc

import (
	"context"
	"errors"
	"net"

	auditv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/audit/v1"
	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
	graphv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/graph/v1"
	identityv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/identity/v1"
	opsv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/ops/v1"
	supportv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/support/v1"
	"github.com/m8platform/platform/iam/internal/config"
	"github.com/m8platform/platform/iam/internal/core"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	health "google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type Services struct {
	Identity identityv1.IdentityServiceServer
	OAuth    identityv1.OAuthFacadeServiceServer
	Authz    authzv1.AuthorizationFacadeServiceServer
	Graph    graphv1.GraphServiceServer
	Support  supportv1.SupportAccessServiceServer
	Audit    auditv1.AuditServiceServer
	Ops      opsv1.OperationsServiceServer
}

type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
	logger     *zap.Logger
}

func New(cfg config.GRPCConfig, logger *zap.Logger, validator core.Validator, services Services) (*Server, error) {
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(validationInterceptor(validator)),
	)

	identityv1.RegisterIdentityServiceServer(grpcServer, services.Identity)
	identityv1.RegisterOAuthFacadeServiceServer(grpcServer, services.OAuth)
	authzv1.RegisterAuthorizationFacadeServiceServer(grpcServer, services.Authz)
	graphv1.RegisterGraphServiceServer(grpcServer, services.Graph)
	supportv1.RegisterSupportAccessServiceServer(grpcServer, services.Support)
	auditv1.RegisterAuditServiceServer(grpcServer, services.Audit)
	opsv1.RegisterOperationsServiceServer(grpcServer, services.Ops)

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthv1.HealthCheckResponse_SERVING)
	healthv1.RegisterHealthServer(grpcServer, healthServer)

	return &Server{
		grpcServer: grpcServer,
		listener:   listener,
		logger:     logger,
	}, nil
}

func (s *Server) Serve() error {
	if s == nil {
		return errors.New("grpc server is nil")
	}
	s.logger.Info("starting grpc server", zap.String("address", s.listener.Addr().String()))
	return s.grpcServer.Serve(s.listener)
}

func (s *Server) Shutdown(context.Context) error {
	if s == nil {
		return nil
	}
	s.grpcServer.GracefulStop()
	return nil
}

func validationInterceptor(validator core.Validator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if validator != nil {
			if message, ok := req.(proto.Message); ok {
				if err := validator.Validate(message); err != nil {
					return nil, status.Error(codes.InvalidArgument, info.FullMethod+": "+err.Error())
				}
			}
		}
		return handler(ctx, req)
	}
}
