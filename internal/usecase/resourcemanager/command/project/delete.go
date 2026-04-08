package projectcommand

import (
	"context"
	"fmt"

	"github.com/m8platform/platform/internal/entity/resourcemanager/project"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundary"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/port"
)

type DeleteInteractor struct {
	TxManager        port.TxManager
	Repository       port.ProjectRepository
	IdempotencyStore port.IdempotencyStore
	OutboxWriter     port.OutboxWriter
	Clock            port.Clock
	UUIDGenerator    port.UUIDGenerator
}

func (i DeleteInteractor) Execute(ctx context.Context, input boundary.DeleteProjectInput) (boundary.DeleteProjectOutput, error) {
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
	return boundary.DeleteProjectOutput{}, err
}
