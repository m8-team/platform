package organizationcommand

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	organizationentity "github.com/m8platform/platform/internal/entity/resourcemanager/organization"
	organizationboundary "github.com/m8platform/platform/internal/usecase/resourcemanager/organization/boundary"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/port"
)

func TestDeleteInteractorAllowMissingUsesErrorsIs(t *testing.T) {
	t.Parallel()

	interactor := DeleteInteractor{
		Executor: CommandExecutor{},
		Reader: organizationReaderStub{
			getByID: func(context.Context, string, bool) (organizationentity.Entity, error) {
				return organizationentity.Entity{}, fmt.Errorf("wrapped: %w", organizationentity.ErrNotFound)
			},
		},
	}

	_, err := interactor.Execute(context.Background(), organizationboundary.DeleteOrganizationInput{
		ID:           "missing",
		AllowMissing: true,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

type organizationReaderStub struct {
	getByID func(context.Context, string, bool) (organizationentity.Entity, error)
	list    func(context.Context, port.OrganizationListParams) (port.OrganizationPage, error)
}

func (s organizationReaderStub) GetByID(ctx context.Context, id string, includeDeleted bool) (organizationentity.Entity, error) {
	if s.getByID != nil {
		return s.getByID(ctx, id, includeDeleted)
	}
	return organizationentity.Entity{}, organizationentity.ErrNotFound
}

func (s organizationReaderStub) List(ctx context.Context, params port.OrganizationListParams) (port.OrganizationPage, error) {
	if s.list != nil {
		return s.list(ctx, params)
	}
	return port.OrganizationPage{}, nil
}

type organizationWriterStub struct {
	updateErr error
}

func (organizationWriterStub) Create(context.Context, organizationentity.Entity) error { return nil }
func (w organizationWriterStub) Update(context.Context, organizationentity.Entity) error {
	return w.updateErr
}
func (w organizationWriterStub) SoftDelete(context.Context, organizationentity.Entity) error {
	return w.updateErr
}
func (w organizationWriterStub) Undelete(context.Context, organizationentity.Entity) error {
	return w.updateErr
}

type hierarchyReaderStub struct{}

func (hierarchyReaderStub) GetOrganizationNode(context.Context, string) (port.HierarchyNode, error) {
	return port.HierarchyNode{}, nil
}

func (hierarchyReaderStub) GetWorkspaceNode(context.Context, string) (port.HierarchyNode, error) {
	return port.HierarchyNode{}, nil
}

func (hierarchyReaderStub) HasActiveWorkspaces(context.Context, string) (bool, error) {
	return false, nil
}

func (hierarchyReaderStub) HasActiveProjects(context.Context, string) (bool, error) {
	return false, nil
}

type clockStub struct{}

func (clockStub) Now() time.Time { return time.Date(2026, 4, 8, 10, 0, 0, 0, time.UTC) }

type uuidStub struct{}

func (uuidStub) NewString() string { return "etag-2" }

var _ = errors.Is
