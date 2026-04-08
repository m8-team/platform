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

type CreateInteractor struct {
	Executor      CommandExecutor
	Writer        port.OrganizationWriter
	OutboxWriter  port.OutboxWriter
	Clock         port.Clock
	UUIDGenerator port.UUIDGenerator
}

func (i CreateInteractor) Execute(ctx context.Context, input organizationboundary.CreateOrganizationInput) (organizationboundary.CreateOrganizationOutput, error) {
	var output organizationboundary.CreateOrganizationOutput
	err := i.Executor.Execute(ctx, "CreateOrganization", "", func(ctx context.Context) error {
		now := i.Clock.Now().UTC()
		entity, err := organizationentity.New(organizationentity.CreateParams{
			ID:          i.UUIDGenerator.NewString(),
			Name:        input.Name,
			Description: input.Description,
			Annotations: input.Annotations,
			Now:         now,
			ETag:        i.UUIDGenerator.NewString(),
		})
		if err != nil {
			return err
		}
		if err := i.Writer.Create(ctx, entity); err != nil {
			return fmt.Errorf("create organization: %w", err)
		}

		record, err := usecasecommon.NewOutboxRecord(
			i.UUIDGenerator,
			organizationentity.EventCreated,
			"organization",
			entity.ID,
			"",
			entity.ETag.String(),
			now,
			entity,
		)
		if err != nil {
			return err
		}
		if err := usecasecommon.WriteOutboxRecord(ctx, i.OutboxWriter, record); err != nil {
			return err
		}

		output = organizationboundary.CreateOrganizationOutput{
			Organization: organizationmapper.ToBoundary(entity),
		}
		return nil
	})
	return output, err
}
