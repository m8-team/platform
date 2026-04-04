package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
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
	"google.golang.org/grpc/credentials/insecure"
	health "google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
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
	grpcServer   *grpc.Server
	grpcListener net.Listener
	httpServer   *http.Server
	httpListener net.Listener
	gatewayConn  *grpc.ClientConn
	logger       *zap.Logger
}

func New(grpcCfg config.GRPCConfig, httpCfg config.HTTPConfig, logger *zap.Logger, validator core.Validator, services Services) (*Server, error) {
	grpcListener, err := net.Listen("tcp", grpcCfg.Address)
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

	server := &Server{
		grpcServer:   grpcServer,
		grpcListener: grpcListener,
		logger:       logger,
	}
	if err := server.initHTTPGateway(httpCfg); err != nil {
		_ = grpcListener.Close()
		return nil, err
	}
	return server, nil
}

func (s *Server) Serve() error {
	if s == nil {
		return errors.New("grpc server is nil")
	}

	expected := 1
	errCh := make(chan error, 2)

	s.logger.Info("starting grpc server", zap.String("address", s.grpcListener.Addr().String()))
	go func() {
		errCh <- normalizeServeError(s.grpcServer.Serve(s.grpcListener))
	}()

	if s.httpServer != nil && s.httpListener != nil {
		expected++
		s.logger.Info("starting http gateway", zap.String("address", s.httpListener.Addr().String()))
		go func() {
			errCh <- normalizeServeError(s.httpServer.Serve(s.httpListener))
		}()
	}

	for i := 0; i < expected; i++ {
		if err := <-errCh; err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s == nil {
		return nil
	}

	var shutdownErr error
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			shutdownErr = errors.Join(shutdownErr, err)
		}
	}
	if s.gatewayConn != nil {
		if err := s.gatewayConn.Close(); err != nil {
			shutdownErr = errors.Join(shutdownErr, err)
		}
	}
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
	if s.httpListener != nil {
		if err := s.httpListener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			shutdownErr = errors.Join(shutdownErr, err)
		}
	}
	if s.grpcListener != nil {
		if err := s.grpcListener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			shutdownErr = errors.Join(shutdownErr, err)
		}
	}
	return shutdownErr
}

func (s *Server) initHTTPGateway(cfg config.HTTPConfig) error {
	if strings.TrimSpace(cfg.Address) == "" {
		return nil
	}

	httpListener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}

	gatewayMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)

	endpoint := localDialAddress(s.grpcListener.Addr())
	dialOptions := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	gatewayConn, err := grpc.NewClient(endpoint, dialOptions...)
	if err != nil {
		_ = httpListener.Close()
		return err
	}
	if err := registerGatewayHandlers(context.Background(), gatewayMux, gatewayConn); err != nil {
		_ = gatewayConn.Close()
		_ = httpListener.Close()
		return err
	}

	rootMux := http.NewServeMux()
	rootMux.Handle("/openapi/", openAPIHandler(cfg.OpenAPIDir))
	rootMux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	})
	rootMux.Handle("/", gatewayMux)

	s.gatewayConn = gatewayConn
	s.httpListener = httpListener
	s.httpServer = &http.Server{
		Handler: rootMux,
	}
	return nil
}

func registerGatewayHandlers(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	if err := identityv1.RegisterIdentityServiceHandler(ctx, mux, conn); err != nil {
		return fmt.Errorf("register identity gateway: %w", err)
	}
	if err := identityv1.RegisterOAuthFacadeServiceHandler(ctx, mux, conn); err != nil {
		return fmt.Errorf("register oauth gateway: %w", err)
	}
	if err := authzv1.RegisterAuthorizationFacadeServiceHandler(ctx, mux, conn); err != nil {
		return fmt.Errorf("register authz gateway: %w", err)
	}
	if err := graphv1.RegisterGraphServiceHandler(ctx, mux, conn); err != nil {
		return fmt.Errorf("register graph gateway: %w", err)
	}
	if err := supportv1.RegisterSupportAccessServiceHandler(ctx, mux, conn); err != nil {
		return fmt.Errorf("register support gateway: %w", err)
	}
	if err := auditv1.RegisterAuditServiceHandler(ctx, mux, conn); err != nil {
		return fmt.Errorf("register audit gateway: %w", err)
	}
	if err := opsv1.RegisterOperationsServiceHandler(ctx, mux, conn); err != nil {
		return fmt.Errorf("register ops gateway: %w", err)
	}
	return nil
}

func openAPIHandler(openAPIDir string) http.Handler {
	if strings.TrimSpace(openAPIDir) == "" {
		return http.NotFoundHandler()
	}
	cleanPath := filepath.Clean(openAPIDir)
	if _, err := os.Stat(cleanPath); err != nil {
		return http.NotFoundHandler()
	}
	return http.StripPrefix("/openapi/", http.FileServer(http.Dir(cleanPath)))
}

func localDialAddress(addr net.Addr) string {
	tcpAddr, ok := addr.(*net.TCPAddr)
	if ok {
		return net.JoinHostPort("127.0.0.1", strconv.Itoa(tcpAddr.Port))
	}
	host, port, err := net.SplitHostPort(addr.String())
	if err != nil {
		return addr.String()
	}
	switch host {
	case "", "0.0.0.0", "::", "[::]":
		host = "127.0.0.1"
	}
	return net.JoinHostPort(host, port)
}

func normalizeServeError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, http.ErrServerClosed):
		return nil
	case errors.Is(err, grpc.ErrServerStopped):
		return nil
	case errors.Is(err, net.ErrClosed):
		return nil
	default:
		return err
	}
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
