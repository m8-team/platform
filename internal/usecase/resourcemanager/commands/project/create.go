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

type CreateInteractor struct {
	TxManager        ports.TxManager
	Repository       ports.ProjectRepository
	HierarchyReader  ports.HierarchyReader
	HierarchyPolicy  domainservices.HierarchyPolicy
	IdempotencyStore ports.IdempotencyStore
	OutboxWriter     ports.OutboxWriter
	Clock            ports.Clock
	UUIDGenerator    ports.UUIDGenerator
}

func (i CreateInteractor) Execute(ctx context.Context, input boundaries.CreateProjectInput) (boundaries.CreateProjectOutput, error) {
	var output boundaries.CreateProjectOutput
	err := i.TxManager.WithinTx(ctx, func(ctx context.Context) error {
		reservation, err := usecasecommon.ReserveIdempotency(ctx, i.IdempotencyStore, "CreateProject:"+input.WorkspaceID, input.Metadata.IdempotencyKey, usecasecommon.DefaultIdempotencyTTL)
		if err != nil {
			return err
		}

		parent, err := i.HierarchyReader.GetWorkspaceNode(ctx, input.WorkspaceID)
		if err != nil {
			return fmt.Errorf("load workspace: %w", err)
		}
		if err := i.HierarchyPolicy.EnsureParentActive(parent.Exists, parent.Deleted); err != nil {
			return err
		}

		now := i.Clock.Now().UTC()
		entity, err := project.New(project.CreateParams{
			ID:          i.UUIDGenerator.NewString(),
			WorkspaceID: input.WorkspaceID,
			Name:        input.Name,
			Description: input.Description,
			Annotations: input.Annotations,
			Now:         now,
			ETag:        i.UUIDGenerator.NewString(),
		})
		if err != nil {
			return err
		}
		if err := i.Repository.Create(ctx, entity); err != nil {
			return fmt.Errorf("create project: %w", err)
		}

		record, err := usecasecommon.NewOutboxRecord(
			i.UUIDGenerator,
			input.Metadata,
			project.EventCreated,
			"project",
			entity.ID,
			entity.WorkspaceID,
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

		output = boundaries.CreateProjectOutput{Project: usecasecommon.ProjectToBoundary(entity)}
		return nil
	})
	return output, err
}
