package workspacecmd

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

func (s CommandService) CreateWorkspace(ctx context.Context, input boundaries.CreateWorkspaceInput) (boundaries.CreateWorkspaceOutput, error) {
	return s.CreateHandler.Execute(ctx, input)
}

func (s CommandService) UpdateWorkspace(ctx context.Context, input boundaries.UpdateWorkspaceInput) (boundaries.UpdateWorkspaceOutput, error) {
	return s.UpdateHandler.Execute(ctx, input)
}

func (s CommandService) DeleteWorkspace(ctx context.Context, input boundaries.DeleteWorkspaceInput) (boundaries.DeleteWorkspaceOutput, error) {
	return s.DeleteHandler.Execute(ctx, input)
}

func (s CommandService) UndeleteWorkspace(ctx context.Context, input boundaries.UndeleteWorkspaceInput) (boundaries.UndeleteWorkspaceOutput, error) {
	return s.UndeleteHandler.Execute(ctx, input)
}
