package grpcpresenter

import (
	resourcemanagerv1 "m8/platform/resourcemanager/v1"

	organizationboundary "github.com/m8platform/platform/internal/usecase/resourcemanager/organization/boundary"
)

type OrganizationPresenter struct{}

func (OrganizationPresenter) PresentGet(output organizationboundary.GetOrganizationOutput) *resourcemanagerv1.Organization {
	return mapOrganization(output.Organization)
}

func (OrganizationPresenter) PresentCreate(output organizationboundary.CreateOrganizationOutput) *resourcemanagerv1.Organization {
	return mapOrganization(output.Organization)
}

func (OrganizationPresenter) PresentUpdate(output organizationboundary.UpdateOrganizationOutput) *resourcemanagerv1.Organization {
	return mapOrganization(output.Organization)
}

func (OrganizationPresenter) PresentUndelete(output organizationboundary.UndeleteOrganizationOutput) *resourcemanagerv1.Organization {
	return mapOrganization(output.Organization)
}

func (OrganizationPresenter) PresentList(output organizationboundary.ListOrganizationsOutput) *resourcemanagerv1.ListOrganizationsResponse {
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
