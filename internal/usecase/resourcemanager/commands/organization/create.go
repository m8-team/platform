package organizationcmd

import (
	"context"
	"fmt"

	"github.com/m8platform/platform/internal/entities/resourcemanager/organization"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundaries"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/ports"
)

type CreateInteractor struct {
	TxManager        ports.TxManager
	Repository       ports.OrganizationRepository
	IdempotencyStore ports.IdempotencyStore
	OutboxWriter     ports.OutboxWriter
	Clock            ports.Clock
	UUIDGenerator    ports.UUIDGenerator
}

func (i CreateInteractor) Execute(ctx context.Context, input boundaries.CreateOrganizationInput) (boundaries.CreateOrganizationOutput, error) {
	var output boundaries.CreateOrganizationOutput
	err := i.TxManager.WithinTx(ctx, func(ctx context.Context) error {
		reservation, err := usecasecommon.ReserveIdempotency(ctx, i.IdempotencyStore, "CreateOrganization", input.Metadata.IdempotencyKey, usecasecommon.DefaultIdempotencyTTL)
		if err != nil {
			return err
		}

		now := i.Clock.Now().UTC()
		entity, err := organization.New(organization.CreateParams{
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
		if err := i.Repository.Create(ctx, entity); err != nil {
			return fmt.Errorf("create organization: %w", err)
		}

		record, err := usecasecommon.NewOutboxRecord(
			i.UUIDGenerator,
			input.Metadata,
			organization.EventCreated,
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
		if err := usecasecommon.CompleteIdempotency(ctx, i.IdempotencyStore, reservation); err != nil {
			return err
		}

		output = boundaries.CreateOrganizationOutput{Organization: usecasecommon.OrganizationToBoundary(entity)}
		return nil
	})
	return output, err
}
