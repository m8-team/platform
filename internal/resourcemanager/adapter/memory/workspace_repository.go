package memory

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/m8-team/platform/internal/platform/types"
	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
)

var _ ports.WorkspaceRepository = (*WorkspaceRepository)(nil)

// WorkspaceRepository is a process-local repository intended for tests and
// local development. It preserves repository semantics, including detached
// reads and atomic compare-and-swap updates, but it is not durable storage.
type WorkspaceRepository struct {
	hierarchyMu sync.Mutex
	hierarchy   map[organization.ID]*organizationLock
	mu          sync.RWMutex
	workspaces  map[workspace.ID]*workspace.Workspace
}

func NewWorkspaceRepository() *WorkspaceRepository {
	return &WorkspaceRepository{
		workspaces: make(map[workspace.ID]*workspace.Workspace),
	}
}

func (r *WorkspaceRepository) Create(ctx context.Context, value *workspace.Workspace) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	if value == nil {
		return ports.ErrNilWorkspace
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if err := contextError(ctx); err != nil {
		return err
	}
	if r.workspaces == nil {
		r.workspaces = make(map[workspace.ID]*workspace.Workspace)
	}
	if _, exists := r.workspaces[value.ID()]; exists {
		return fmt.Errorf("%w: workspace_id=%s", ports.ErrWorkspaceAlreadyExists, value.ID())
	}

	r.workspaces[value.ID()] = value.Clone()
	return nil
}

func (r *WorkspaceRepository) Get(
	ctx context.Context,
	id workspace.ID,
) (*workspace.Workspace, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := contextError(ctx); err != nil {
		return nil, err
	}
	value, exists := r.workspaces[id]
	if !exists {
		return nil, fmt.Errorf("%w: workspace_id=%s", ports.ErrWorkspaceNotFound, id)
	}

	return value.Clone(), nil
}

func (r *WorkspaceRepository) Update(
	ctx context.Context,
	value *workspace.Workspace,
	expectedVersion types.Version,
) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	if value == nil {
		return ports.ErrNilWorkspace
	}
	if err := expectedVersion.Validate(); err != nil {
		return fmt.Errorf("expected workspace version: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if err := contextError(ctx); err != nil {
		return err
	}
	current, exists := r.workspaces[value.ID()]
	if !exists {
		return fmt.Errorf("%w: workspace_id=%s", ports.ErrWorkspaceNotFound, value.ID())
	}
	if !current.Version().Equal(expectedVersion) {
		return &ports.WorkspaceVersionConflictError{
			ID:       value.ID(),
			Expected: expectedVersion,
			Actual:   current.Version(),
		}
	}
	nextVersion, err := expectedVersion.Next()
	if err != nil {
		return fmt.Errorf("next workspace version: %w", err)
	}
	if !value.Version().Equal(nextVersion) {
		return fmt.Errorf(
			"%w: workspace_id=%s expected_next=%s incoming=%s",
			ports.ErrInvalidWorkspaceVersion,
			value.ID(),
			nextVersion,
			value.Version(),
		)
	}

	r.workspaces[value.ID()] = value.Clone()
	return nil
}

func (r *WorkspaceRepository) List(
	ctx context.Context,
	options ports.ListWorkspacesOptions,
) (ports.ListWorkspacesResult, error) {
	if err := contextError(ctx); err != nil {
		return ports.ListWorkspacesResult{}, err
	}

	options = options.WithDefaults()
	if err := options.Validate(); err != nil {
		return ports.ListWorkspacesResult{}, err
	}

	r.mu.RLock()
	values := make([]*workspace.Workspace, 0, len(r.workspaces))
	for _, value := range r.workspaces {
		if (options.OrganizationID.IsZero() || value.OrganizationID().Equal(options.OrganizationID)) && matchesWorkspace(value, options.Filter) {
			values = append(values, value)
		}
	}
	r.mu.RUnlock()

	if err := contextError(ctx); err != nil {
		return ports.ListWorkspacesResult{}, err
	}

	slices.SortFunc(values, func(a, b *workspace.Workspace) int { return compareWorkspaces(a, b, options.Order) })

	totalSize := len(values)
	start := 0
	if options.After != nil {
		start = sort.Search(len(values), func(i int) bool {
			return compareWorkspaceToCursor(values[i], *options.After, options.Order) > 0
		})
	}

	end := min(start+options.PageSize, len(values))
	// Stored aggregates are immutable after insertion; only clone the returned page.
	page := make([]*workspace.Workspace, end-start)
	for i, value := range values[start:end] {
		page[i] = value.Clone()
	}

	var next *ports.WorkspaceListCursor
	if end < len(values) && len(page) > 0 {
		cursor := newWorkspaceCursor(page[len(page)-1])
		next = &cursor
	}

	return ports.ListWorkspacesResult{
		Workspaces: page,
		Next:       next,
		TotalSize:  totalSize,
	}, nil
}

