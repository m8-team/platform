package command

import (
	"context"
	"fmt"
	"time"

	appcommon "github.com/m8platform/platform/internal/application/common"
	"github.com/m8platform/platform/internal/domain/hierarchy"
	"github.com/m8platform/platform/internal/domain/organization"
	"github.com/m8platform/platform/internal/infra/events"
	"github.com/m8platform/platform/internal/ports"
)

const (
	defaultIdempotencyTTL = 24 * time.Hour
	defaultPurgeWindow    = 30 * 24 * time.Hour
)

type CreateOrganization struct {
	Metadata    appcommon.Metadata
	Name        string
	Description string
	Annotations map[string]string
}

type CreateOrganizationHandler struct {
	TxManager   ports.TxManager
	Repository  ports.OrganizationRepository
	Idempotency ports.IdempotencyStore
	Outbox      ports.OutboxWriter
	Clock       ports.Clock
	UUIDs       ports.UUIDGenerator
}

func (h CreateOrganizationHandler) Handle(ctx context.Context, cmd CreateOrganization) (organization.Organization, error) {
	var created organization.Organization
	err := h.TxManager.WithinTx(ctx, func(ctx context.Context) error {
		reservation, err := appcommon.ReserveIdempotency(ctx, h.Idempotency, "CreateOrganization", cmd.Metadata.IdempotencyKey, defaultIdempotencyTTL)
		if err != nil {
			return err
		}

		now := h.Clock.Now().UTC()
		created, err = organization.New(organization.CreateParams{
			ID:          h.UUIDs.NewString(),
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
			return fmt.Errorf("create organization: %w", err)
		}
		if err := appendOrganizationEvent(ctx, h.Outbox, h.UUIDs, created.CreatedEvent(now), cmd.Metadata); err != nil {
			return err
		}
		return appcommon.CompleteIdempotency(ctx, h.Idempotency, reservation)
	})
	return created, err
}

type UpdateOrganization struct {
	Metadata    appcommon.Metadata
	ID          string
	ETag        string
	UpdateMask  []string
	Name        *string
	Description *string
	Annotations map[string]string
}

type UpdateOrganizationHandler struct {
	TxManager   ports.TxManager
	Repository  ports.OrganizationRepository
	Idempotency ports.IdempotencyStore
	Outbox      ports.OutboxWriter
	Clock       ports.Clock
	UUIDs       ports.UUIDGenerator
}

func (h UpdateOrganizationHandler) Handle(ctx context.Context, cmd UpdateOrganization) (organization.Organization, error) {
	var updated organization.Organization
	err := h.TxManager.WithinTx(ctx, func(ctx context.Context) error {
		reservation, err := appcommon.ReserveIdempotency(ctx, h.Idempotency, "UpdateOrganization:"+cmd.ID, cmd.Metadata.IdempotencyKey, defaultIdempotencyTTL)
		if err != nil {
			return err
		}
		if err := appcommon.ValidateMask(cmd.UpdateMask, []string{"name", "description", "annotations"}); err != nil {
			return err
		}

		updated, err = h.Repository.GetByID(ctx, cmd.ID, true)
		if err != nil {
			return fmt.Errorf("load organization: %w", err)
		}
		now := h.Clock.Now().UTC()
		if err := updated.Update(cmd.UpdateMask, organization.UpdateFields{
			ID:          cmd.ID,
			Name:        cmd.Name,
			Description: cmd.Description,
			Annotations: cmd.Annotations,
			ETag:        cmd.ETag,
		}, now, h.UUIDs.NewString()); err != nil {
			return err
		}
		if err := h.Repository.Update(ctx, updated); err != nil {
			return fmt.Errorf("update organization: %w", err)
		}
		if err := appendOrganizationEvent(ctx, h.Outbox, h.UUIDs, updated.UpdatedEvent(now), cmd.Metadata); err != nil {
			return err
		}
		return appcommon.CompleteIdempotency(ctx, h.Idempotency, reservation)
	})
	return updated, err
}

type DeleteOrganization struct {
	Metadata     appcommon.Metadata
	ID           string
	ETag         string
	AllowMissing bool
}

type DeleteOrganizationHandler struct {
	TxManager   ports.TxManager
	Repository  ports.OrganizationRepository
	Hierarchy   ports.HierarchyRepository
	Idempotency ports.IdempotencyStore
	Outbox      ports.OutboxWriter
	Clock       ports.Clock
	UUIDs       ports.UUIDGenerator
}

func (h DeleteOrganizationHandler) Handle(ctx context.Context, cmd DeleteOrganization) error {
	return h.TxManager.WithinTx(ctx, func(ctx context.Context) error {
		reservation, err := appcommon.ReserveIdempotency(ctx, h.Idempotency, "DeleteOrganization:"+cmd.ID, cmd.Metadata.IdempotencyKey, defaultIdempotencyTTL)
		if err != nil {
			return err
		}

		aggregate, err := h.Repository.GetByID(ctx, cmd.ID, true)
		if err != nil {
			if cmd.AllowMissing && err == organization.ErrNotFound {
				return appcommon.CompleteIdempotency(ctx, h.Idempotency, reservation)
			}
			return fmt.Errorf("load organization: %w", err)
		}

		hasChildren, err := h.Hierarchy.HasActiveWorkspaces(ctx, cmd.ID)
		if err != nil {
			return fmt.Errorf("check active workspaces: %w", err)
		}
		if err := hierarchy.EnsureDeleteAllowed(hasChildren); err != nil {
			return err
		}

		now := h.Clock.Now().UTC()
		if err := aggregate.SoftDelete(now, now.Add(defaultPurgeWindow), cmd.ETag, h.UUIDs.NewString()); err != nil {
			return err
		}
		if err := h.Repository.Update(ctx, aggregate); err != nil {
			return fmt.Errorf("persist organization delete: %w", err)
		}
		if err := appendOrganizationEvent(ctx, h.Outbox, h.UUIDs, aggregate.DeletedEvent(now), cmd.Metadata); err != nil {
			return err
		}
		return appcommon.CompleteIdempotency(ctx, h.Idempotency, reservation)
	})
}

type UndeleteOrganization struct {
	Metadata appcommon.Metadata
	ID       string
}

type UndeleteOrganizationHandler struct {
	TxManager   ports.TxManager
	Repository  ports.OrganizationRepository
	Idempotency ports.IdempotencyStore
	Outbox      ports.OutboxWriter
	Clock       ports.Clock
	UUIDs       ports.UUIDGenerator
}

func (h UndeleteOrganizationHandler) Handle(ctx context.Context, cmd UndeleteOrganization) (organization.Organization, error) {
	var restored organization.Organization
	err := h.TxManager.WithinTx(ctx, func(ctx context.Context) error {
		reservation, err := appcommon.ReserveIdempotency(ctx, h.Idempotency, "UndeleteOrganization:"+cmd.ID, cmd.Metadata.IdempotencyKey, defaultIdempotencyTTL)
		if err != nil {
			return err
		}

		restored, err = h.Repository.GetByID(ctx, cmd.ID, true)
		if err != nil {
			return fmt.Errorf("load organization: %w", err)
		}
		now := h.Clock.Now().UTC()
		if err := restored.Undelete(now, h.UUIDs.NewString()); err != nil {
			return err
		}
		if err := h.Repository.Update(ctx, restored); err != nil {
			return fmt.Errorf("persist organization undelete: %w", err)
		}
		if err := appendOrganizationEvent(ctx, h.Outbox, h.UUIDs, restored.UndeletedEvent(now), cmd.Metadata); err != nil {
			return err
		}
		return appcommon.CompleteIdempotency(ctx, h.Idempotency, reservation)
	})
	return restored, err
}

func appendOrganizationEvent(
	ctx context.Context,
	outbox ports.OutboxWriter,
	generator ports.UUIDGenerator,
	event organization.Event,
	meta appcommon.Metadata,
) error {
	if outbox == nil {
		return nil
	}
	record, err := events.OrganizationRecord(generator, event, meta)
	if err != nil {
		return err
	}
	return outbox.Append(ctx, record)
}
