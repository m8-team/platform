package organization

import (
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/m8platform/platform/internal/entity/resourcemanager/shared"
)

type Entity struct {
	ID          string
	State       State
	Name        string
	Description string
	CreateTime  time.Time
	UpdateTime  time.Time
	DeleteTime  *time.Time
	PurgeTime   *time.Time
	ETag        shared.ETag
	Annotations shared.Metadata
}

type CreateParams struct {
	ID          string
	Name        string
	Description string
	Annotations map[string]string
	Now         time.Time
	ETag        string
}

type UpdateParams struct {
	ID          string
	Name        *string
	Description *string
	Annotations map[string]string
	ETag        string
}

func New(params CreateParams) (Entity, error) {
	if _, err := uuid.Parse(params.ID); err != nil {
		return Entity{}, ErrInvalidID
	}
	if params.Now.IsZero() {
		params.Now = time.Now().UTC()
	}
	return Entity{
		ID:          params.ID,
		State:       StateActive,
		Name:        params.Name,
		Description: params.Description,
		CreateTime:  params.Now.UTC(),
		UpdateTime:  params.Now.UTC(),
		ETag:        shared.ETag(params.ETag),
		Annotations: shared.CloneMetadata(params.Annotations),
	}, nil
}

func (e Entity) IsDeleted() bool {
	return e.State == StateDeleted
}

func (e Entity) EnsureETag(expected string) error {
	if expected == "" || e.ETag == "" {
		return nil
	}
	if e.ETag.String() != expected {
		return ErrETagMismatch
	}
	return nil
}

func (e *Entity) Update(mask []string, params UpdateParams, now time.Time, nextETag string) error {
	if params.ID != "" && params.ID != e.ID {
		return ErrImmutableID
	}
	if e.IsDeleted() {
		return ErrDeleted
	}
	if err := e.EnsureETag(params.ETag); err != nil {
		return err
	}
	for _, path := range mask {
		if !slices.Contains(AllowedUpdatePaths, path) {
			return fmt.Errorf("%w: %s", ErrInvalidUpdatePath, path)
		}
	}
	for _, path := range mask {
		switch path {
		case "name":
			if params.Name != nil {
				e.Name = *params.Name
			}
		case "description":
			if params.Description != nil {
				e.Description = *params.Description
			}
		case "annotations":
			e.Annotations = shared.CloneMetadata(params.Annotations)
		}
	}
	e.UpdateTime = now.UTC()
	e.ETag = shared.ETag(nextETag)
	return nil
}

func (e *Entity) SoftDelete(now time.Time, purgeAt time.Time, expectedETag string, nextETag string) error {
	if err := e.EnsureETag(expectedETag); err != nil {
		return err
	}
	if e.IsDeleted() {
		return ErrAlreadyDeleted
	}
	if !purgeAt.After(now) {
		return ErrInvalidPurgeWindow
	}
	deleteTime := now.UTC()
	purgeTime := purgeAt.UTC()
	e.State = StateDeleted
	e.DeleteTime = &deleteTime
	e.PurgeTime = &purgeTime
	e.UpdateTime = deleteTime
	e.ETag = shared.ETag(nextETag)
	return nil
}

func (e *Entity) Undelete(now time.Time, nextETag string) error {
	if !e.IsDeleted() {
		return ErrNotDeleted
	}
	e.State = StateActive
	e.DeleteTime = nil
	e.PurgeTime = nil
	e.UpdateTime = now.UTC()
	e.ETag = shared.ETag(nextETag)
	return nil
}
