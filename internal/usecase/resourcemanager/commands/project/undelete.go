package projectcmd

import (
	"context"
	"fmt"

	"github.com/m8platform/platform/internal/domainservices/resourcemanager"
	"github.com/m8platform/platform/internal/entities/resourcemanager/project"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundaries"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/ports"
)

type UndeleteInteractor struct {
	TxManager        ports.TxManager
	Repository       ports.ProjectRepository
	HierarchyReader  ports.HierarchyReader
	UndeletePolicy   domainservices.UndeletePolicy
	IdempotencyStore ports.IdempotencyStore
	OutboxWriter     ports.OutboxWriter
	Clock            ports.Clock
	UUIDGenerator    ports.UUIDGenerator
}

func (i UndeleteInteractor) Execute(ctx context.Context, input boundaries.UndeleteProjectInput) (boundaries.UndeleteProjectOutput, error) {
	var output boundaries.UndeleteProjectOutput
	err := i.TxManager.WithinTx(ctx, func(ctx context.Context) error {
		reservation, err := usecasecommon.ReserveIdempotency(ctx, i.IdempotencyStore, "UndeleteProject:"+input.ID, input.Metadata.IdempotencyKey, usecasecommon.DefaultIdempotencyTTL)
		if err != nil {
			return err
		}

		entity, err := i.Repository.GetByID(ctx, input.ID, true)
		if err != nil {
			return fmt.Errorf("load project: %w", err)
		}
		parent, err := i.HierarchyReader.GetWorkspaceNode(ctx, entity.WorkspaceID)
		if err != nil {
			return fmt.Errorf("load workspace: %w", err)
		}
		if err := i.UndeletePolicy.EnsureParentAllowsUndelete(parent.Exists, parent.Deleted); err != nil {
			return err
		}

		now := i.Clock.Now().UTC()
		if err := entity.Undelete(now, i.UUIDGenerator.NewString()); err != nil {
			return err
		}
		if err := i.Repository.Update(ctx, entity); err != nil {
			return fmt.Errorf("persist project undelete: %w", err)
		}

		record, err := usecasecommon.NewOutboxRecord(i.UUIDGenerator, input.Metadata, project.EventUndeleted, "project", entity.ID, entity.WorkspaceID, entity.ETag.String(), now, entity)
		if err != nil {
			return err
		}
		if err := usecasecommon.WriteOutboxRecord(ctx, i.OutboxWriter, record); err != nil {
			return err
		}
		if err := usecasecommon.CompleteIdempotency(ctx, i.IdempotencyStore, reservation); err != nil {
			return err
		}

		output = boundaries.UndeleteProjectOutput{Project: usecasecommon.ProjectToBoundary(entity)}
		return nil
	})
	return output, err
}
