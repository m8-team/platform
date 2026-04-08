package grpcpresenter

import (
	resourcemanagerv1 "m8/platform/resourcemanager/v1"

	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundaries"
)

type OrganizationPresenter struct{}

func (OrganizationPresenter) PresentGet(output boundaries.GetOrganizationOutput) *resourcemanagerv1.Organization {
	return mapOrganization(output.Organization)
}

func (OrganizationPresenter) PresentCreate(output boundaries.CreateOrganizationOutput) *resourcemanagerv1.Organization {
	return mapOrganization(output.Organization)
}

func (OrganizationPresenter) PresentUpdate(output boundaries.UpdateOrganizationOutput) *resourcemanagerv1.Organization {
	return mapOrganization(output.Organization)
}

func (OrganizationPresenter) PresentUndelete(output boundaries.UndeleteOrganizationOutput) *resourcemanagerv1.Organization {
	return mapOrganization(output.Organization)
}

func (OrganizationPresenter) PresentList(output boundaries.ListOrganizationsOutput) *resourcemanagerv1.ListOrganizationsResponse {
	items := make([]*resourcemanagerv1.Organization, 0, len(output.Organizations))
	for _, item := range output.Organizations {
		items = append(items, mapOrganization(item))
	}
	return &resourcemanagerv1.ListOrganizationsResponse{
		Organizations: items,
		NextPageToken: output.NextPageToken,
		TotalSize:     output.TotalSize,
	}
}
