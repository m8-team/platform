package projectcommand

import (
	"context"
	"fmt"

	"github.com/m8platform/platform/internal/domainservices/resourcemanager"
	"github.com/m8platform/platform/internal/entity/resourcemanager/project"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundary"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/port"
)

type UndeleteInteractor struct {
	TxManager        port.TxManager
	Repository       port.ProjectRepository
	HierarchyReader  port.HierarchyReader
	UndeletePolicy   domainservices.UndeletePolicy
	IdempotencyStore port.IdempotencyStore
	OutboxWriter     port.OutboxWriter
	Clock            port.Clock
	UUIDGenerator    port.UUIDGenerator
}

func (i UndeleteInteractor) Execute(ctx context.Context, input boundary.UndeleteProjectInput) (boundary.UndeleteProjectOutput, error) {
	var output boundary.UndeleteProjectOutput
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

		output = boundary.UndeleteProjectOutput{Project: usecasecommon.ProjectToBoundary(entity)}
		return nil
	})
	return output, err
}
