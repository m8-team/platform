package grpctransport

import (
	"context"
	"testing"

	"github.com/m8platform/platform/internal/application/query"
	"github.com/m8platform/platform/internal/domain/organization"
	"github.com/m8platform/platform/internal/testutil"
	resourcemanagerv1 "m8/platform/resourcemanager/v1"
)

func TestOrganizationServerGetSmoke(t *testing.T) {
	repo := testutil.NewOrganizationRepository()
	item, err := organization.New(organization.CreateParams{
		ID:          "11111111-1111-4111-8111-111111111111",
		Name:        "Acme",
		Description: "Root organization",
		Now:         testutil.Clock{}.Now(),
		ETag:        "etag-1",
	})
	if err != nil {
		t.Fatalf("organization.New() error = %v", err)
	}
	repo.Items[item.ID] = item

	server := &OrganizationServer{
		Get: query.GetOrganizationHandler{Repository: repo},
	}

	response, err := server.GetOrganization(context.Background(), &resourcemanagerv1.GetOrganizationRequest{
		Id: item.ID,
	})
	if err != nil {
		t.Fatalf("GetOrganization() error = %v", err)
	}
	if response.GetId() != item.ID {
		t.Fatalf("expected id %q, got %q", item.ID, response.GetId())
	}
	if response.GetName() != item.Name {
		t.Fatalf("expected name %q, got %q", item.Name, response.GetName())
	}
}
