package command

import (
	"context"
	"fmt"

	appcommon "github.com/m8platform/platform/internal/application/common"
	"github.com/m8platform/platform/internal/domain/hierarchy"
	"github.com/m8platform/platform/internal/domain/workspace"
	"github.com/m8platform/platform/internal/infra/events"
	"github.com/m8platform/platform/internal/ports"
)

type CreateWorkspace struct {
	Metadata       appcommon.Metadata
	OrganizationID string
	Name           string
	Description    string
	Annotations    map[string]string
}

type CreateWorkspaceHandler struct {
	TxManager   ports.TxManager
	Repository  ports.WorkspaceRepository
	Hierarchy   ports.HierarchyRepository
	Idempotency ports.IdempotencyStore
	Outbox      ports.OutboxWriter
	Clock       ports.Clock
	UUIDs       ports.UUIDGenerator
}

func (h CreateWorkspaceHandler) Handle(ctx context.Context, cmd CreateWorkspace) (workspace.Workspace, error) {
	var created workspace.Workspace
	err := h.TxManager.WithinTx(ctx, func(ctx context.Context) error {
		reservation, err := appcommon.ReserveIdempotency(ctx, h.Idempotency, "CreateWorkspace:"+cmd.OrganizationID, cmd.Metadata.IdempotencyKey, defaultIdempotencyTTL)
		if err != nil {
			return err
		}
		parent, err := h.Hierarchy.GetOrganizationNode(ctx, cmd.OrganizationID)
		if err != nil {
			return fmt.Errorf("load organization parent: %w", err)
		}
		if err := hierarchy.EnsureParentActive(parent.Exists, parent.Deleted); err != nil {
			return err
		}

		now := h.Clock.Now().UTC()
		created, err = workspace.New(workspace.CreateParams{
			ID:             h.UUIDs.NewString(),
			OrganizationID: cmd.OrganizationID,
			Name:           cmd.Name,
			Description:    cmd.Description,
			Annotations:    cmd.Annotations,
			Now:            now,
			ETag:           h.UUIDs.NewString(),
		})
		if err != nil {
			return err
		}
		if err := h.Repository.Create(ctx, created); err != nil {
			return fmt.Errorf("create workspace: %w", err)
		}
		if err := appendWorkspaceEvent(ctx, h.Outbox, h.UUIDs, created.CreatedEvent(now), cmd.Metadata); err != nil {
			return err
		}
		return appcommon.CompleteIdempotency(ctx, h.Idempotency, reservation)
	})
	return created, err
}

type UpdateWorkspace struct {
	Metadata       appcommon.Metadata
	ID             string
	OrganizationID string
	ETag           string
	UpdateMask     []string
	Name           *string
	Description    *string
	Annotations    map[string]string
}

type UpdateWorkspaceHandler struct {
	TxManager   ports.TxManager
	Repository  ports.WorkspaceRepository
	Idempotency ports.IdempotencyStore
	Outbox      ports.OutboxWriter
	Clock       ports.Clock
	UUIDs       ports.UUIDGenerator
}

func (h UpdateWorkspaceHandler) Handle(ctx context.Context, cmd UpdateWorkspace) (workspace.Workspace, error) {
	var updated workspace.Workspace
	err := h.TxManager.WithinTx(ctx, func(ctx context.Context) error {
		reservation, err := appcommon.ReserveIdempotency(ctx, h.Idempotency, "UpdateWorkspace:"+cmd.ID, cmd.Metadata.IdempotencyKey, defaultIdempotencyTTL)
		if err != nil {
			return err
		}
		if err := appcommon.ValidateMask(cmd.UpdateMask, []string{"name", "description", "annotations"}); err != nil {
			return err
		}

		updated, err = h.Repository.GetByID(ctx, cmd.ID, true)
		if err != nil {
			return fmt.Errorf("load workspace: %w", err)
		}
		now := h.Clock.Now().UTC()
		if err := updated.Update(cmd.UpdateMask, workspace.UpdateFields{
			ID:             cmd.ID,
			OrganizationID: cmd.OrganizationID,
			Name:           cmd.Name,
			Description:    cmd.Description,
			Annotations:    cmd.Annotations,
			ETag:           cmd.ETag,
		}, now, h.UUIDs.NewString()); err != nil {
			return err
		}
		if err := h.Repository.Update(ctx, updated); err != nil {
			return fmt.Errorf("update workspace: %w", err)
		}
		if err := appendWorkspaceEvent(ctx, h.Outbox, h.UUIDs, updated.UpdatedEvent(now), cmd.Metadata); err != nil {
			return err
		}
		return appcommon.CompleteIdempotency(ctx, h.Idempotency, reservation)
	})
	return updated, err
}

