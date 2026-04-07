package command

import (
	"context"
	"fmt"

	appcommon "github.com/m8platform/platform/internal/application/common"
	"github.com/m8platform/platform/internal/domain/hierarchy"
	"github.com/m8platform/platform/internal/domain/project"
	"github.com/m8platform/platform/internal/infra/events"
	"github.com/m8platform/platform/internal/ports"
)

type CreateProject struct {
	Metadata    appcommon.Metadata
	WorkspaceID string
	Name        string
	Description string
	Annotations map[string]string
}

type CreateProjectHandler struct {
	TxManager   ports.TxManager
	Repository  ports.ProjectRepository
	Hierarchy   ports.HierarchyRepository
	Idempotency ports.IdempotencyStore
	Outbox      ports.OutboxWriter
	Clock       ports.Clock
	UUIDs       ports.UUIDGenerator
}

func (h CreateProjectHandler) Handle(ctx context.Context, cmd CreateProject) (project.Project, error) {
	var created project.Project
	err := h.TxManager.WithinTx(ctx, func(ctx context.Context) error {
		reservation, err := appcommon.ReserveIdempotency(ctx, h.Idempotency, "CreateProject:"+cmd.WorkspaceID, cmd.Metadata.IdempotencyKey, defaultIdempotencyTTL)
		if err != nil {
			return err
		}
		parent, err := h.Hierarchy.GetWorkspaceNode(ctx, cmd.WorkspaceID)
		if err != nil {
			return fmt.Errorf("load workspace parent: %w", err)
		}
		if err := hierarchy.EnsureParentActive(parent.Exists, parent.Deleted); err != nil {
			return err
		}

		now := h.Clock.Now().UTC()
		created, err = project.New(project.CreateParams{
			ID:          h.UUIDs.NewString(),
			WorkspaceID: cmd.WorkspaceID,
			Name:        cmd.Name,
			Description: cmd.Description,
			Annotations: cmd.Annotations,
			Now:         now,
			ETag:        h.UUIDs.NewString(),
		})
		if err != nil {
			return err
		}
		if err := h.Repository.Create(ctx, created); err != nil {
			return fmt.Errorf("create project: %w", err)
		}
		if err := appendProjectEvent(ctx, h.Outbox, h.UUIDs, created.CreatedEvent(now), cmd.Metadata); err != nil {
			return err
		}
		return appcommon.CompleteIdempotency(ctx, h.Idempotency, reservation)
	})
	return created, err
}

type UpdateProject struct {
	Metadata    appcommon.Metadata
	ID          string
	WorkspaceID string
	ETag        string
	UpdateMask  []string
	Name        *string
	Description *string
	Annotations map[string]string
}

type UpdateProjectHandler struct {
	TxManager   ports.TxManager
	Repository  ports.ProjectRepository
	Idempotency ports.IdempotencyStore
	Outbox      ports.OutboxWriter
	Clock       ports.Clock
	UUIDs       ports.UUIDGenerator
}

func (h UpdateProjectHandler) Handle(ctx context.Context, cmd UpdateProject) (project.Project, error) {
	var updated project.Project
	err := h.TxManager.WithinTx(ctx, func(ctx context.Context) error {
		reservation, err := appcommon.ReserveIdempotency(ctx, h.Idempotency, "UpdateProject:"+cmd.ID, cmd.Metadata.IdempotencyKey, defaultIdempotencyTTL)
		if err != nil {
			return err
		}
		if err := appcommon.ValidateMask(cmd.UpdateMask, []string{"name", "description", "annotations"}); err != nil {
			return err
		}

		updated, err = h.Repository.GetByID(ctx, cmd.ID, true)
		if err != nil {
			return fmt.Errorf("load project: %w", err)
		}
		now := h.Clock.Now().UTC()
		if err := updated.Update(cmd.UpdateMask, project.UpdateFields{
			ID:          cmd.ID,
			WorkspaceID: cmd.WorkspaceID,
			Name:        cmd.Name,
			Description: cmd.Description,
			Annotations: cmd.Annotations,
			ETag:        cmd.ETag,
		}, now, h.UUIDs.NewString()); err != nil {
			return err
		}
		if err := h.Repository.Update(ctx, updated); err != nil {
			return fmt.Errorf("update project: %w", err)
		}
		if err := appendProjectEvent(ctx, h.Outbox, h.UUIDs, updated.UpdatedEvent(now), cmd.Metadata); err != nil {
			return err
		}
		return appcommon.CompleteIdempotency(ctx, h.Idempotency, reservation)
	})
	return updated, err
}

