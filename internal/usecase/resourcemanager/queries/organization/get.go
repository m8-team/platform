package organizationqry

import (
	"context"
	"fmt"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundaries"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/ports"
)

type GetInteractor struct {
	Repository ports.OrganizationRepository
}

func (i GetInteractor) Execute(ctx context.Context, input boundaries.GetOrganizationInput) (boundaries.GetOrganizationOutput, error) {
	entity, err := i.Repository.GetByID(ctx, input.ID, true)
	if err != nil {
		return boundaries.GetOrganizationOutput{}, fmt.Errorf("get organization: %w", err)
	}
	return boundaries.GetOrganizationOutput{Organization: usecasecommon.OrganizationToBoundary(entity)}, nil
}
