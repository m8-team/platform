package grpcserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"google.golang.org/grpc"
)

var (
	ErrServerRequired       = errors.New("gRPC server is required")
	ErrServerAlreadyStarted = errors.New("gRPC server is already started")
)

// Server owns the listener lifecycle for a grpc.Server. The underlying
// grpc.Server is also exposed separately by the Fx module for technical
// adapters that need it.
type Server struct {
	config     Config
	grpcServer *grpc.Server

	mu        sync.RWMutex
	listener  net.Listener
	serveDone chan struct{}
	serveErr  error
}

func NewServer(config Config, grpcServer *grpc.Server) (*Server, error) {
	config = config.normalized()
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if grpcServer == nil {
		return nil, ErrServerRequired
	}

	return &Server{
		config:     config,
		grpcServer: grpcServer,
	}, nil
}

// GRPCServer returns the underlying gRPC server.
func (s *Server) GRPCServer() *grpc.Server {
	if s == nil {
		return nil
	}

	return s.grpcServer
}

// Address returns the bound listener address after Start. It returns nil
// before the server has started.
func (s *Server) Address() net.Addr {
	if s == nil {
		return nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.listener == nil {
		return nil
	}

	return s.listener.Addr()
}

func (s *Server) Start(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("start gRPC server: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener != nil {
		return ErrServerAlreadyStarted
	}

	listener, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		return fmt.Errorf("listen for gRPC on %s: %w", s.config.Address, err)
	}

	s.listener = listener
	s.serveDone = make(chan struct{})

	go s.serve(listener, s.serveDone)

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	if s == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	s.mu.RLock()
	started := s.listener != nil
	serveDone := s.serveDone
	s.mu.RUnlock()
	if !started {
		return nil
	}

	gracefulDone := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(gracefulDone)
	}()

	forced := false
	select {
	case <-gracefulDone:
	case <-ctx.Done():
		forced = true
		s.grpcServer.Stop()
		<-gracefulDone
	}

	if serveDone != nil {
		<-serveDone
	}
	if forced {
		return fmt.Errorf("gracefully stop gRPC server: %w", ctx.Err())
	}

	s.mu.RLock()
	serveErr := s.serveErr
	s.mu.RUnlock()
	if serveErr != nil {
		return fmt.Errorf("serve gRPC: %w", serveErr)
	}

	return nil
}

func (s *Server) serve(listener net.Listener, done chan struct{}) {
	err := s.grpcServer.Serve(listener)
	if errors.Is(err, grpc.ErrServerStopped) {
		err = nil
	}

	s.mu.Lock()
	s.serveErr = err
	s.mu.Unlock()

	close(done)
}
