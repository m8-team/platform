package workspacecommand

import (
	"context"
	"fmt"

	"github.com/m8platform/platform/internal/entity/resourcemanager/workspace"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundary"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/port"
)

type UpdateInteractor struct {
	TxManager        port.TxManager
	Repository       port.WorkspaceRepository
	IdempotencyStore port.IdempotencyStore
	OutboxWriter     port.OutboxWriter
	Clock            port.Clock
	UUIDGenerator    port.UUIDGenerator
}

func (i UpdateInteractor) Execute(ctx context.Context, input boundary.UpdateWorkspaceInput) (boundary.UpdateWorkspaceOutput, error) {
	var output boundary.UpdateWorkspaceOutput
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

		output = boundary.UpdateWorkspaceOutput{Workspace: usecasecommon.WorkspaceToBoundary(entity)}
		return nil
	})
	return output, err
}
