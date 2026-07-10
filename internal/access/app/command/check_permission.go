package command

import (
	"fmt"

	"github.com/m8platform/platform/internal/access/domain"
)

type CheckPermissionCommand struct {
	SubjectType string
	SubjectID   string

	Permission string

	ResourceType string
	ResourceID   string

	ModelRevision string

	FailMode          domain.FailMode
	Critical          bool
	FailOpenReference string
}

func (c CheckPermissionCommand) ToDomain() (domain.CheckPermissionRequest, error) {
	subject, err := domain.NewSubject(c.SubjectType, c.SubjectID)
	if err != nil {
		return domain.CheckPermissionRequest{}, fmt.Errorf("build permission check subject: %w", err)
	}

	permission, err := domain.NewPermission(c.Permission)
	if err != nil {
		return domain.CheckPermissionRequest{}, fmt.Errorf("build permission check permission: %w", err)
	}

	resource, err := domain.NewResource(c.ResourceType, c.ResourceID)
	if err != nil {
		return domain.CheckPermissionRequest{}, fmt.Errorf("build permission check resource: %w", err)
	}

	revision, err := domain.NewModelRevision(c.ModelRevision)
	if err != nil {
		return domain.CheckPermissionRequest{}, fmt.Errorf("build permission check model revision: %w", err)
	}

	return domain.NewCheckPermissionRequest(domain.CheckPermissionInput{
		Subject:           subject,
		Permission:        permission,
		Resource:          resource,
		ModelRevision:     revision,
		FailMode:          c.FailMode,
		Critical:          c.Critical,
		FailOpenReference: c.FailOpenReference,
	})
}
