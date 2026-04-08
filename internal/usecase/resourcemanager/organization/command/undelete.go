package organizationcommand

import (
	"context"
	"fmt"

	organizationentity "github.com/m8platform/platform/internal/entity/resourcemanager/organization"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	organizationboundary "github.com/m8platform/platform/internal/usecase/resourcemanager/organization/boundary"
	organizationmapper "github.com/m8platform/platform/internal/usecase/resourcemanager/organization/mapper"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/port"
)

type UndeleteInteractor struct {
	Executor      CommandExecutor
	Reader        port.OrganizationReader
	Writer        port.OrganizationWriter
	OutboxWriter  port.OutboxWriter
	Clock         port.Clock
	UUIDGenerator port.UUIDGenerator
}

func (i UndeleteInteractor) Execute(ctx context.Context, input organizationboundary.UndeleteOrganizationInput) (organizationboundary.UndeleteOrganizationOutput, error) {
	var output organizationboundary.UndeleteOrganizationOutput
	err := i.Executor.Execute(ctx, "UndeleteOrganization:"+input.ID, input.Metadata.IdempotencyKey, func(ctx context.Context) error {
		entity, err := i.Reader.GetByID(ctx, input.ID, true)
		if err != nil {
			return fmt.Errorf("load organization: %w", err)
		}

		now := i.Clock.Now().UTC()
		if err := entity.Undelete(now, i.UUIDGenerator.NewString()); err != nil {
			return err
		}
		if err := i.Writer.Undelete(ctx, entity); err != nil {
			return fmt.Errorf("persist organization undelete: %w", err)
		}

		record, err := usecasecommon.NewOutboxRecord(i.UUIDGenerator, input.Metadata, organizationentity.EventUndeleted, "organization", entity.ID, "", entity.ETag.String(), now, entity)
		if err != nil {
			return err
		}
		if err := usecasecommon.WriteOutboxRecord(ctx, i.OutboxWriter, record); err != nil {
			return err
		}

		output = organizationboundary.UndeleteOrganizationOutput{
			Organization: organizationmapper.ToBoundary(entity),
		}
		return nil
	})
	return output, err
}
