package projectcmd

import (
	"context"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundaries"
)

type CommandService struct {
	CreateHandler   CreateInteractor
	UpdateHandler   UpdateInteractor
	DeleteHandler   DeleteInteractor
	UndeleteHandler UndeleteInteractor
}

func (s CommandService) CreateProject(ctx context.Context, input boundaries.CreateProjectInput) (boundaries.CreateProjectOutput, error) {
	return s.CreateHandler.Execute(ctx, input)
}

func (s CommandService) UpdateProject(ctx context.Context, input boundaries.UpdateProjectInput) (boundaries.UpdateProjectOutput, error) {
	return s.UpdateHandler.Execute(ctx, input)
}

func (s CommandService) DeleteProject(ctx context.Context, input boundaries.DeleteProjectInput) (boundaries.DeleteProjectOutput, error) {
	return s.DeleteHandler.Execute(ctx, input)
}

func (s CommandService) UndeleteProject(ctx context.Context, input boundaries.UndeleteProjectInput) (boundaries.UndeleteProjectOutput, error) {
	return s.UndeleteHandler.Execute(ctx, input)
}
