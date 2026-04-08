package workspacecommand

import (
	"context"
	"fmt"

	"github.com/m8platform/platform/internal/domainservices/resourcemanager"
	"github.com/m8platform/platform/internal/entity/resourcemanager/workspace"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/boundary"
	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/port"
)

type CreateInteractor struct {
	TxManager        port.TxManager
	Repository       port.WorkspaceRepository
	HierarchyReader  port.HierarchyReader
	HierarchyPolicy  domainservices.HierarchyPolicy
	IdempotencyStore port.IdempotencyStore
	OutboxWriter     port.OutboxWriter
	Clock            port.Clock
	UUIDGenerator    port.UUIDGenerator
}

func (i CreateInteractor) Execute(ctx context.Context, input boundary.CreateWorkspaceInput) (boundary.CreateWorkspaceOutput, error) {
	var output boundary.CreateWorkspaceOutput
	err := i.TxManager.WithinTx(ctx, func(ctx context.Context) error {
		reservation, err := usecasecommon.ReserveIdempotency(ctx, i.IdempotencyStore, "CreateWorkspace:"+input.OrganizationID, input.Metadata.IdempotencyKey, usecasecommon.DefaultIdempotencyTTL)
		if err != nil {
			return err
		}

		parent, err := i.HierarchyReader.GetOrganizationNode(ctx, input.OrganizationID)
		if err != nil {
			return fmt.Errorf("load organization: %w", err)
		}
		if err := i.HierarchyPolicy.EnsureParentActive(parent.Exists, parent.Deleted); err != nil {
			return err
		}

		now := i.Clock.Now().UTC()
		entity, err := workspace.New(workspace.CreateParams{
			ID:             i.UUIDGenerator.NewString(),
			OrganizationID: input.OrganizationID,
			Name:           input.Name,
			Description:    input.Description,
			Annotations:    input.Annotations,
			Now:            now,
			ETag:           i.UUIDGenerator.NewString(),
		})
		if err != nil {
			return err
		}
		if err := i.Repository.Create(ctx, entity); err != nil {
			return fmt.Errorf("create workspace: %w", err)
		}

		record, err := usecasecommon.NewOutboxRecord(
			i.UUIDGenerator,
			input.Metadata,
			workspace.EventCreated,
			"workspace",
			entity.ID,
			entity.OrganizationID,
			entity.ETag.String(),
			now,
			entity,
		)
		if err != nil {
			return err
		}
		if err := usecasecommon.WriteOutboxRecord(ctx, i.OutboxWriter, record); err != nil {
			return err
		}
		if err := usecasecommon.CompleteIdempotency(ctx, i.IdempotencyStore, reservation); err != nil {
			return err
		}

		output = boundary.CreateWorkspaceOutput{Workspace: usecasecommon.WorkspaceToBoundary(entity)}
		return nil
	})
	return output, err
}
