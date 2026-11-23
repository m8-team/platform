// Package http exposes the AuthZEN HTTP binding for an AccessService server.
package http

import (
	"context"
	"net/http"

	"buf.build/go/protovalidate"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/m8-team/platform/internal/access/adapters/internal/authzenbatch"
	accesspb "github.com/m8-team/platform/internal/proto/m8/platform/access/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func NewHandler(ctx context.Context, server accesspb.AccessServiceServer) (http.Handler, error) {
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{EmitDefaultValues: true},
		}),
		runtime.WithForwardResponseRewriter(authZENResponse),
	)
	if err := accesspb.RegisterAccessServiceHandlerServer(ctx, mux, validatedServer{server}); err != nil {
		return nil, err
	}
	return mux, nil
}

func authZENResponse(_ context.Context, response proto.Message) (any, error) {
	// A single decision on the batch endpoint has the same JSON shape as /evaluation.
	if batch, ok := response.(*accesspb.EvaluateAccessBatchResponse); ok && batch.Decision != nil && len(batch.Evaluations) == 0 {
		return &accesspb.EvaluateAccessResponse{Decision: batch.GetDecision(), Context: batch.Context}, nil
	}
	return response, nil
}

type validatedServer struct {
	accesspb.AccessServiceServer
}

func (s validatedServer) EvaluateAccessBatch(ctx context.Context, req *accesspb.EvaluateAccessBatchRequest) (*accesspb.EvaluateAccessBatchResponse, error) {
	if err := validateBatchRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if len(req.Evaluations) != 0 {
		return s.AccessServiceServer.EvaluateAccessBatch(ctx, req)
	}
	decision, err := s.EvaluateAccess(ctx, authzenbatch.Evaluation(req, nil))
	if err != nil {
		return nil, err
	}
	if decision == nil {
		return nil, status.Error(codes.Internal, "missing access decision")
	}
	return &accesspb.EvaluateAccessBatchResponse{Decision: proto.Bool(decision.Decision), Context: decision.Context}, nil
}

func (s validatedServer) GetConfiguration(ctx context.Context, req *accesspb.GetConfigurationRequest) (*accesspb.ConfigurationResponse, error) {
	metadata, err := s.AccessServiceServer.GetConfiguration(ctx, req)
	if err != nil {
		return nil, err
	}
	if metadata == nil || protovalidate.Validate(metadata) != nil {
		return nil, status.Error(codes.Internal, "invalid AuthZEN configuration")
	}
	return metadata, nil
}
