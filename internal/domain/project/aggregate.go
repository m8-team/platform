package project

import (
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
)

var allowedUpdatePaths = []string{"name", "description", "annotations"}

type Project struct {
	ID          string
	WorkspaceID string
	State       State
	Name        string
	Description string
	CreateTime  time.Time
	UpdateTime  time.Time
	DeleteTime  *time.Time
	PurgeTime   *time.Time
	ETag        string
	Annotations map[string]string
}

type CreateParams struct {
	ID          string
	WorkspaceID string
	Name        string
	Description string
	Annotations map[string]string
	Now         time.Time
	ETag        string
}

type UpdateFields struct {
	ID          string
	WorkspaceID string
	Name        *string
	Description *string
	Annotations map[string]string
	ETag        string
}

func New(params CreateParams) (Project, error) {
	if _, err := uuid.Parse(params.ID); err != nil {
		return Project{}, ErrInvalidID
	}
	if _, err := uuid.Parse(params.WorkspaceID); err != nil {
		return Project{}, ErrInvalidParentID
	}
	if params.Now.IsZero() {
		params.Now = time.Now().UTC()
	}

	return Project{
		ID:          params.ID,
		WorkspaceID: params.WorkspaceID,
		State:       StateActive,
		Name:        params.Name,
		Description: params.Description,
		CreateTime:  params.Now,
		UpdateTime:  params.Now,
		ETag:        params.ETag,
		Annotations: cloneMap(params.Annotations),
	}, nil
}

func (p Project) IsDeleted() bool {
	return p.State == StateDeleted
}

func (p Project) EnsureETag(expected string) error {
	if expected == "" || p.ETag == "" {
		return nil
	}
	if p.ETag != expected {
		return ErrETagMismatch
	}
	return nil
}

func (p *Project) Update(mask []string, fields UpdateFields, now time.Time, nextETag string) error {
	if p == nil {
		return ErrNotFound
	}
	if fields.ID != "" && fields.ID != p.ID {
		return ErrImmutableID
	}
	if p.IsDeleted() {
		return ErrDeleted
	}
	if fields.WorkspaceID != "" && fields.WorkspaceID != p.WorkspaceID {
		return ErrImmutableParent
	}
	if err := p.EnsureETag(fields.ETag); err != nil {
		return err
	}
	for _, path := range mask {
		if !slices.Contains(allowedUpdatePaths, path) {
			return fmt.Errorf("%w: %s", ErrInvalidUpdatePath, path)
		}
	}
	for _, path := range mask {
		switch path {
		case "name":
			if fields.Name != nil {
				p.Name = *fields.Name
			}
		case "description":
			if fields.Description != nil {
				p.Description = *fields.Description
			}
		case "annotations":
			p.Annotations = cloneMap(fields.Annotations)
		}
	}
	p.UpdateTime = now.UTC()
	p.ETag = nextETag
	return nil
}

func (p *Project) SoftDelete(now time.Time, purgeAt time.Time, expectedETag string, nextETag string) error {
	if p == nil {
		return ErrNotFound
	}
	if err := p.EnsureETag(expectedETag); err != nil {
		return err
	}
	if p.IsDeleted() {
		return ErrAlreadyDeleted
	}
	if !purgeAt.After(now) {
		return ErrInvalidPurgeWindow
	}

	deleteTime := now.UTC()
	purgeTime := purgeAt.UTC()
	p.State = StateDeleted
	p.DeleteTime = &deleteTime
	p.PurgeTime = &purgeTime
	p.UpdateTime = deleteTime
	p.ETag = nextETag
	return nil
}

func (p *Project) Undelete(now time.Time, nextETag string) error {
	if p == nil {
		return ErrNotFound
	}
	if !p.IsDeleted() {
		return ErrNotDeleted
	}
	p.State = StateActive
	p.DeleteTime = nil
	p.PurgeTime = nil
	p.UpdateTime = now.UTC()
	p.ETag = nextETag
	return nil
}

func (p Project) CreatedEvent(at time.Time) Event {
	return Event{Type: EventCreated, Aggregate: p, OccurredAt: at}
}

func (p Project) UpdatedEvent(at time.Time) Event {
	return Event{Type: EventUpdated, Aggregate: p, OccurredAt: at}
}

func (p Project) DeletedEvent(at time.Time) Event {
	return Event{Type: EventDeleted, Aggregate: p, OccurredAt: at}
}

func (p Project) UndeletedEvent(at time.Time) Event {
	return Event{Type: EventUndeleted, Aggregate: p, OccurredAt: at}
}

func cloneMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	out := make(map[string]string, len(input))
	for k, v := range input {
		out[k] = v
	}
	return out
}
