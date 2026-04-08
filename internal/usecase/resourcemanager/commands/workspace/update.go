package workspacecmd

import (
	"context"
	"fmt"

	"github.com/m8platform/platform/internal/entities/resourcemanager/workspace"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundaries"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/ports"
)

type UpdateInteractor struct {
	TxManager        ports.TxManager
	Repository       ports.WorkspaceRepository
	IdempotencyStore ports.IdempotencyStore
	OutboxWriter     ports.OutboxWriter
	Clock            ports.Clock
	UUIDGenerator    ports.UUIDGenerator
}

func (i UpdateInteractor) Execute(ctx context.Context, input boundaries.UpdateWorkspaceInput) (boundaries.UpdateWorkspaceOutput, error) {
	var output boundaries.UpdateWorkspaceOutput
	err := i.TxManager.WithinTx(ctx, func(ctx context.Context) error {
		reservation, err := usecasecommon.ReserveIdempotency(ctx, i.IdempotencyStore, "UpdateWorkspace:"+input.ID, input.Metadata.IdempotencyKey, usecasecommon.DefaultIdempotencyTTL)
		if err != nil {
			return err
		}
		if err := validateMask(input.UpdateMask, workspace.AllowedUpdatePaths); err != nil {
			return err
		}

		entity, err := i.Repository.GetByID(ctx, input.ID, true)
		if err != nil {
			return fmt.Errorf("load workspace: %w", err)
		}
		now := i.Clock.Now().UTC()
		if err := entity.Update(input.UpdateMask, workspace.UpdateParams{
			ID:             input.ID,
			OrganizationID: input.OrganizationID,
			Name:           input.Name,
			Description:    input.Description,
			Annotations:    input.Annotations,
			ETag:           input.ETag,
		}, now, i.UUIDGenerator.NewString()); err != nil {
			return err
		}
		if err := i.Repository.Update(ctx, entity); err != nil {
			return fmt.Errorf("update workspace: %w", err)
		}

		record, err := usecasecommon.NewOutboxRecord(i.UUIDGenerator, input.Metadata, workspace.EventUpdated, "workspace", entity.ID, entity.OrganizationID, entity.ETag.String(), now, entity)
		if err != nil {
			return err
		}
		if err := usecasecommon.WriteOutboxRecord(ctx, i.OutboxWriter, record); err != nil {
			return err
		}
		if err := usecasecommon.CompleteIdempotency(ctx, i.IdempotencyStore, reservation); err != nil {
			return err
		}

		output = boundaries.UpdateWorkspaceOutput{Workspace: usecasecommon.WorkspaceToBoundary(entity)}
		return nil
	})
	return output, err
}
