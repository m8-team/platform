package grpcadapter

import (
	"errors"
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	pb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// Schema validation does not exercise server persistence, authorization, kind
// registration or field-mask merging; those remain application responsibilities.
func TestResourceContract(t *testing.T) {
	validator, err := protovalidate.New()
	if err != nil {
		t.Fatal(err)
	}
	const id = "00000000-0000-4000-8000-000000000001"

	tests := []struct {
		name    string
		message proto.Message
		invalid bool
	}{
		{name: "bare resource kind", message: &pb.ResourceRef{Kind: "workload_service", Id: id}},
		{name: "namespaced resource kind", message: &pb.ResourceRef{Kind: "agents.agent", Id: id}},
		{name: "missing resource kind", message: &pb.ResourceRef{Id: id}, invalid: true},
		{name: "missing resource ID", message: &pb.ResourceRef{Kind: "project"}, invalid: true},
		{name: "malformed resource ID", message: &pb.ResourceRef{Kind: "project", Id: "invalid"}, invalid: true},
		{name: "missing owner selection", message: &pb.OwnerScope{}, invalid: true},
		{name: "organization owner", message: &pb.OwnerScope{Owner: &pb.OwnerScope_OrganizationId{OrganizationId: id}}},
		{name: "workspace owner", message: &pb.OwnerScope{Owner: &pb.OwnerScope_WorkspaceId{WorkspaceId: id}}},
		{name: "project owner", message: &pb.OwnerScope{Owner: &pb.OwnerScope_ProjectId{ProjectId: id}}},
		{name: "empty selected organization", message: &pb.OwnerScope{Owner: &pb.OwnerScope_OrganizationId{}}, invalid: true},
		{name: "empty selected workspace", message: &pb.OwnerScope{Owner: &pb.OwnerScope_WorkspaceId{}}, invalid: true},
		{name: "empty selected project", message: &pb.OwnerScope{Owner: &pb.OwnerScope_ProjectId{}}, invalid: true},
		{name: "malformed organization", message: &pb.OwnerScope{Owner: &pb.OwnerScope_OrganizationId{OrganizationId: "invalid"}}, invalid: true},
		{name: "malformed workspace", message: &pb.OwnerScope{Owner: &pb.OwnerScope_WorkspaceId{WorkspaceId: "invalid"}}, invalid: true},
		{name: "malformed project", message: &pb.OwnerScope{Owner: &pb.OwnerScope_ProjectId{ProjectId: "invalid"}}, invalid: true},
		{name: "omitted optional owner filter", message: &pb.ListResourcesRequest{Kind: "project", OrganizationId: id}},
		{name: "present empty owner filter", message: &pb.ListResourcesRequest{Kind: "project", OrganizationId: id, Owner: &pb.OwnerScope{}}, invalid: true},
		{
			name: "sparse metadata update with owner omitted",
			message: &pb.UpdateResourceRequest{
				Id: id,
				Resource: &pb.ResourceRecord{Resource: &pb.ResourceRecord_Project{Project: &pb.Project{
					Metadata: &pb.ResourceMetadata{Description: "Updated description"},
				}}},
				UpdateMask:       &fieldmaskpb.FieldMask{Paths: []string{"metadata.description"}},
				ExpectedRevision: 1,
				IdempotencyKey:   "update-project-description",
			},
		},
		{name: "metadata without server fields or display name", message: &pb.ResourceMetadata{}},
		{name: "metadata with server ID and revision", message: &pb.ResourceMetadata{Id: id, Revision: 1}},
		{name: "malformed metadata ID", message: &pb.ResourceMetadata{Id: "invalid"}, invalid: true},
		{name: "unknown metadata state", message: &pb.ResourceMetadata{State: pb.ResourceState(99)}, invalid: true},
		{name: "negative metadata revision", message: &pb.ResourceMetadata{Revision: -1}, invalid: true},
		{name: "Unicode name at limit", message: &pb.ResourceMetadata{Name: strings.Repeat("Я", 256)}},
		{name: "Unicode name over limit", message: &pb.ResourceMetadata{Name: strings.Repeat("Я", 257)}, invalid: true},
		{name: "Unicode description at limit", message: &pb.ResourceMetadata{Description: strings.Repeat("Я", 1024)}},
		{name: "Unicode description over limit", message: &pb.ResourceMetadata{Description: strings.Repeat("Я", 1025)}, invalid: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertResourceContractValidation(t, validator, tt.message, tt.invalid)
		})
	}
	for _, state := range []pb.ResourceState{
		pb.ResourceState_RESOURCE_STATE_UNSPECIFIED,
		pb.ResourceState_RESOURCE_STATE_ACTIVE,
		pb.ResourceState_RESOURCE_STATE_SUSPENDED,
		pb.ResourceState_RESOURCE_STATE_ARCHIVED,
		pb.ResourceState_RESOURCE_STATE_DELETED,
	} {
		t.Run(state.String(), func(t *testing.T) {
			assertResourceContractValidation(t, validator, &pb.ResourceMetadata{State: state}, false)
		})
	}
}

