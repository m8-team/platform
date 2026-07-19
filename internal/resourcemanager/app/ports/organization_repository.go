package ports

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/m8-team/platform/internal/platform/types"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
)

const (
	DefaultOrganizationPageSize = 50
	MaxOrganizationPageSize     = 1000
)

var (
	ErrNilOrganization                   = errors.New("organization is required")
	ErrOrganizationNotFound              = errors.New("organization not found")
	ErrOrganizationAlreadyExists         = errors.New("organization already exists")
	ErrOrganizationVersionConflict       = errors.New("organization version conflict")
	ErrOrganizationRepositoryUnavailable = errors.New("organization repository is unavailable")
	ErrInvalidOrganizationVersion        = errors.New("invalid organization version")
	ErrInvalidListOrganizationsOptions   = errors.New("invalid list organizations options")
)

// OrganizationRepository is the persistence port consumed by organization
// application use cases. Implementations must return detached aggregates so a
// caller cannot mutate stored state without an explicit Update call.
type OrganizationRepository interface {
	Create(ctx context.Context, value *organization.Organization) error
	Get(ctx context.Context, id organization.ID) (*organization.Organization, error)
	Update(ctx context.Context, value *organization.Organization, expectedVersion types.Version) error
	List(ctx context.Context, options ListOrganizationsOptions) (ListOrganizationsResult, error)
}

// OrganizationVersionConflictError reports a failed compare-and-swap update.
type OrganizationVersionConflictError struct {
	ID       organization.ID
	Expected types.Version
	Actual   types.Version
}

func (e *OrganizationVersionConflictError) Error() string {
	return fmt.Sprintf(
		"%v: organization_id=%s expected=%s actual=%s",
		ErrOrganizationVersionConflict,
		e.ID,
		e.Expected,
		e.Actual,
	)
}

func (e *OrganizationVersionConflictError) Unwrap() error {
	return ErrOrganizationVersionConflict
}

type OrganizationOrderField string

const (
	OrganizationOrderFieldID         OrganizationOrderField = "id"
	OrganizationOrderFieldName       OrganizationOrderField = "name"
	OrganizationOrderFieldCreateTime OrganizationOrderField = "create_time"
	OrganizationOrderFieldUpdateTime OrganizationOrderField = "update_time"
)

func (f OrganizationOrderField) IsValid() bool {
	switch f {
	case OrganizationOrderFieldID,
		OrganizationOrderFieldName,
		OrganizationOrderFieldCreateTime,
		OrganizationOrderFieldUpdateTime:
		return true
	default:
		return false
	}
}

type SortDirection string

const (
	SortDirectionAscending  SortDirection = "asc"
	SortDirectionDescending SortDirection = "desc"
)

func (d SortDirection) IsValid() bool {
	return d == SortDirectionAscending || d == SortDirectionDescending
}

type OrganizationFilter struct {
	States      []organization.State
	NameEquals  *string
	LabelsEqual map[string]string
	ShowDeleted bool
}

// OrganizationOrder defines one explicit sort field. Organization ID is an
// implicit final tie-breaker in the same direction.
type OrganizationOrder struct {
	Field     OrganizationOrderField
	Direction SortDirection
}

// OrganizationListCursor is a transport-neutral keyset cursor. The
// application/transport layer is responsible for encoding and authenticating
// it as an opaque page token.
type OrganizationListCursor struct {
	ID         organization.ID
	Name       string
	CreateTime time.Time
	UpdateTime time.Time
}

type ListOrganizationsOptions struct {
	Filter   OrganizationFilter
	Order    OrganizationOrder
	PageSize int
	After    *OrganizationListCursor
}

func (o ListOrganizationsOptions) WithDefaults() ListOrganizationsOptions {
	if o.PageSize == 0 {
		o.PageSize = DefaultOrganizationPageSize
	}
	if o.Order.Field == "" {
		o.Order.Field = OrganizationOrderFieldID
	}
	if o.Order.Direction == "" {
		o.Order.Direction = SortDirectionAscending
	}

	return o
}

func (o ListOrganizationsOptions) Validate() error {
	o = o.WithDefaults()

	if o.PageSize < 1 || o.PageSize > MaxOrganizationPageSize {
		return fmt.Errorf(
			"%w: page size %d must be between 1 and %d",
			ErrInvalidListOrganizationsOptions,
			o.PageSize,
			MaxOrganizationPageSize,
		)
	}
	if !o.Order.Field.IsValid() {
		return fmt.Errorf("%w: unsupported order field %q", ErrInvalidListOrganizationsOptions, o.Order.Field)
	}
	if !o.Order.Direction.IsValid() {
		return fmt.Errorf("%w: unsupported sort direction %q", ErrInvalidListOrganizationsOptions, o.Order.Direction)
	}
	for _, state := range o.Filter.States {
		if !state.IsValid() || state == organization.StateUnspecified {
			return fmt.Errorf("%w: unsupported organization state %q", ErrInvalidListOrganizationsOptions, state)
		}
	}
	if o.After != nil {
		if err := o.After.ID.Validate(); err != nil {
			return fmt.Errorf("%w: cursor organization id: %v", ErrInvalidListOrganizationsOptions, err)
		}
		switch o.Order.Field {
		case OrganizationOrderFieldCreateTime:
			if o.After.CreateTime.IsZero() {
				return fmt.Errorf("%w: cursor create_time is required", ErrInvalidListOrganizationsOptions)
			}
		case OrganizationOrderFieldUpdateTime:
			if o.After.UpdateTime.IsZero() {
				return fmt.Errorf("%w: cursor update_time is required", ErrInvalidListOrganizationsOptions)
			}
		}
	}

	return nil
}

type ListOrganizationsResult struct {
	Organizations []*organization.Organization
	Next          *OrganizationListCursor
	TotalSize     int
}
