package organizationcommand

import (
	"context"
	"errors"
	"fmt"

	domainservices "github.com/m8platform/platform/internal/domainservices/resourcemanager"
	organizationentity "github.com/m8platform/platform/internal/entity/resourcemanager/organization"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	organizationboundary "github.com/m8platform/platform/internal/usecase/resourcemanager/organization/boundary"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/port"
)

type DeleteInteractor struct {
	Executor        CommandExecutor
	Reader          port.OrganizationReader
	Writer          port.OrganizationWriter
	HierarchyReader port.HierarchyReader
	DeletePolicy    domainservices.DeletePolicy
	OutboxWriter    port.OutboxWriter
	Clock           port.Clock
	UUIDGenerator   port.UUIDGenerator
}

func (i DeleteInteractor) Execute(ctx context.Context, input organizationboundary.DeleteOrganizationInput) (organizationboundary.DeleteOrganizationOutput, error) {
	err := i.Executor.Execute(ctx, "DeleteOrganization:"+input.ID, input.Metadata.IdempotencyKey, func(ctx context.Context) error {
		entity, err := i.Reader.GetByID(ctx, input.ID, true)
		if err != nil {
			if input.AllowMissing && errors.Is(err, organizationentity.ErrNotFound) {
				return nil
			}
			return fmt.Errorf("load organization: %w", err)
		}

		hasChildren, err := i.HierarchyReader.HasActiveWorkspaces(ctx, input.ID)
		if err != nil {
			return fmt.Errorf("check active workspaces: %w", err)
		}
		if err := i.DeletePolicy.EnsureAllowed(hasChildren); err != nil {
			return err
		}

		now := i.Clock.Now().UTC()
		if err := entity.SoftDelete(now, now.Add(usecasecommon.DefaultPurgeWindow), input.ETag, i.UUIDGenerator.NewString()); err != nil {
			return err
		}
		if err := i.Writer.SoftDelete(ctx, entity); err != nil {
			return fmt.Errorf("persist organization delete: %w", err)
		}

		record, err := usecasecommon.NewOutboxRecord(i.UUIDGenerator, input.Metadata, organizationentity.EventDeleted, "organization", entity.ID, "", entity.ETag.String(), now, entity)
		if err != nil {
			return err
		}
		if err := usecasecommon.WriteOutboxRecord(ctx, i.OutboxWriter, record); err != nil {
			return err
		}
		return nil
	})
	return organizationboundary.DeleteOrganizationOutput{}, err
}
