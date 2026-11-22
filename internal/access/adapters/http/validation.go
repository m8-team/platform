package http

import (
	"context"
	"fmt"

	"buf.build/go/protovalidate"
	accesspb "github.com/m8-team/platform/internal/proto/m8/platform/access/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// validateBatchRequest applies item-level inheritance and ensures every
// effective subject/action/resource is valid, including an empty batch.
func validateBatchRequest(req *accesspb.EvaluateAccessBatchRequest) error {
	if req == nil {
		return fmt.Errorf("batch request is required")
	}
	if err := protovalidate.Validate(req); err != nil {
		return err
	}
	if len(req.Evaluations) == 0 {
		return protovalidate.Validate(&accesspb.EvaluateAccessRequest{
			Subject: req.Subject, Action: req.Action, Resource: req.Resource, Context: req.Context,
		})
	}
	for i, evaluation := range req.Evaluations {
		if evaluation == nil {
			return fmt.Errorf("evaluation %d is required", i)
		}
		subject, action, resource := req.Subject, req.Action, req.Resource
		evalContext := req.Context
		if evaluation.Subject != nil {
			subject = evaluation.Subject
		}
		if evaluation.Action != nil {
			action = evaluation.Action
		}
		if evaluation.Resource != nil {
			resource = evaluation.Resource
		}
		if evaluation.Context != nil {
			evalContext = evaluation.Context
		}
		if err := protovalidate.Validate(&accesspb.EvaluateAccessRequest{
			Subject: subject, Action: action, Resource: resource, Context: evalContext,
		}); err != nil {
			return fmt.Errorf("evaluation %d: %w", i, err)
		}
	}
	return nil
}

func (s validatedServer) EvaluateAccess(ctx context.Context, req *accesspb.EvaluateAccessRequest) (*accesspb.EvaluateAccessResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return s.AccessServiceServer.EvaluateAccess(ctx, req)
}

func (s validatedServer) SearchResources(ctx context.Context, req *accesspb.SearchResourcesRequest) (*accesspb.SearchResourcesResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return s.AccessServiceServer.SearchResources(ctx, req)
}

func (s validatedServer) SearchSubjects(ctx context.Context, req *accesspb.SearchSubjectsRequest) (*accesspb.SearchSubjectsResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return s.AccessServiceServer.SearchSubjects(ctx, req)
}

func (s validatedServer) SearchActions(ctx context.Context, req *accesspb.SearchActionsRequest) (*accesspb.SearchActionsResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return s.AccessServiceServer.SearchActions(ctx, req)
}
