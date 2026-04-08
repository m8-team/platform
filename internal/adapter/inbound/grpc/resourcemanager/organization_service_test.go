package grpcadapter

import (
	"context"
	"testing"
	"time"

	resourcemanagerv1 "m8/platform/resourcemanager/v1"

	grpcpresenter "github.com/m8platform/platform/internal/adapter/presenters/grpc/resourcemanager"
	organizationboundary "github.com/m8platform/platform/internal/usecase/resourcemanager/organization/boundary"
	"google.golang.org/grpc/metadata"
)

type captureOrganizationCommands struct {
	createInput organizationboundary.CreateOrganizationInput
}

func (c *captureOrganizationCommands) CreateOrganization(_ context.Context, input organizationboundary.CreateOrganizationInput) (organizationboundary.CreateOrganizationOutput, error) {
	c.createInput = input
	return organizationboundary.CreateOrganizationOutput{
		Organization: organizationboundary.Organization{
			ID:         "00000000-0000-0000-0000-000000000001",
			State:      "ACTIVE",
			Name:       input.Name,
			CreateTime: time.Date(2026, 4, 8, 10, 0, 0, 0, time.UTC),
			UpdateTime: time.Date(2026, 4, 8, 10, 0, 0, 0, time.UTC),
		},
	}, nil
}

func (*captureOrganizationCommands) UpdateOrganization(context.Context, organizationboundary.UpdateOrganizationInput) (organizationboundary.UpdateOrganizationOutput, error) {
	panic("unexpected call")
}

func (*captureOrganizationCommands) DeleteOrganization(context.Context, organizationboundary.DeleteOrganizationInput) (organizationboundary.DeleteOrganizationOutput, error) {
	panic("unexpected call")
}

func (*captureOrganizationCommands) UndeleteOrganization(context.Context, organizationboundary.UndeleteOrganizationInput) (organizationboundary.UndeleteOrganizationOutput, error) {
	panic("unexpected call")
}

type noopOrganizationQueries struct{}

func (noopOrganizationQueries) GetOrganization(context.Context, organizationboundary.GetOrganizationInput) (organizationboundary.GetOrganizationOutput, error) {
	panic("unexpected call")
}

func (noopOrganizationQueries) ListOrganizations(context.Context, organizationboundary.ListOrganizationsInput) (organizationboundary.ListOrganizationsOutput, error) {
	panic("unexpected call")
}

func TestOrganizationServiceCreateMapsMetadata(t *testing.T) {
	t.Parallel()

	commands := &captureOrganizationCommands{}
	server := OrganizationServiceServer{
		Commands:  commands,
		Queries:   noopOrganizationQueries{},
		Presenter: grpcpresenter.OrganizationPresenter{},
	}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"idempotency-key", "idem-1",
		"x-actor", "user-1",
		"x-correlation-id", "corr-1",
	))

	resp, err := server.CreateOrganization(ctx, &resourcemanagerv1.CreateOrganizationRequest{
		Organization: &resourcemanagerv1.Organization{
			Name:        "Acme",
			Description: "Root organization",
		},
	})
	if err != nil {
		t.Fatalf("CreateOrganization returned error: %v", err)
	}
	if commands.createInput.Metadata.IdempotencyKey != "idem-1" {
		t.Fatalf("expected idempotency metadata to be mapped, got %q", commands.createInput.Metadata.IdempotencyKey)
	}
	if commands.createInput.Metadata.Actor != "user-1" {
		t.Fatalf("expected actor metadata to be mapped, got %q", commands.createInput.Metadata.Actor)
	}
	if resp.GetName() != "Acme" {
		t.Fatalf("expected response name to be mapped, got %q", resp.GetName())
	}
}
