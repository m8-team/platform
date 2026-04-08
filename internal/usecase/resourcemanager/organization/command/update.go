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

type UpdateInteractor struct {
	Executor       CommandExecutor
	Reader         port.OrganizationReader
	Writer         port.OrganizationWriter
	OutboxWriter   port.OutboxWriter
	Clock          port.Clock
	UUIDGenerator  port.UUIDGenerator
	InputValidator UpdateInputValidator
}

func (i UpdateInteractor) Execute(ctx context.Context, input organizationboundary.UpdateOrganizationInput) (organizationboundary.UpdateOrganizationOutput, error) {
	var output organizationboundary.UpdateOrganizationOutput
	err := i.Executor.Execute(ctx, "UpdateOrganization:"+input.ID, input.Metadata.IdempotencyKey, func(ctx context.Context) error {
		if i.InputValidator != nil {
			if err := i.InputValidator.Validate(input); err != nil {
				return err
			}
		}

		entity, err := i.Reader.GetByID(ctx, input.ID, true)
		if err != nil {
			return fmt.Errorf("load organization: %w", err)
		}

		var annotations map[string]string
		if input.Annotations != nil {
			annotations = cloneMap(*input.Annotations)
		}

		now := i.Clock.Now().UTC()
		if err := entity.Update(input.UpdateMask, organizationentity.UpdateParams{
			ID:          input.ID,
			Name:        input.Name,
			Description: input.Description,
			Annotations: annotations,
			ETag:        input.ETag,
		}, now, i.UUIDGenerator.NewString()); err != nil {
			return err
		}
		if err := i.Writer.Update(ctx, entity); err != nil {
			return fmt.Errorf("update organization: %w", err)
		}

		record, err := usecasecommon.NewOutboxRecord(i.UUIDGenerator, input.Metadata, organizationentity.EventUpdated, "organization", entity.ID, "", entity.ETag.String(), now, entity)
		if err != nil {
			return err
		}
		if err := usecasecommon.WriteOutboxRecord(ctx, i.OutboxWriter, record); err != nil {
			return err
		}

		output = organizationboundary.UpdateOrganizationOutput{
			Organization: organizationmapper.ToBoundary(entity),
		}
		return nil
	})
	return output, err
}

func cloneMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}
