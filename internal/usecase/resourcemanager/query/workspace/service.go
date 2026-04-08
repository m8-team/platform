package workspacequery

import (
	"context"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundary"
)

type QueryService struct {
	GetHandler  GetInteractor
	ListHandler ListInteractor
}

func (s QueryService) GetWorkspace(ctx context.Context, input boundary.GetWorkspaceInput) (boundary.GetWorkspaceOutput, error) {
	return s.GetHandler.Execute(ctx, input)
}

func (s QueryService) ListWorkspaces(ctx context.Context, input boundary.ListWorkspacesInput) (boundary.ListWorkspacesOutput, error) {
	return s.ListHandler.Execute(ctx, input)
}