type DeleteProject struct {
	Metadata     appcommon.Metadata
	ID           string
	ETag         string
	AllowMissing bool
}

type DeleteProjectHandler struct {
	TxManager   ports.TxManager
	Repository  ports.ProjectRepository
	Idempotency ports.IdempotencyStore
	Outbox      ports.OutboxWriter
	Clock       ports.Clock
	UUIDs       ports.UUIDGenerator
}

func (h DeleteProjectHandler) Handle(ctx context.Context, cmd DeleteProject) error {
	return h.TxManager.WithinTx(ctx, func(ctx context.Context) error {
		reservation, err := appcommon.ReserveIdempotency(ctx, h.Idempotency, "DeleteProject:"+cmd.ID, cmd.Metadata.IdempotencyKey, defaultIdempotencyTTL)
		if err != nil {
			return err
		}
		aggregate, err := h.Repository.GetByID(ctx, cmd.ID, true)
		if err != nil {
			if cmd.AllowMissing && err == project.ErrNotFound {
				return appcommon.CompleteIdempotency(ctx, h.Idempotency, reservation)
			}
			return fmt.Errorf("load project: %w", err)
		}
		now := h.Clock.Now().UTC()
		if err := aggregate.SoftDelete(now, now.Add(defaultPurgeWindow), cmd.ETag, h.UUIDs.NewString()); err != nil {
			return err
		}
		if err := h.Repository.Update(ctx, aggregate); err != nil {
			return fmt.Errorf("persist project delete: %w", err)
		}
		if err := appendProjectEvent(ctx, h.Outbox, h.UUIDs, aggregate.DeletedEvent(now), cmd.Metadata); err != nil {
			return err
		}
		return appcommon.CompleteIdempotency(ctx, h.Idempotency, reservation)
	})
}

type UndeleteProject struct {
	Metadata appcommon.Metadata
	ID       string
}

type UndeleteProjectHandler struct {
	TxManager   ports.TxManager
	Repository  ports.ProjectRepository
	Hierarchy   ports.HierarchyRepository
	Idempotency ports.IdempotencyStore
	Outbox      ports.OutboxWriter
	Clock       ports.Clock
	UUIDs       ports.UUIDGenerator
}

func (h UndeleteProjectHandler) Handle(ctx context.Context, cmd UndeleteProject) (project.Project, error) {
	var restored project.Project
	err := h.TxManager.WithinTx(ctx, func(ctx context.Context) error {
		reservation, err := appcommon.ReserveIdempotency(ctx, h.Idempotency, "UndeleteProject:"+cmd.ID, cmd.Metadata.IdempotencyKey, defaultIdempotencyTTL)
		if err != nil {
			return err
		}
		restored, err = h.Repository.GetByID(ctx, cmd.ID, true)
		if err != nil {
			return fmt.Errorf("load project: %w", err)
		}
		parent, err := h.Hierarchy.GetWorkspaceNode(ctx, restored.WorkspaceID)
		if err != nil {
			return fmt.Errorf("load workspace parent: %w", err)
		}
		if err := hierarchy.EnsureUndeleteAllowed(parent.Exists, parent.Deleted); err != nil {
			return err
		}

		now := h.Clock.Now().UTC()
		if err := restored.Undelete(now, h.UUIDs.NewString()); err != nil {
			return err
		}
		if err := h.Repository.Update(ctx, restored); err != nil {
			return fmt.Errorf("persist project undelete: %w", err)
		}
		if err := appendProjectEvent(ctx, h.Outbox, h.UUIDs, restored.UndeletedEvent(now), cmd.Metadata); err != nil {
			return err
		}
		return appcommon.CompleteIdempotency(ctx, h.Idempotency, reservation)
	})
	return restored, err
}

func appendProjectEvent(
	ctx context.Context,
	outbox ports.OutboxWriter,
	generator ports.UUIDGenerator,
	event project.Event,
	meta appcommon.Metadata,
) error {
	if outbox == nil {
		return nil
	}
	record, err := events.ProjectRecord(generator, event, meta)
	if err != nil {
		return err
	}
	return outbox.Append(ctx, record)
}
