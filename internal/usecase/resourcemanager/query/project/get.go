package projectquery

import (
	"context"
	"fmt"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundary"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/port"
)

type GetInteractor struct {
	Repository port.ProjectRepository
}

func (i GetInteractor) Execute(ctx context.Context, input boundary.GetProjectInput) (boundary.GetProjectOutput, error) {
	entity, err := i.Repository.GetByID(ctx, input.ID, true)
	if err != nil {
		return boundary.GetProjectOutput{}, fmt.Errorf("get project: %w", err)
	}
	return boundary.GetProjectOutput{Project: usecasecommon.ProjectToBoundary(entity)}, nil
}
