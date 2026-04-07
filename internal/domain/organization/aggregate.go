package organization

import (
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
)

var allowedUpdatePaths = []string{"name", "description", "annotations"}

type Organization struct {
	ID          string
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
	Name        string
	Description string
	Annotations map[string]string
	Now         time.Time
	ETag        string
}

type UpdateFields struct {
	ID          string
	Name        *string
	Description *string
	Annotations map[string]string
	ETag        string
}

func New(params CreateParams) (Organization, error) {
	if _, err := uuid.Parse(params.ID); err != nil {
		return Organization{}, ErrInvalidID
	}
	if params.Now.IsZero() {
		params.Now = time.Now().UTC()
	}

	return Organization{
		ID:          params.ID,
		State:       StateActive,
		Name:        params.Name,
		Description: params.Description,
		CreateTime:  params.Now,
		UpdateTime:  params.Now,
		ETag:        params.ETag,
		Annotations: cloneMap(params.Annotations),
	}, nil
}

func (o Organization) IsDeleted() bool {
	return o.State == StateDeleted
}

func (o Organization) EnsureETag(expected string) error {
	if expected == "" || o.ETag == "" {
		return nil
	}
	if o.ETag != expected {
		return ErrETagMismatch
	}
	return nil
}

func (o *Organization) Update(mask []string, fields UpdateFields, now time.Time, nextETag string) error {
	if o == nil {
		return ErrNotFound
	}
	if fields.ID != "" && fields.ID != o.ID {
		return ErrImmutableID
	}
	if o.IsDeleted() {
		return ErrDeleted
	}
	if err := o.EnsureETag(fields.ETag); err != nil {
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
				o.Name = *fields.Name
			}
		case "description":
			if fields.Description != nil {
				o.Description = *fields.Description
			}
		case "annotations":
			o.Annotations = cloneMap(fields.Annotations)
		}
	}
	o.UpdateTime = now.UTC()
	o.ETag = nextETag
	return nil
}

func (o *Organization) SoftDelete(now time.Time, purgeAt time.Time, expectedETag string, nextETag string) error {
	if o == nil {
		return ErrNotFound
	}
	if err := o.EnsureETag(expectedETag); err != nil {
		return err
	}
	if o.IsDeleted() {
		return ErrAlreadyDeleted
	}
	if !purgeAt.After(now) {
		return ErrInvalidPurgeWindow
	}

	deleteTime := now.UTC()
	purgeTime := purgeAt.UTC()
	o.State = StateDeleted
	o.DeleteTime = &deleteTime
	o.PurgeTime = &purgeTime
	o.UpdateTime = deleteTime
	o.ETag = nextETag
	return nil
}

func (o *Organization) Undelete(now time.Time, nextETag string) error {
	if o == nil {
		return ErrNotFound
	}
	if !o.IsDeleted() {
		return ErrNotDeleted
	}
	o.State = StateActive
	o.DeleteTime = nil
	o.PurgeTime = nil
	o.UpdateTime = now.UTC()
	o.ETag = nextETag
	return nil
}

func (o Organization) CreatedEvent(at time.Time) Event {
	return Event{Type: EventCreated, Aggregate: o, OccurredAt: at}
}

func (o Organization) UpdatedEvent(at time.Time) Event {
	return Event{Type: EventUpdated, Aggregate: o, OccurredAt: at}
}

func (o Organization) DeletedEvent(at time.Time) Event {
	return Event{Type: EventDeleted, Aggregate: o, OccurredAt: at}
}

func (o Organization) UndeletedEvent(at time.Time) Event {
	return Event{Type: EventUndeleted, Aggregate: o, OccurredAt: at}
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
