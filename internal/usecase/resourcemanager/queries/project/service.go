package projectqry

import (
	"context"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundaries"
)

type QueryService struct {
	GetHandler  GetInteractor
	ListHandler ListInteractor
}

func (s QueryService) GetProject(ctx context.Context, input boundaries.GetProjectInput) (boundaries.GetProjectOutput, error) {
	return s.GetHandler.Execute(ctx, input)
}

func (s QueryService) ListProjects(ctx context.Context, input boundaries.ListProjectsInput) (boundaries.ListProjectsOutput, error) {
	return s.ListHandler.Execute(ctx, input)
}
