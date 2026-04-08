package projectcommand

import (
	"context"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundary"
)

type CommandService struct {
	CreateHandler   CreateInteractor
	UpdateHandler   UpdateInteractor
	DeleteHandler   DeleteInteractor
	UndeleteHandler UndeleteInteractor
}

func (s CommandService) CreateProject(ctx context.Context, input boundary.CreateProjectInput) (boundary.CreateProjectOutput, error) {
	return s.CreateHandler.Execute(ctx, input)
}

func (s CommandService) UpdateProject(ctx context.Context, input boundary.UpdateProjectInput) (boundary.UpdateProjectOutput, error) {
	return s.UpdateHandler.Execute(ctx, input)
}

func (s CommandService) DeleteProject(ctx context.Context, input boundary.DeleteProjectInput) (boundary.DeleteProjectOutput, error) {
	return s.DeleteHandler.Execute(ctx, input)
}

func (s CommandService) UndeleteProject(ctx context.Context, input boundary.UndeleteProjectInput) (boundary.UndeleteProjectOutput, error) {
	return s.UndeleteHandler.Execute(ctx, input)
}
