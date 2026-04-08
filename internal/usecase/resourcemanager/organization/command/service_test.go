package organizationcommand

import (
	"context"
	"errors"
	"testing"

	organizationboundary "github.com/m8platform/platform/internal/usecase/resourcemanager/organization/boundary"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/requestmeta"
)

func TestCommandServiceValidatesMetadata(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("metadata invalid")
	service := CommandService{
		CreateHandler: createStub{},
		MetadataValidator: metadataValidatorFunc(func(requestmeta.RequestMetadata) error {
			return expectedErr
		}),
	}

	_, err := service.CreateOrganization(context.Background(), organizationboundary.CreateOrganizationInput{})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected metadata validation error, got %v", err)
	}
}

type createStub struct{}

func (createStub) Execute(context.Context, organizationboundary.CreateOrganizationInput) (organizationboundary.CreateOrganizationOutput, error) {
	return organizationboundary.CreateOrganizationOutput{}, nil
}

type metadataValidatorFunc func(requestmeta.RequestMetadata) error

func (f metadataValidatorFunc) Validate(metadata requestmeta.RequestMetadata) error {
	return f(metadata)
}
