package grpctransport

import (
	"errors"

	"google.golang.org/grpc"
	resourcemanagerv1 "m8/platform/resourcemanager/v1"
)

type ServerSet struct {
	Organization *OrganizationServer
	Workspace    *WorkspaceServer
	Project      *ProjectServer
}

func (s ServerSet) Register(server grpc.ServiceRegistrar) error {
	if server == nil {
		return errors.New("grpc registrar is nil")
	}
	if s.Organization != nil {
		resourcemanagerv1.RegisterOrganizationServiceServer(server, s.Organization)
	}
	if s.Workspace != nil {
		resourcemanagerv1.RegisterWorkspaceServiceServer(server, s.Workspace)
	}
	if s.Project != nil {
		resourcemanagerv1.RegisterProjectServiceServer(server, s.Project)
	}
	return nil
}

func commandErrorInvalidArgument(message string) error {
	return errors.New(message)
}

func stringPtr(value string) *string {
	return &value
}