func TestResourceDescriptorContract(t *testing.T) {
	validator, err := protovalidate.New()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		change  func(*pb.ResourceDescriptor)
		invalid bool
	}{
		{name: "create without server assigned ID or owning organization"},
		{
			name: "catalog ID differs from referenced resource ID",
			change: func(descriptor *pb.ResourceDescriptor) {
				descriptor.Metadata.Id = "00000000-0000-4000-8000-000000000003"
				descriptor.OwningOrganizationId = descriptor.Owner.GetOrganizationId()
			},
		},
		{name: "missing resource", change: func(descriptor *pb.ResourceDescriptor) { descriptor.Resource = nil }, invalid: true},
		{name: "empty resource", change: func(descriptor *pb.ResourceDescriptor) { descriptor.Resource = &pb.ResourceRef{} }, invalid: true},
		{name: "missing owner", change: func(descriptor *pb.ResourceDescriptor) { descriptor.Owner = nil }, invalid: true},
		{name: "empty owner", change: func(descriptor *pb.ResourceDescriptor) { descriptor.Owner = &pb.OwnerScope{} }, invalid: true},
		{name: "missing metadata", change: func(descriptor *pb.ResourceDescriptor) { descriptor.Metadata = nil }, invalid: true},
		{name: "malformed metadata ID", change: func(descriptor *pb.ResourceDescriptor) { descriptor.Metadata.Id = "invalid" }, invalid: true},
		{name: "malformed derived organization ID", change: func(descriptor *pb.ResourceDescriptor) { descriptor.OwningOrganizationId = "invalid" }, invalid: true},
		{name: "missing module", change: func(descriptor *pb.ResourceDescriptor) { descriptor.Module = "" }, invalid: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			descriptor := &pb.ResourceDescriptor{
				Resource: &pb.ResourceRef{Kind: "agents.agent", Id: "00000000-0000-4000-8000-000000000001"},
				Owner:    &pb.OwnerScope{Owner: &pb.OwnerScope_OrganizationId{OrganizationId: "00000000-0000-4000-8000-000000000002"}},
				Metadata: &pb.ResourceMetadata{},
				Module:   "agents",
			}
			if tt.change != nil {
				tt.change(descriptor)
			}
			assertResourceContractValidation(t, validator, descriptor, tt.invalid)
		})
	}
}

func TestResourceKindPolicyContract(t *testing.T) {
	validator, err := protovalidate.New()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		change  func(*pb.ResourceKindPolicy)
		invalid bool
	}{
		{name: "empty owners and false binding permission deny placement and binding"},
		{
			name: "all supported owners with binding enabled",
			change: func(policy *pb.ResourceKindPolicy) {
				policy.AllowedOwnerKinds = []string{"organization", "workspace", "project"}
				policy.EnvironmentBindingAllowed = true
			},
		},
		{name: "missing kind", change: func(policy *pb.ResourceKindPolicy) { policy.Kind = "" }, invalid: true},
		{name: "missing responsible module", change: func(policy *pb.ResourceKindPolicy) { policy.ResponsibleModule = "" }, invalid: true},
		{name: "duplicate owners", change: func(policy *pb.ResourceKindPolicy) { policy.AllowedOwnerKinds = []string{"project", "project"} }, invalid: true},
		{name: "empty owner kind", change: func(policy *pb.ResourceKindPolicy) { policy.AllowedOwnerKinds = []string{""} }, invalid: true},
		{name: "tenant is not a descriptor owner", change: func(policy *pb.ResourceKindPolicy) { policy.AllowedOwnerKinds = []string{"tenant"} }, invalid: true},
		{name: "group is not a containment owner", change: func(policy *pb.ResourceKindPolicy) { policy.AllowedOwnerKinds = []string{"group"} }, invalid: true},
		{name: "environment is not a containment owner", change: func(policy *pb.ResourceKindPolicy) { policy.AllowedOwnerKinds = []string{"environment"} }, invalid: true},
		{name: "unknown owner kind", change: func(policy *pb.ResourceKindPolicy) { policy.AllowedOwnerKinds = []string{"agents.agent"} }, invalid: true},
		{name: "owner kinds are case sensitive", change: func(policy *pb.ResourceKindPolicy) { policy.AllowedOwnerKinds = []string{"Organization"} }, invalid: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := &pb.ResourceKindPolicy{Kind: "agents.agent", ResponsibleModule: "agents"}
			if tt.change != nil {
				tt.change(policy)
			}
			assertResourceContractValidation(t, validator, policy, tt.invalid)
		})
	}
}

func assertResourceContractValidation(t *testing.T, validator protovalidate.Validator, message proto.Message, invalid bool) {
	t.Helper()
	err := validator.Validate(message)
	if !invalid {
		if err != nil {
			t.Fatalf("valid contract message rejected: %v", err)
		}
		return
	}
	var validationError *protovalidate.ValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("want ValidationError, got %T: %v", err, err)
	}
}
