// Package cataloggrpc exposes the versioned Access Catalog gRPC API.
// It has no infrastructure dependencies and never authenticates request fields.
package cataloggrpc

import (
	"context"
	"errors"

	"buf.build/go/protovalidate"
	"github.com/m8-team/platform/internal/access/app/catalog"
	"github.com/m8-team/platform/internal/access/domain/shared"
	accesspb "github.com/m8-team/platform/internal/proto/m8/platform/access/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server maps public Protobuf requests to the authenticated catalog use cases.
// No production registration is performed until a durable Registry is supplied.
type Server struct {
	accesspb.UnimplementedAccessCatalogServiceServer
	api *catalog.API
}

func New(api *catalog.API) (*Server, error) {
	if api == nil {
		return nil, catalog.ErrCatalogUnavailable
	}
	return &Server{api: api}, nil
}

func asStatus(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, catalog.ErrCatalogArgument):
		return status.Error(codes.InvalidArgument, "invalid catalog request")
	case errors.Is(err, catalog.ErrCatalogUnauthenticated):
		return status.Error(codes.Unauthenticated, "authenticated catalog identity required")
	case errors.Is(err, catalog.ErrCatalogForbidden), errors.Is(err, catalog.ErrUnauthorizedPublisher):
		return status.Error(codes.PermissionDenied, "catalog namespace access denied")
	case errors.Is(err, catalog.ErrCatalogNotFound):
		return status.Error(codes.NotFound, "catalog release not found")
	case errors.Is(err, shared.ErrRevisionConflict):
		return status.Error(codes.Aborted, "catalog revision conflict")
	case errors.Is(err, catalog.ErrCatalogConflict):
		return status.Error(codes.AlreadyExists, "catalog release or idempotency conflict")
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "request canceled")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "request deadline exceeded")
	default:
		return status.Error(codes.Unavailable, "authorization catalog unavailable")
	}
}

func (s *Server) PublishModuleManifest(ctx context.Context, req *accesspb.PublishModuleManifestRequest) (*accesspb.CatalogManifestRecord, error) {
	if s == nil || s.api == nil {
		return nil, asStatus(catalog.ErrCatalogUnavailable)
	}
	if req == nil || protovalidate.Validate(req) != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid publication request")
	}
	input, err := fromProto(req.GetManifest())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid catalog manifest")
	}
	record, err := s.api.Publish(ctx, input, req.GetExpectedRevision(), req.GetIdempotencyKey())
	if err != nil {
		return nil, asStatus(err)
	}
	return toProtoRecord(record), nil
}

func (s *Server) GetModuleManifest(ctx context.Context, req *accesspb.GetModuleManifestRequest) (*accesspb.CatalogManifestRecord, error) {
	if s == nil || s.api == nil {
		return nil, asStatus(catalog.ErrCatalogUnavailable)
	}
	if req == nil || protovalidate.Validate(req) != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid catalog lookup")
	}
	record, err := s.api.Get(ctx, req.GetModuleId(), req.GetVersion())
	if err != nil {
		return nil, asStatus(err)
	}
	return toProtoRecord(record), nil
}
func (s *Server) ListModuleManifests(ctx context.Context, req *accesspb.ListModuleManifestsRequest) (*accesspb.ListModuleManifestsResponse, error) {
	if s == nil || s.api == nil {
		return nil, asStatus(catalog.ErrCatalogUnavailable)
	}
	if req == nil || protovalidate.Validate(req) != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid catalog list")
	}
	records, next, err := s.api.List(ctx, req.GetModuleId(), int(req.GetPageSize()), req.GetPageToken())
	if err != nil {
		return nil, asStatus(err)
	}
	resp := &accesspb.ListModuleManifestsResponse{NextPageToken: next, Manifests: make([]*accesspb.CatalogManifestRecord, 0, len(records))}
	for _, record := range records {
		resp.Manifests = append(resp.Manifests, toProtoRecord(record))
	}
	return resp, nil
}

func (s *Server) ListCatalogResourceTypes(ctx context.Context, req *accesspb.ListCatalogResourceTypesRequest) (*accesspb.ListCatalogResourceTypesResponse, error) {
	if s == nil || s.api == nil {
		return nil, asStatus(catalog.ErrCatalogUnavailable)
	}
	if req == nil || protovalidate.Validate(req) != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid resource catalog query")
	}
	record, err := s.api.Get(ctx, req.GetModuleId(), req.GetVersion())
	if err != nil {
		return nil, asStatus(err)
	}
	return &accesspb.ListCatalogResourceTypesResponse{ResourceTypes: toProtoManifest(record.Manifest).Resources}, nil
}
func (s *Server) ListCatalogPermissions(ctx context.Context, req *accesspb.ListCatalogPermissionsRequest) (*accesspb.ListCatalogPermissionsResponse, error) {
	if s == nil || s.api == nil {
		return nil, asStatus(catalog.ErrCatalogUnavailable)
	}
	if req == nil || protovalidate.Validate(req) != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid permission catalog query")
	}
	record, err := s.api.Get(ctx, req.GetModuleId(), req.GetVersion())
	if err != nil {
		return nil, asStatus(err)
	}
	return &accesspb.ListCatalogPermissionsResponse{Permissions: toProtoManifest(record.Manifest).Permissions}, nil
}
func (s *Server) ListCatalogRoleTemplates(ctx context.Context, req *accesspb.ListCatalogRoleTemplatesRequest) (*accesspb.ListCatalogRoleTemplatesResponse, error) {
	if s == nil || s.api == nil {
		return nil, asStatus(catalog.ErrCatalogUnavailable)
	}
	if req == nil || protovalidate.Validate(req) != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid role template catalog query")
	}
	record, err := s.api.Get(ctx, req.GetModuleId(), req.GetVersion())
	if err != nil {
		return nil, asStatus(err)
	}
	return &accesspb.ListCatalogRoleTemplatesResponse{Roles: toProtoManifest(record.Manifest).Roles}, nil
}
