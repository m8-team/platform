package workspacecommand

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

func (s CommandService) CreateWorkspace(ctx context.Context, input boundary.CreateWorkspaceInput) (boundary.CreateWorkspaceOutput, error) {
	return s.CreateHandler.Execute(ctx, input)
}

func (s CommandService) UpdateWorkspace(ctx context.Context, input boundary.UpdateWorkspaceInput) (boundary.UpdateWorkspaceOutput, error) {
	return s.UpdateHandler.Execute(ctx, input)
}

func (s CommandService) DeleteWorkspace(ctx context.Context, input boundary.DeleteWorkspaceInput) (boundary.DeleteWorkspaceOutput, error) {
	return s.DeleteHandler.Execute(ctx, input)
}

func (s CommandService) UndeleteWorkspace(ctx context.Context, input boundary.UndeleteWorkspaceInput) (boundary.UndeleteWorkspaceOutput, error) {
	return s.UndeleteHandler.Execute(ctx, input)
}
