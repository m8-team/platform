package grpc

import legacyaudit "github.com/m8platform/platform/iam/internal/audit"

type Server struct {
	*legacyaudit.Service
}

func NewServer(service *legacyaudit.Service) *Server {
	return &Server{Service: service}
}
