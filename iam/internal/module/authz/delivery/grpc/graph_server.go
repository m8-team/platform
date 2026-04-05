package grpc

import (
	"context"

	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
	graphv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/graph/v1"
	"github.com/m8platform/platform/iam/internal/core"
	authzentity "github.com/m8platform/platform/iam/internal/module/authz/entity"
)

type GraphServer struct {
	graphv1.UnimplementedGraphServiceServer

	store core.DocumentStore
}

func NewGraphServer(store core.DocumentStore) *GraphServer {
	return &GraphServer{store: store}
}

func (s *GraphServer) ListSubjectAccessBindings(ctx context.Context, req *graphv1.ListSubjectAccessBindingsRequest) (*graphv1.ListSubjectAccessBindingsResponse, error) {
	bindings, err := ListBindingsForSubject(ctx, s.store, req.GetSubject())
	if err != nil {
		return nil, err
	}
	edges := make([]*graphv1.ExplainEdge, 0, len(bindings))
	for _, binding := range bindings {
		edges = append(edges, &graphv1.ExplainEdge{
			EdgeId:      binding.GetBindingId(),
			FromSubject: binding.GetSubject(),
			ToResource:  binding.GetResource(),
			Relation:    binding.GetRoleId(),
			Source:      "binding_operations",
			ExpiresAt:   binding.GetExpiresAt(),
		})
	}
	return &graphv1.ListSubjectAccessBindingsResponse{Bindings: bindings, ExplainEdges: edges}, nil
}

func (s *GraphServer) ListResourceSubjects(ctx context.Context, req *graphv1.ListResourceSubjectsRequest) (*graphv1.ListResourceSubjectsResponse, error) {
	bindings, err := ListBindingsForResource(ctx, s.store, req.GetResource())
	if err != nil {
		return nil, err
	}
	subjects := make([]*authzv1.SubjectRef, 0, len(bindings))
	for _, binding := range bindings {
		subjects = append(subjects, binding.GetSubject())
	}
	return &graphv1.ListResourceSubjectsResponse{Subjects: subjects, Bindings: bindings}, nil
}

func (s *GraphServer) SimulateChangeImpact(_ context.Context, req *graphv1.SimulateChangeImpactRequest) (*graphv1.SimulateChangeImpactResponse, error) {
	impacts := make([]*graphv1.ChangeImpact, 0, len(req.GetDelta().GetMutations()))
	for _, mutation := range req.GetDelta().GetMutations() {
		impact := &graphv1.ChangeImpact{Subject: mutation.GetBinding().GetSubject()}
		switch mutation.GetKind() {
		case authzv1.BindingMutationKind_BINDING_MUTATION_KIND_ADD:
			impact.AddedPermissions = authzentity.PermissionsForRole(mutation.GetBinding().GetRoleId())
		case authzv1.BindingMutationKind_BINDING_MUTATION_KIND_REMOVE:
			impact.RemovedPermissions = authzentity.PermissionsForRole(mutation.GetBinding().GetRoleId())
		}
		impacts = append(impacts, impact)
	}
	return &graphv1.SimulateChangeImpactResponse{Impacts: impacts}, nil
}
