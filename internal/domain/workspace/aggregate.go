package workspace

import (
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
)

var allowedUpdatePaths = []string{"name", "description", "annotations"}

type Workspace struct {
	ID             string
	OrganizationID string
	State          State
	Name           string
	Description    string
	CreateTime     time.Time
	UpdateTime     time.Time
	DeleteTime     *time.Time
	PurgeTime      *time.Time
	ETag           string
	Annotations    map[string]string
}

type CreateParams struct {
	ID             string
	OrganizationID string
	Name           string
	Description    string
	Annotations    map[string]string
	Now            time.Time
	ETag           string
}

type UpdateFields struct {
	ID             string
	OrganizationID string
	Name           *string
	Description    *string
	Annotations    map[string]string
	ETag           string
}

func New(params CreateParams) (Workspace, error) {
	if _, err := uuid.Parse(params.ID); err != nil {
		return Workspace{}, ErrInvalidID
	}
	if _, err := uuid.Parse(params.OrganizationID); err != nil {
		return Workspace{}, ErrInvalidParentID
	}
	if params.Now.IsZero() {
		params.Now = time.Now().UTC()
	}

	return Workspace{
		ID:             params.ID,
		OrganizationID: params.OrganizationID,
		State:          StateActive,
		Name:           params.Name,
		Description:    params.Description,
		CreateTime:     params.Now,
		UpdateTime:     params.Now,
		ETag:           params.ETag,
		Annotations:    cloneMap(params.Annotations),
	}, nil
}

func (w Workspace) IsDeleted() bool {
	return w.State == StateDeleted
}

func (w Workspace) EnsureETag(expected string) error {
	if expected == "" || w.ETag == "" {
		return nil
	}
	if w.ETag != expected {
		return ErrETagMismatch
	}
	return nil
}

func (w *Workspace) Update(mask []string, fields UpdateFields, now time.Time, nextETag string) error {
	if w == nil {
		return ErrNotFound
	}
	if fields.ID != "" && fields.ID != w.ID {
		return ErrImmutableID
	}
	if w.IsDeleted() {
		return ErrDeleted
	}
	if fields.OrganizationID != "" && fields.OrganizationID != w.OrganizationID {
		return ErrImmutableParent
	}
	if err := w.EnsureETag(fields.ETag); err != nil {
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
				w.Name = *fields.Name
			}
		case "description":
			if fields.Description != nil {
				w.Description = *fields.Description
			}
		case "annotations":
			w.Annotations = cloneMap(fields.Annotations)
		}
	}
	w.UpdateTime = now.UTC()
	w.ETag = nextETag
	return nil
}

func (w *Workspace) SoftDelete(now time.Time, purgeAt time.Time, expectedETag string, nextETag string) error {
	if w == nil {
		return ErrNotFound
	}
	if err := w.EnsureETag(expectedETag); err != nil {
		return err
	}
	if w.IsDeleted() {
		return ErrAlreadyDeleted
	}
	if !purgeAt.After(now) {
		return ErrInvalidPurgeWindow
	}

	deleteTime := now.UTC()
	purgeTime := purgeAt.UTC()
	w.State = StateDeleted
	w.DeleteTime = &deleteTime
	w.PurgeTime = &purgeTime
	w.UpdateTime = deleteTime
	w.ETag = nextETag
	return nil
}

func (w *Workspace) Undelete(now time.Time, nextETag string) error {
	if w == nil {
		return ErrNotFound
	}
	if !w.IsDeleted() {
		return ErrNotDeleted
	}
	w.State = StateActive
	w.DeleteTime = nil
	w.PurgeTime = nil
	w.UpdateTime = now.UTC()
	w.ETag = nextETag
	return nil
}

func (w Workspace) CreatedEvent(at time.Time) Event {
	return Event{Type: EventCreated, Aggregate: w, OccurredAt: at}
}

func (w Workspace) UpdatedEvent(at time.Time) Event {
	return Event{Type: EventUpdated, Aggregate: w, OccurredAt: at}
}

func (w Workspace) DeletedEvent(at time.Time) Event {
	return Event{Type: EventDeleted, Aggregate: w, OccurredAt: at}
}

func (w Workspace) UndeletedEvent(at time.Time) Event {
	return Event{Type: EventUndeleted, Aggregate: w, OccurredAt: at}
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
