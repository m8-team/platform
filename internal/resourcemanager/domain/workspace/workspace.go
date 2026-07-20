package workspace

import (
	"errors"
	"fmt"
	"time"

	"github.com/m8-team/platform/internal/platform/types"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
)

const ResourceType = "resourcemanager.workspace"

type ID = organization.ID
type State = organization.State

const (
	StateCreating  = organization.StateCreating
	StateActive    = organization.StateActive
	StateSuspended = organization.StateSuspended
	StateDeleting  = organization.StateDeleting
	StateDeleted   = organization.StateDeleted
	StateFailed    = organization.StateFailed
)

var (
	ErrNilWorkspace        = errors.New("workspace is nil")
	ErrWorkspaceNotDeleted = errors.New("workspace is not deleted")
)

type Workspace struct {
	organizationID organization.ID
	resource       *organization.Organization
}

type CreateParams struct {
	ID                ID
	OrganizationID    organization.ID
	Name, Description string
	Labels            map[string]string
	Now               time.Time
}

func New(p CreateParams) (*Workspace, error) {
	if err := p.OrganizationID.Validate(); err != nil {
		return nil, fmt.Errorf("organization id: %w", err)
	}
	r, err := organization.New(organization.CreateParams{ID: p.ID, Name: p.Name, Description: p.Description, Labels: p.Labels, Now: p.Now})
	if err != nil {
		return nil, err
	}
	return &Workspace{organizationID: p.OrganizationID, resource: r}, nil
}

func (w *Workspace) Clone() *Workspace {
	if w == nil {
		return nil
	}
	return &Workspace{organizationID: w.organizationID, resource: w.resource.Clone()}
}

func (w *Workspace) ID() ID                          { return w.resource.ID() }
func (w *Workspace) OrganizationID() organization.ID { return w.organizationID }
func (w *Workspace) State() State                    { return w.resource.State() }
func (w *Workspace) Name() string                    { return w.resource.Name() }
func (w *Workspace) Description() string             { return w.resource.Description() }
func (w *Workspace) CreateTime() time.Time           { return w.resource.CreateTime() }
func (w *Workspace) UpdateTime() time.Time           { return w.resource.UpdateTime() }
func (w *Workspace) DeleteTime() *time.Time          { return w.resource.DeleteTime() }
func (w *Workspace) PurgeTime() *time.Time           { return w.resource.PurgeTime() }
func (w *Workspace) Version() types.Version          { return w.resource.Version() }
func (w *Workspace) Labels() map[string]string       { return w.resource.Labels() }
func (w *Workspace) IsDeleted() bool                 { return w.resource.IsDeleted() }

func (w *Workspace) Update(name, description *string, labels *map[string]string, now time.Time, version types.Version) error {
	if w == nil {
		return ErrNilWorkspace
	}
	return w.resource.Update(organization.UpdateParams{Name: name, Description: description, Labels: labels, Now: now, ExpectedVersion: version})
}
func (w *Workspace) Delete(now, purge time.Time, version types.Version) error {
	if w == nil {
		return ErrNilWorkspace
	}
	return w.resource.Delete(organization.DeleteParams{Now: now, PurgeTime: purge, ExpectedVersion: version})
}
func (w *Workspace) Undelete(now time.Time) error {
	if w == nil {
		return ErrNilWorkspace
	}
	return w.resource.Undelete(organization.UndeleteParams{Now: now, ExpectedVersion: w.resource.Version()})
}
