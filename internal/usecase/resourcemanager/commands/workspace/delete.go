package workspacecmd

import (
	"context"
	"fmt"

	"github.com/m8platform/platform/internal/domainservices/resourcemanager"
	"github.com/m8platform/platform/internal/entities/resourcemanager/workspace"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundaries"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/ports"
)

type DeleteInteractor struct {
	TxManager        ports.TxManager
	Repository       ports.WorkspaceRepository
	HierarchyReader  ports.HierarchyReader
	DeletePolicy     domainservices.DeletePolicy
	IdempotencyStore ports.IdempotencyStore
	OutboxWriter     ports.OutboxWriter
	Clock            ports.Clock
	UUIDGenerator    ports.UUIDGenerator
}

func (i DeleteInteractor) Execute(ctx context.Context, input boundaries.DeleteWorkspaceInput) (boundaries.DeleteWorkspaceOutput, error) {
	err := i.TxManager.WithinTx(ctx, func(ctx context.Context) error {
		reservation, err := usecasecommon.ReserveIdempotency(ctx, i.IdempotencyStore, "DeleteWorkspace:"+input.ID, input.Metadata.IdempotencyKey, usecasecommon.DefaultIdempotencyTTL)
		if err != nil {
			return err
		}

		entity, err := i.Repository.GetByID(ctx, input.ID, true)
		if err != nil {
			if input.AllowMissing && err == workspace.ErrNotFound {
				return usecasecommon.CompleteIdempotency(ctx, i.IdempotencyStore, reservation)
			}
			return fmt.Errorf("load workspace: %w", err)
		}
		hasChildren, err := i.HierarchyReader.HasActiveProjects(ctx, input.ID)
		if err != nil {
			return fmt.Errorf("check active projects: %w", err)
		}
		if err := i.DeletePolicy.EnsureAllowed(hasChildren); err != nil {
			return err
		}

		now := i.Clock.Now().UTC()
		if err := entity.SoftDelete(now, now.Add(usecasecommon.DefaultPurgeWindow), input.ETag, i.UUIDGenerator.NewString()); err != nil {
			return err
		}
		if err := i.Repository.Update(ctx, entity); err != nil {
			return fmt.Errorf("persist workspace delete: %w", err)
		}

		record, err := usecasecommon.NewOutboxRecord(i.UUIDGenerator, input.Metadata, workspace.EventDeleted, "workspace", entity.ID, entity.OrganizationID, entity.ETag.String(), now, entity)
		if err != nil {
			return err
		}
		if err := usecasecommon.WriteOutboxRecord(ctx, i.OutboxWriter, record); err != nil {
			return err
		}
		return usecasecommon.CompleteIdempotency(ctx, i.IdempotencyStore, reservation)
	})
	return boundaries.DeleteWorkspaceOutput{}, err
}
