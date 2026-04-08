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

type UndeleteInteractor struct {
	TxManager        ports.TxManager
	Repository       ports.WorkspaceRepository
	HierarchyReader  ports.HierarchyReader
	UndeletePolicy   domainservices.UndeletePolicy
	IdempotencyStore ports.IdempotencyStore
	OutboxWriter     ports.OutboxWriter
	Clock            ports.Clock
	UUIDGenerator    ports.UUIDGenerator
}

func (i UndeleteInteractor) Execute(ctx context.Context, input boundaries.UndeleteWorkspaceInput) (boundaries.UndeleteWorkspaceOutput, error) {
	var output boundaries.UndeleteWorkspaceOutput
	err := i.TxManager.WithinTx(ctx, func(ctx context.Context) error {
		reservation, err := usecasecommon.ReserveIdempotency(ctx, i.IdempotencyStore, "UndeleteWorkspace:"+input.ID, input.Metadata.IdempotencyKey, usecasecommon.DefaultIdempotencyTTL)
		if err != nil {
			return err
		}

		entity, err := i.Repository.GetByID(ctx, input.ID, true)
		if err != nil {
			return fmt.Errorf("load workspace: %w", err)
		}
		parent, err := i.HierarchyReader.GetOrganizationNode(ctx, entity.OrganizationID)
		if err != nil {
			return fmt.Errorf("load organization: %w", err)
		}
		if err := i.UndeletePolicy.EnsureParentAllowsUndelete(parent.Exists, parent.Deleted); err != nil {
			return err
		}

		now := i.Clock.Now().UTC()
		if err := entity.Undelete(now, i.UUIDGenerator.NewString()); err != nil {
			return err
		}
		if err := i.Repository.Update(ctx, entity); err != nil {
			return fmt.Errorf("persist workspace undelete: %w", err)
		}

		record, err := usecasecommon.NewOutboxRecord(i.UUIDGenerator, input.Metadata, workspace.EventUndeleted, "workspace", entity.ID, entity.OrganizationID, entity.ETag.String(), now, entity)
		if err != nil {
			return err
		}
		if err := usecasecommon.WriteOutboxRecord(ctx, i.OutboxWriter, record); err != nil {
			return err
		}
		if err := usecasecommon.CompleteIdempotency(ctx, i.IdempotencyStore, reservation); err != nil {
			return err
		}

		output = boundaries.UndeleteWorkspaceOutput{Workspace: usecasecommon.WorkspaceToBoundary(entity)}
		return nil
	})
	return output, err
}