func matchesWorkspace(value *workspace.Workspace, filter ports.WorkspaceFilter) bool {
	if !filter.ShowDeleted && value.State() == workspace.StateDeleted {
		return false
	}
	if len(filter.States) > 0 && !containsWorkspaceState(filter.States, value.State()) {
		return false
	}
	if filter.NameEquals != nil && value.Name() != *filter.NameEquals {
		return false
	}

	for key, expected := range filter.LabelsEqual {
		if actual, exists := value.Label(key); !exists || actual != expected {
			return false
		}
	}

	return true
}

func containsWorkspaceState(states []workspace.State, target workspace.State) bool {
	for _, state := range states {
		if state == target {
			return true
		}
	}
	return false
}

func compareWorkspaces(
	left *workspace.Workspace,
	right *workspace.Workspace,
	order ports.WorkspaceOrder,
) int {
	result := compareWorkspaceValues(
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

func compareWorkspaceToCursor(
	value *workspace.Workspace,
	cursor ports.WorkspaceListCursor,
	order ports.WorkspaceOrder,
) int {
	result := compareWorkspaceValues(
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

func compareWorkspaceValues(
	leftID workspace.ID,
	leftName string,
	leftCreateTime time.Time,
	leftUpdateTime time.Time,
	rightID workspace.ID,
	rightName string,
	rightCreateTime time.Time,
	rightUpdateTime time.Time,
	field ports.WorkspaceOrderField,
) int {
	var result int
	switch field {
	case ports.WorkspaceOrderFieldName:
		result = strings.Compare(leftName, rightName)
	case ports.WorkspaceOrderFieldCreateTime:
		result = leftCreateTime.Compare(rightCreateTime)
	case ports.WorkspaceOrderFieldUpdateTime:
		result = leftUpdateTime.Compare(rightUpdateTime)
	case ports.WorkspaceOrderFieldID:
		return leftID.Compare(rightID)
	}

	if result != 0 {
		return result
	}
	return leftID.Compare(rightID)
}

func newWorkspaceCursor(value *workspace.Workspace) ports.WorkspaceListCursor {
	return ports.WorkspaceListCursor{
		ID:         value.ID(),
		Name:       value.Name(),
		CreateTime: value.CreateTime(),
		UpdateTime: value.UpdateTime(),
	}
}

func (r *WorkspaceRepository) HasNonDeleted(
	ctx context.Context,
	organizationID organization.ID,
) (bool, error) {
	if err := contextError(ctx); err != nil {
		return false, err
	}
	if err := organizationID.Validate(); err != nil {
		return false, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, value := range r.workspaces {
		if value.OrganizationID().Equal(organizationID) && !value.IsDeleted() {
			return true, nil
		}
	}
	return false, nil
}

func (r *WorkspaceRepository) WithOrganizationLock(
	ctx context.Context,
	organizationID organization.ID,
	fn func(context.Context) error,
) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	if err := organizationID.Validate(); err != nil {
		return err
	}
	if fn == nil {
		return fmt.Errorf("organization hierarchy mutation is required")
	}

	r.hierarchyMu.Lock()
	if r.hierarchy == nil {
		r.hierarchy = make(map[organization.ID]*organizationLock)
	}
	lock := r.hierarchy[organizationID]
	if lock == nil {
		lock = &organizationLock{permit: make(chan struct{}, 1)}
		lock.permit <- struct{}{}
		r.hierarchy[organizationID] = lock
	}
	lock.references++
	r.hierarchyMu.Unlock()
	defer func() {
		r.hierarchyMu.Lock()
		defer r.hierarchyMu.Unlock()
		lock.references--
		if lock.references == 0 {
			delete(r.hierarchy, organizationID)
		}
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-lock.permit:
	}
	defer func() { lock.permit <- struct{}{} }()
	if err := ctx.Err(); err != nil {
		return err
	}
	return fn(ctx)
}

// Reference counting removes idle locks while preserving waiter identity.
type organizationLock struct {
	permit     chan struct{}
	references int
}
