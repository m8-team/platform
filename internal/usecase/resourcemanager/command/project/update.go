package projectcommand

import (
	"context"
	"fmt"

	"github.com/m8platform/platform/internal/entity/resourcemanager/project"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundary"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/port"
)

type UpdateInteractor struct {
	TxManager        port.TxManager
	Repository       port.ProjectRepository
	IdempotencyStore port.IdempotencyStore
	OutboxWriter     port.OutboxWriter
	Clock            port.Clock
	UUIDGenerator    port.UUIDGenerator
}

func (i UpdateInteractor) Execute(ctx context.Context, input boundary.UpdateProjectInput) (boundary.UpdateProjectOutput, error) {
	var output boundary.UpdateProjectOutput
	err := i.TxManager.WithinTx(ctx, func(ctx context.Context) error {
		reservation, err := usecasecommon.ReserveIdempotency(ctx, i.IdempotencyStore, "UpdateProject:"+input.ID, input.Metadata.IdempotencyKey, usecasecommon.DefaultIdempotencyTTL)
		if err != nil {
			return err
		}
		if err := validateMask(input.UpdateMask, project.AllowedUpdatePaths); err != nil {
			return err
		}

		entity, err := i.Repository.GetByID(ctx, input.ID, true)
		if err != nil {
			return fmt.Errorf("load project: %w", err)
		}
		now := i.Clock.Now().UTC()
		if err := entity.Update(input.UpdateMask, project.UpdateParams{
			ID:          input.ID,
			WorkspaceID: input.WorkspaceID,
			Name:        input.Name,
			Description: input.Description,
			Annotations: input.Annotations,
			ETag:        input.ETag,
		}, now, i.UUIDGenerator.NewString()); err != nil {
			return err
		}
		if err := i.Repository.Update(ctx, entity); err != nil {
			return fmt.Errorf("update project: %w", err)
		}

		record, err := usecasecommon.NewOutboxRecord(i.UUIDGenerator, input.Metadata, project.EventUpdated, "project", entity.ID, entity.WorkspaceID, entity.ETag.String(), now, entity)
		if err != nil {
			return err
		}
		if err := usecasecommon.WriteOutboxRecord(ctx, i.OutboxWriter, record); err != nil {
			return err
		}
		if err := usecasecommon.CompleteIdempotency(ctx, i.IdempotencyStore, reservation); err != nil {
			return err
		}

		output = boundary.UpdateProjectOutput{Project: usecasecommon.ProjectToBoundary(entity)}
		return nil
	})
	return output, err
}
