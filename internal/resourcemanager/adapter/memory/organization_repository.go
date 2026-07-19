package memory

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/m8-team/platform/internal/platform/types"
	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
)

var _ ports.OrganizationRepository = (*OrganizationRepository)(nil)

// OrganizationRepository is a process-local repository intended for tests and
// local development. It preserves repository semantics, including detached
// reads and atomic compare-and-swap updates, but it is not durable storage.
type OrganizationRepository struct {
	mu            sync.RWMutex
	organizations map[organization.ID]*organization.Organization
}

func NewOrganizationRepository() *OrganizationRepository {
	return &OrganizationRepository{
		organizations: make(map[organization.ID]*organization.Organization),
	}
}

func (r *OrganizationRepository) Create(ctx context.Context, value *organization.Organization) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	if value == nil {
		return ports.ErrNilOrganization
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if err := contextError(ctx); err != nil {
		return err
	}
	if r.organizations == nil {
		r.organizations = make(map[organization.ID]*organization.Organization)
	}
	if _, exists := r.organizations[value.ID()]; exists {
		return fmt.Errorf("%w: organization_id=%s", ports.ErrOrganizationAlreadyExists, value.ID())
	}

	r.organizations[value.ID()] = value.Clone()
	return nil
}

func (r *OrganizationRepository) Get(
	ctx context.Context,
	id organization.ID,
) (*organization.Organization, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := contextError(ctx); err != nil {
		return nil, err
	}
	value, exists := r.organizations[id]
	if !exists {
		return nil, fmt.Errorf("%w: organization_id=%s", ports.ErrOrganizationNotFound, id)
	}

	return value.Clone(), nil
}

func (r *OrganizationRepository) Update(
	ctx context.Context,
	value *organization.Organization,
	expectedVersion types.Version,
) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	if value == nil {
		return ports.ErrNilOrganization
	}
	if err := expectedVersion.Validate(); err != nil {
		return fmt.Errorf("expected organization version: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if err := contextError(ctx); err != nil {
		return err
	}
	current, exists := r.organizations[value.ID()]
	if !exists {
		return fmt.Errorf("%w: organization_id=%s", ports.ErrOrganizationNotFound, value.ID())
	}
	if !current.Version().Equal(expectedVersion) {
		return &ports.OrganizationVersionConflictError{
			ID:       value.ID(),
			Expected: expectedVersion,
			Actual:   current.Version(),
		}
	}
	nextVersion, err := expectedVersion.Next()
	if err != nil {
		return fmt.Errorf("next organization version: %w", err)
	}
	if !value.Version().Equal(nextVersion) {
		return fmt.Errorf(
			"%w: organization_id=%s expected_next=%s incoming=%s",
			ports.ErrInvalidOrganizationVersion,
			value.ID(),
			nextVersion,
			value.Version(),
		)
	}

	r.organizations[value.ID()] = value.Clone()
	return nil
}

func (r *OrganizationRepository) List(
	ctx context.Context,
	options ports.ListOrganizationsOptions,
) (ports.ListOrganizationsResult, error) {
	if err := contextError(ctx); err != nil {
		return ports.ListOrganizationsResult{}, err
	}

	options = options.WithDefaults()
	if err := options.Validate(); err != nil {
		return ports.ListOrganizationsResult{}, err
	}

	r.mu.RLock()
	values := make([]*organization.Organization, 0, len(r.organizations))
	for _, value := range r.organizations {
		if matchesOrganization(value, options.Filter) {
			values = append(values, value.Clone())
		}
	}
	r.mu.RUnlock()

	if err := contextError(ctx); err != nil {
		return ports.ListOrganizationsResult{}, err
	}

	sort.Slice(values, func(i, j int) bool {
		return compareOrganizations(values[i], values[j], options.Order) < 0
	})

	totalSize := len(values)
	start := 0
	if options.After != nil {
		start = sort.Search(len(values), func(i int) bool {
			return compareOrganizationToCursor(values[i], *options.After, options.Order) > 0
		})
	}

	end := min(start+options.PageSize, len(values))
	page := values[start:end]

	var next *ports.OrganizationListCursor
	if end < len(values) && len(page) > 0 {
		cursor := newOrganizationCursor(page[len(page)-1])
		next = &cursor
	}

	return ports.ListOrganizationsResult{
		Organizations: page,
		Next:          next,
		TotalSize:     totalSize,
	}, nil
}

func contextError(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

func matchesOrganization(value *organization.Organization, filter ports.OrganizationFilter) bool {
	if !filter.ShowDeleted && value.State() == organization.StateDeleted {
		return false
	}
	if len(filter.States) > 0 && !containsState(filter.States, value.State()) {
		return false
	}
	if filter.NameEquals != nil && value.Name() != *filter.NameEquals {
		return false
	}

	labels := value.Labels()
	for key, expected := range filter.LabelsEqual {
		if actual, exists := labels[key]; !exists || actual != expected {
			return false
		}
	}

	return true
}

func containsState(states []organization.State, target organization.State) bool {
	for _, state := range states {
		if state == target {
			return true
		}
	}
	return false
}

func compareOrganizations(
	left *organization.Organization,
	right *organization.Organization,
	order ports.OrganizationOrder,
) int {
	result := compareOrganizationValues(
		left.ID(),
		left.Name(),
		left.CreateTime(),
		left.UpdateTime(),
		right.ID(),
		right.Name(),
		right.CreateTime(),
		right.UpdateTime(),
		order.Field,
	)
	if order.Direction == ports.SortDirectionDescending {
		return -result
	}
	return result
}

func compareOrganizationToCursor(
	value *organization.Organization,
	cursor ports.OrganizationListCursor,
	order ports.OrganizationOrder,
) int {
	result := compareOrganizationValues(
		value.ID(),
		value.Name(),
		value.CreateTime(),
		value.UpdateTime(),
		cursor.ID,
		cursor.Name,
		cursor.CreateTime,
		cursor.UpdateTime,
		order.Field,
	)
	if order.Direction == ports.SortDirectionDescending {
		return -result
	}
	return result
}

func compareOrganizationValues(
	leftID organization.ID,
	leftName string,
	leftCreateTime time.Time,
	leftUpdateTime time.Time,
	rightID organization.ID,
	rightName string,
	rightCreateTime time.Time,
	rightUpdateTime time.Time,
	field ports.OrganizationOrderField,
) int {
	var result int
	switch field {
	case ports.OrganizationOrderFieldName:
		result = strings.Compare(leftName, rightName)
	case ports.OrganizationOrderFieldCreateTime:
		result = leftCreateTime.Compare(rightCreateTime)
	case ports.OrganizationOrderFieldUpdateTime:
		result = leftUpdateTime.Compare(rightUpdateTime)
	case ports.OrganizationOrderFieldID:
		return strings.Compare(leftID.String(), rightID.String())
	}

	if result != 0 {
		return result
	}
	return strings.Compare(leftID.String(), rightID.String())
}

func newOrganizationCursor(value *organization.Organization) ports.OrganizationListCursor {
	return ports.OrganizationListCursor{
		ID:         value.ID(),
		Name:       value.Name(),
		CreateTime: value.CreateTime(),
		UpdateTime: value.UpdateTime(),
	}
}
