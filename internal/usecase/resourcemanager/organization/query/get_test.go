package organizationquery

import (
	"context"
	"testing"
	"time"

	organizationentity "github.com/m8platform/platform/internal/entity/resourcemanager/organization"
	organizationboundary "github.com/m8platform/platform/internal/usecase/resourcemanager/organization/boundary"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/port"
)

func TestGetInteractorUsesIncludeDeleted(t *testing.T) {
	t.Parallel()

	var includeDeleted bool
	interactor := GetInteractor{
		Reader: organizationGetReaderStub{
			getByID: func(_ context.Context, _ string, value bool) (organizationentity.Entity, error) {
				includeDeleted = value
				return organizationentity.Entity{
					ID:         "org-1",
					State:      organizationentity.StateDeleted,
					Name:       "Acme",
					CreateTime: time.Date(2026, 4, 8, 10, 0, 0, 0, time.UTC),
					UpdateTime: time.Date(2026, 4, 8, 10, 0, 0, 0, time.UTC),
				}, nil
			},
		},
	}

	_, err := interactor.Execute(context.Background(), organizationboundary.GetOrganizationInput{
		ID:             "org-1",
		IncludeDeleted: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !includeDeleted {
		t.Fatal("expected includeDeleted=true to be passed to repository")
	}
}

type organizationGetReaderStub struct {
	getByID func(context.Context, string, bool) (organizationentity.Entity, error)
}

func (s organizationGetReaderStub) GetByID(ctx context.Context, id string, includeDeleted bool) (organizationentity.Entity, error) {
	return s.getByID(ctx, id, includeDeleted)
}

func (organizationGetReaderStub) List(context.Context, port.OrganizationListParams) (port.OrganizationPage, error) {
	return port.OrganizationPage{}, nil
}
