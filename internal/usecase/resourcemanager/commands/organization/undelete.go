package organizationcmd

import (
	"context"
	"fmt"

	"github.com/m8platform/platform/internal/entities/resourcemanager/organization"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundaries"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/ports"
)

type UndeleteInteractor struct {
	TxManager        ports.TxManager
	Repository       ports.OrganizationRepository
	IdempotencyStore ports.IdempotencyStore
	OutboxWriter     ports.OutboxWriter
	Clock            ports.Clock
	UUIDGenerator    ports.UUIDGenerator
}

func (i UndeleteInteractor) Execute(ctx context.Context, input boundaries.UndeleteOrganizationInput) (boundaries.UndeleteOrganizationOutput, error) {
	var output boundaries.UndeleteOrganizationOutput
	err := i.TxManager.WithinTx(ctx, func(ctx context.Context) error {
		reservation, err := usecasecommon.ReserveIdempotency(ctx, i.IdempotencyStore, "UndeleteOrganization:"+input.ID, input.Metadata.IdempotencyKey, usecasecommon.DefaultIdempotencyTTL)
		if err != nil {
			return err
		}

		entity, err := i.Repository.GetByID(ctx, input.ID, true)
		if err != nil {
			return fmt.Errorf("load organization: %w", err)
		}
		now := i.Clock.Now().UTC()
		if err := entity.Undelete(now, i.UUIDGenerator.NewString()); err != nil {
			return err
		}
		if err := i.Repository.Update(ctx, entity); err != nil {
			return fmt.Errorf("persist organization undelete: %w", err)
		}

		record, err := usecasecommon.NewOutboxRecord(i.UUIDGenerator, input.Metadata, organization.EventUndeleted, "organization", entity.ID, "", entity.ETag.String(), now, entity)
		if err != nil {
			return err
		}
		if err := usecasecommon.WriteOutboxRecord(ctx, i.OutboxWriter, record); err != nil {
			return err
		}
		if err := usecasecommon.CompleteIdempotency(ctx, i.IdempotencyStore, reservation); err != nil {
			return err
		}

		output = boundaries.UndeleteOrganizationOutput{Organization: usecasecommon.OrganizationToBoundary(entity)}
		return nil
	})
	return output, err
}