type DeleteWorkspace struct {
	Metadata     appcommon.Metadata
	ID           string
	ETag         string
	AllowMissing bool
}

type DeleteWorkspaceHandler struct {
	TxManager   ports.TxManager
	Repository  ports.WorkspaceRepository
	Hierarchy   ports.HierarchyRepository
	Idempotency ports.IdempotencyStore
	Outbox      ports.OutboxWriter
	Clock       ports.Clock
	UUIDs       ports.UUIDGenerator
}

func (h DeleteWorkspaceHandler) Handle(ctx context.Context, cmd DeleteWorkspace) error {
	return h.TxManager.WithinTx(ctx, func(ctx context.Context) error {
		reservation, err := appcommon.ReserveIdempotency(ctx, h.Idempotency, "DeleteWorkspace:"+cmd.ID, cmd.Metadata.IdempotencyKey, defaultIdempotencyTTL)
		if err != nil {
			return err
		}
		aggregate, err := h.Repository.GetByID(ctx, cmd.ID, true)
		if err != nil {
			if cmd.AllowMissing && err == workspace.ErrNotFound {
				return appcommon.CompleteIdempotency(ctx, h.Idempotency, reservation)
			}
			return fmt.Errorf("load workspace: %w", err)
		}
		hasChildren, err := h.Hierarchy.HasActiveProjects(ctx, cmd.ID)
		if err != nil {
			return fmt.Errorf("check active projects: %w", err)
		}
		if err := hierarchy.EnsureDeleteAllowed(hasChildren); err != nil {
			return err
		}

		now := h.Clock.Now().UTC()
		if err := aggregate.SoftDelete(now, now.Add(defaultPurgeWindow), cmd.ETag, h.UUIDs.NewString()); err != nil {
			return err
		}
		if err := h.Repository.Update(ctx, aggregate); err != nil {
			return fmt.Errorf("persist workspace delete: %w", err)
		}
		if err := appendWorkspaceEvent(ctx, h.Outbox, h.UUIDs, aggregate.DeletedEvent(now), cmd.Metadata); err != nil {
			return err
		}
		return appcommon.CompleteIdempotency(ctx, h.Idempotency, reservation)
	})
}

type UndeleteWorkspace struct {
	Metadata appcommon.Metadata
	ID       string
}

type UndeleteWorkspaceHandler struct {
	TxManager   ports.TxManager
	Repository  ports.WorkspaceRepository
	Hierarchy   ports.HierarchyRepository
	Idempotency ports.IdempotencyStore
	Outbox      ports.OutboxWriter
	Clock       ports.Clock
	UUIDs       ports.UUIDGenerator
}

func (h UndeleteWorkspaceHandler) Handle(ctx context.Context, cmd UndeleteWorkspace) (workspace.Workspace, error) {
	var restored workspace.Workspace
	err := h.TxManager.WithinTx(ctx, func(ctx context.Context) error {
		reservation, err := appcommon.ReserveIdempotency(ctx, h.Idempotency, "UndeleteWorkspace:"+cmd.ID, cmd.Metadata.IdempotencyKey, defaultIdempotencyTTL)
		if err != nil {
			return err
		}
		restored, err = h.Repository.GetByID(ctx, cmd.ID, true)
		if err != nil {
			return fmt.Errorf("load workspace: %w", err)
		}
		parent, err := h.Hierarchy.GetOrganizationNode(ctx, restored.OrganizationID)
		if err != nil {
			return fmt.Errorf("load organization parent: %w", err)
		}
		if err := hierarchy.EnsureUndeleteAllowed(parent.Exists, parent.Deleted); err != nil {
			return err
		}

		now := h.Clock.Now().UTC()
		if err := restored.Undelete(now, h.UUIDs.NewString()); err != nil {
			return err
		}
		if err := h.Repository.Update(ctx, restored); err != nil {
			return fmt.Errorf("persist workspace undelete: %w", err)
		}
		if err := appendWorkspaceEvent(ctx, h.Outbox, h.UUIDs, restored.UndeletedEvent(now), cmd.Metadata); err != nil {
			return err
		}
		return appcommon.CompleteIdempotency(ctx, h.Idempotency, reservation)
	})
	return restored, err
}

func appendWorkspaceEvent(
	ctx context.Context,
	outbox ports.OutboxWriter,
	generator ports.UUIDGenerator,
	event workspace.Event,
	meta appcommon.Metadata,
) error {
	if outbox == nil {
		return nil
	}
	record, err := events.WorkspaceRecord(generator, event, meta)
	if err != nil {
		return err
	}
	return outbox.Append(ctx, record)
}
