package projectquery

import (
	"context"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundary"
)

type QueryService struct {
	GetHandler  GetInteractor
	ListHandler ListInteractor
}

func (s QueryService) GetProject(ctx context.Context, input boundary.GetProjectInput) (boundary.GetProjectOutput, error) {
	return s.GetHandler.Execute(ctx, input)
}

func (s QueryService) ListProjects(ctx context.Context, input boundary.ListProjectsInput) (boundary.ListProjectsOutput, error) {
	return s.ListHandler.Execute(ctx, input)
}
