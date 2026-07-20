package ports

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/m8-team/platform/internal/platform/types"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
)

const (
	DefaultWorkspacePageSize = 50
	MaxWorkspacePageSize     = 1000
)

var (
	ErrNilWorkspace                   = errors.New("workspace is required")
	ErrWorkspaceNotFound              = errors.New("workspace not found")
	ErrWorkspaceAlreadyExists         = errors.New("workspace already exists")
	ErrWorkspaceVersionConflict       = errors.New("workspace version conflict")
	ErrWorkspaceRepositoryUnavailable = errors.New("workspace repository is unavailable")
	ErrInvalidWorkspaceVersion        = errors.New("invalid workspace version")
	ErrInvalidListWorkspacesOptions   = errors.New("invalid list workspaces options")
)

// WorkspaceRepository is the persistence port consumed by workspace
// application use cases. Implementations must return detached aggregates so a
// caller cannot mutate stored state without an explicit Update call.
type WorkspaceRepository interface {
	Create(ctx context.Context, value *workspace.Workspace) error
	Get(ctx context.Context, id workspace.ID) (*workspace.Workspace, error)
	Update(ctx context.Context, value *workspace.Workspace, expectedVersion types.Version) error
	List(ctx context.Context, options ListWorkspacesOptions) (ListWorkspacesResult, error)
	HasNonDeleted(ctx context.Context, organizationID organization.ID) (bool, error)
	WithOrganizationLock(ctx context.Context, organizationID organization.ID, fn func(context.Context) error) error
}

type OrganizationLookup interface {
	Get(context.Context, organization.ID) (*organization.Organization, error)
}

// WorkspaceVersionConflictError reports a failed compare-and-swap update.
type WorkspaceVersionConflictError struct {
	ID       workspace.ID
	Expected types.Version
	Actual   types.Version
}

func (e *WorkspaceVersionConflictError) Error() string {
	return fmt.Sprintf(
		"%v: workspace_id=%s expected=%s actual=%s",
		ErrWorkspaceVersionConflict,
		e.ID,
		e.Expected,
		e.Actual,
	)
}

func (e *WorkspaceVersionConflictError) Unwrap() error {
	return ErrWorkspaceVersionConflict
}

type WorkspaceOrderField string

const (
	WorkspaceOrderFieldID         WorkspaceOrderField = "id"
	WorkspaceOrderFieldName       WorkspaceOrderField = "name"
	WorkspaceOrderFieldCreateTime WorkspaceOrderField = "create_time"
	WorkspaceOrderFieldUpdateTime WorkspaceOrderField = "update_time"
)

func (f WorkspaceOrderField) IsValid() bool {
	switch f {
	case WorkspaceOrderFieldID,
		WorkspaceOrderFieldName,
		WorkspaceOrderFieldCreateTime,
		WorkspaceOrderFieldUpdateTime:
		return true
	default:
		return false
	}
}

type WorkspaceFilter struct {
	States      []workspace.State
	NameEquals  *string
	LabelsEqual map[string]string
	ShowDeleted bool
}

// WorkspaceOrder defines one explicit sort field. Workspace ID is an
// implicit final tie-breaker in the same direction.
type WorkspaceOrder struct {
	Field     WorkspaceOrderField
	Direction SortDirection
}

// WorkspaceListCursor is a transport-neutral keyset cursor. The
// application/transport layer is responsible for encoding and authenticating
// it as an opaque page token.
type WorkspaceListCursor struct {
	ID         workspace.ID
	Name       string
	CreateTime time.Time
	UpdateTime time.Time
}

type ListWorkspacesOptions struct {
	OrganizationID organization.ID
	Filter         WorkspaceFilter
	Order          WorkspaceOrder
	PageSize       int
	After          *WorkspaceListCursor
}

func (o ListWorkspacesOptions) WithDefaults() ListWorkspacesOptions {
	if o.PageSize == 0 {
		o.PageSize = DefaultWorkspacePageSize
	}
	if o.Order.Field == "" {
		o.Order.Field = WorkspaceOrderFieldID
	}
	if o.Order.Direction == "" {
		o.Order.Direction = SortDirectionAscending
	}

	return o
}

func (o ListWorkspacesOptions) Validate() error {
	o = o.WithDefaults()

	if err := o.OrganizationID.Validate(); err != nil {
		return fmt.Errorf("%w: organization id: %v", ErrInvalidListWorkspacesOptions, err)
	}
	if o.PageSize < 1 || o.PageSize > MaxWorkspacePageSize {
		return fmt.Errorf(
			"%w: page size %d must be between 1 and %d",
			ErrInvalidListWorkspacesOptions,
			o.PageSize,
			MaxWorkspacePageSize,
		)
	}
	if !o.Order.Field.IsValid() {
		return fmt.Errorf("%w: unsupported order field %q", ErrInvalidListWorkspacesOptions, o.Order.Field)
	}
	if !o.Order.Direction.IsValid() {
		return fmt.Errorf("%w: unsupported sort direction %q", ErrInvalidListWorkspacesOptions, o.Order.Direction)
	}
	for _, state := range o.Filter.States {
		if !state.IsValid() || state == workspace.StateUnspecified {
			return fmt.Errorf("%w: unsupported workspace state %q", ErrInvalidListWorkspacesOptions, state)
		}
	}
	if o.After != nil {
		if err := o.After.ID.Validate(); err != nil {
			return fmt.Errorf("%w: cursor workspace id: %v", ErrInvalidListWorkspacesOptions, err)
		}
		switch o.Order.Field {
		case WorkspaceOrderFieldCreateTime:
			if o.After.CreateTime.IsZero() {
				return fmt.Errorf("%w: cursor create_time is required", ErrInvalidListWorkspacesOptions)
			}
		case WorkspaceOrderFieldUpdateTime:
			if o.After.UpdateTime.IsZero() {
				return fmt.Errorf("%w: cursor update_time is required", ErrInvalidListWorkspacesOptions)
			}
		}
	}

	return nil
}

type ListWorkspacesResult struct {
	Workspaces []*workspace.Workspace
	Next       *WorkspaceListCursor
	TotalSize  int
}
