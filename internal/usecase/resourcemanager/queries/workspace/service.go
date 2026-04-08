package workspaceqry

import (
	"context"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundaries"
)

type QueryService struct {
	GetHandler  GetInteractor
	ListHandler ListInteractor
}

func (s QueryService) GetWorkspace(ctx context.Context, input boundaries.GetWorkspaceInput) (boundaries.GetWorkspaceOutput, error) {
	return s.GetHandler.Execute(ctx, input)
}

func (s QueryService) ListWorkspaces(ctx context.Context, input boundaries.ListWorkspacesInput) (boundaries.ListWorkspacesOutput, error) {
	return s.ListHandler.Execute(ctx, input)
}
