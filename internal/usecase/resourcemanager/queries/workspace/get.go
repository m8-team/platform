package workspaceqry

import (
	"context"
	"fmt"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundaries"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/ports"
)

type GetInteractor struct {
	Repository ports.WorkspaceRepository
}

func (i GetInteractor) Execute(ctx context.Context, input boundaries.GetWorkspaceInput) (boundaries.GetWorkspaceOutput, error) {
	entity, err := i.Repository.GetByID(ctx, input.ID, true)
	if err != nil {
		return boundaries.GetWorkspaceOutput{}, fmt.Errorf("get workspace: %w", err)
	}
	return boundaries.GetWorkspaceOutput{Workspace: usecasecommon.WorkspaceToBoundary(entity)}, nil
}
