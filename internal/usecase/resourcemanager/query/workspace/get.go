package workspacequery

import (
	"context"
	"fmt"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundary"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/port"
)

type GetInteractor struct {
	Repository port.WorkspaceRepository
}

func (i GetInteractor) Execute(ctx context.Context, input boundary.GetWorkspaceInput) (boundary.GetWorkspaceOutput, error) {
	entity, err := i.Repository.GetByID(ctx, input.ID, true)
	if err != nil {
		return boundary.GetWorkspaceOutput{}, fmt.Errorf("get workspace: %w", err)
	}
	return boundary.GetWorkspaceOutput{Workspace: usecasecommon.WorkspaceToBoundary(entity)}, nil
}
