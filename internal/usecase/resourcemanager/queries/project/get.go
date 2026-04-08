package projectqry

import (
	"context"
	"fmt"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundaries"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/ports"
)

type GetInteractor struct {
	Repository ports.ProjectRepository
}

func (i GetInteractor) Execute(ctx context.Context, input boundaries.GetProjectInput) (boundaries.GetProjectOutput, error) {
	entity, err := i.Repository.GetByID(ctx, input.ID, true)
	if err != nil {
		return boundaries.GetProjectOutput{}, fmt.Errorf("get project: %w", err)
	}
	return boundaries.GetProjectOutput{Project: usecasecommon.ProjectToBoundary(entity)}, nil
}
