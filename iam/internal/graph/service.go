package graph

import (
	"context"

	authzsvc "github.com/m8platform/platform/iam/internal/authz"

	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
	graphv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/graph/v1"
	"github.com/m8platform/platform/iam/internal/core"
)

type Service struct {
	graphv1.UnimplementedGraphServiceServer

	store core.DocumentStore
}

func NewService(store core.DocumentStore) *Service {
	return &Service{store: store}
}

func (s *Service) ListSubjectAccessBindings(ctx context.Context, req *graphv1.ListSubjectAccessBindingsRequest) (*graphv1.ListSubjectAccessBindingsResponse, error) {
	bindings, err := authzsvc.ListBindingsForSubject(ctx, s.store, req.GetSubject())
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

func (s *Service) ListResourceSubjects(ctx context.Context, req *graphv1.ListResourceSubjectsRequest) (*graphv1.ListResourceSubjectsResponse, error) {
	bindings, err := authzsvc.ListBindingsForResource(ctx, s.store, req.GetResource())
	if err != nil {
		return nil, err
	}
	subjects := make([]*authzv1.SubjectRef, 0, len(bindings))
	for _, binding := range bindings {
		subjects = append(subjects, binding.GetSubject())
	}
	return &graphv1.ListResourceSubjectsResponse{Subjects: subjects, Bindings: bindings}, nil
}

func (s *Service) SimulateChangeImpact(_ context.Context, req *graphv1.SimulateChangeImpactRequest) (*graphv1.SimulateChangeImpactResponse, error) {
	impacts := make([]*graphv1.ChangeImpact, 0, len(req.GetDelta().GetMutations()))
	for _, mutation := range req.GetDelta().GetMutations() {
		impact := &graphv1.ChangeImpact{
			Subject: mutation.GetBinding().GetSubject(),
		}
		switch mutation.GetKind() {
		case authzv1.BindingMutationKind_BINDING_MUTATION_KIND_ADD:
			impact.AddedPermissions = authzsvc.PermissionsForRole(mutation.GetBinding().GetRoleId())
		case authzv1.BindingMutationKind_BINDING_MUTATION_KIND_REMOVE:
			impact.RemovedPermissions = authzsvc.PermissionsForRole(mutation.GetBinding().GetRoleId())
		}
		impacts = append(impacts, impact)
	}
	return &graphv1.SimulateChangeImpactResponse{Impacts: impacts}, nil
}
