package projectcmd

import (
	"context"
	"fmt"

	"github.com/m8platform/platform/internal/entities/resourcemanager/project"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundaries"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/ports"
)

type DeleteInteractor struct {
	TxManager        ports.TxManager
	Repository       ports.ProjectRepository
	IdempotencyStore ports.IdempotencyStore
	OutboxWriter     ports.OutboxWriter
	Clock            ports.Clock
	UUIDGenerator    ports.UUIDGenerator
}

func (i DeleteInteractor) Execute(ctx context.Context, input boundaries.DeleteProjectInput) (boundaries.DeleteProjectOutput, error) {
	err := i.TxManager.WithinTx(ctx, func(ctx context.Context) error {
		reservation, err := usecasecommon.ReserveIdempotency(ctx, i.IdempotencyStore, "DeleteProject:"+input.ID, input.Metadata.IdempotencyKey, usecasecommon.DefaultIdempotencyTTL)
		if err != nil {
			return err
		}

		entity, err := i.Repository.GetByID(ctx, input.ID, true)
		if err != nil {
			if input.AllowMissing && err == project.ErrNotFound {
				return usecasecommon.CompleteIdempotency(ctx, i.IdempotencyStore, reservation)
			}
			return fmt.Errorf("load project: %w", err)
		}

		now := i.Clock.Now().UTC()
		if err := entity.SoftDelete(now, now.Add(usecasecommon.DefaultPurgeWindow), input.ETag, i.UUIDGenerator.NewString()); err != nil {
			return err
		}
		if err := i.Repository.Update(ctx, entity); err != nil {
			return fmt.Errorf("persist project delete: %w", err)
		}

		record, err := usecasecommon.NewOutboxRecord(i.UUIDGenerator, input.Metadata, project.EventDeleted, "project", entity.ID, entity.WorkspaceID, entity.ETag.String(), now, entity)
		if err != nil {
			return err
		}
		if err := usecasecommon.WriteOutboxRecord(ctx, i.OutboxWriter, record); err != nil {
			return err
		}
		return usecasecommon.CompleteIdempotency(ctx, i.IdempotencyStore, reservation)
	})
	return boundaries.DeleteProjectOutput{}, err
}
