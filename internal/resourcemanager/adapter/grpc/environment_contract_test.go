package grpcadapter

import (
	"errors"
	"testing"
	"time"

	"buf.build/go/protovalidate"
	pb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// These tests exercise the generated contract independently of the development
// server, which does not yet enforce or persist the complete Environment model.
func TestEnvironmentContract(t *testing.T) {
	validator, err := protovalidate.New()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		change  func(*pb.Environment)
		invalid bool
	}{
		{name: "create without server assigned ID"},
		{
			name: "resource with server assigned ID",
			change: func(env *pb.Environment) {
				env.Metadata.Id = "00000000-0000-4000-8000-000000000002"
			},
		},
		{
			name: "custom key independent of kind",
			change: func(env *pb.Environment) {
				env.Key = "release-candidate-42"
			},
		},
		{
			name: "production with expired timestamp remains readable",
			change: func(env *pb.Environment) {
				env.ExpiresAt = timestamppb.New(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))
			},
		},
		{
			name: "preview without expiry",
			change: func(env *pb.Environment) {
				env.Kind = pb.EnvironmentKind_ENVIRONMENT_KIND_PREVIEW
			},
		},
		{name: "missing metadata", change: func(env *pb.Environment) { env.Metadata = nil }, invalid: true},
		{name: "malformed metadata ID", change: func(env *pb.Environment) { env.Metadata.Id = "invalid" }, invalid: true},
		{name: "missing organization ID", change: func(env *pb.Environment) { env.OrganizationId = "" }, invalid: true},
		{name: "malformed organization ID", change: func(env *pb.Environment) { env.OrganizationId = "invalid" }, invalid: true},
		{name: "empty key", change: func(env *pb.Environment) { env.Key = "" }, invalid: true},
		{name: "uppercase key", change: func(env *pb.Environment) { env.Key = "Production" }, invalid: true},
		{name: "underscore in key", change: func(env *pb.Environment) { env.Key = "release_candidate" }, invalid: true},
		{name: "leading hyphen in key", change: func(env *pb.Environment) { env.Key = "-production" }, invalid: true},
		{name: "trailing hyphen in key", change: func(env *pb.Environment) { env.Key = "production-" }, invalid: true},
		{name: "consecutive hyphens in key", change: func(env *pb.Environment) { env.Key = "release--candidate" }, invalid: true},
		{name: "unspecified kind", change: func(env *pb.Environment) { env.Kind = pb.EnvironmentKind_ENVIRONMENT_KIND_UNSPECIFIED }, invalid: true},
		{name: "unknown kind", change: func(env *pb.Environment) { env.Kind = pb.EnvironmentKind(99) }, invalid: true},
		{name: "unspecified isolation", change: func(env *pb.Environment) { env.Isolation = pb.IsolationProfile_ISOLATION_PROFILE_UNSPECIFIED }, invalid: true},
		{name: "unknown isolation", change: func(env *pb.Environment) { env.Isolation = pb.IsolationProfile(99) }, invalid: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := &pb.Environment{
				Metadata:       &pb.ResourceMetadata{Name: "Production"},
				OrganizationId: "00000000-0000-4000-8000-000000000001",
				Key:            "production",
				Kind:           pb.EnvironmentKind_ENVIRONMENT_KIND_PRODUCTION,
				Isolation:      pb.IsolationProfile_ISOLATION_PROFILE_LOGICAL,
			}
			if tt.change != nil {
				tt.change(env)
			}
			assertEnvironmentContractValidation(t, validator, env, tt.invalid)
		})
	}
}

func TestEnvironmentBindingContract(t *testing.T) {
	validator, err := protovalidate.New()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		change  func(*pb.EnvironmentBinding)
		invalid bool
	}{
		{name: "create without server assigned ID or configuration"},
		{
			name: "resource with server assigned ID and revision",
			change: func(binding *pb.EnvironmentBinding) {
				binding.Id = "00000000-0000-4000-8000-000000000003"
				binding.Revision = 1
				binding.ConfigurationRef = "config/releases/42"
			},
		},
		{name: "malformed binding ID", change: func(binding *pb.EnvironmentBinding) { binding.Id = "invalid" }, invalid: true},
		{name: "missing resource", change: func(binding *pb.EnvironmentBinding) { binding.Resource = nil }, invalid: true},
		{name: "missing resource kind", change: func(binding *pb.EnvironmentBinding) { binding.Resource.Kind = "" }, invalid: true},
		{name: "missing resource ID", change: func(binding *pb.EnvironmentBinding) { binding.Resource.Id = "" }, invalid: true},
		{name: "malformed resource ID", change: func(binding *pb.EnvironmentBinding) { binding.Resource.Id = "invalid" }, invalid: true},
		{name: "missing environment ID", change: func(binding *pb.EnvironmentBinding) { binding.EnvironmentId = "" }, invalid: true},
		{name: "malformed environment ID", change: func(binding *pb.EnvironmentBinding) { binding.EnvironmentId = "invalid" }, invalid: true},
		{name: "empty version reference", change: func(binding *pb.EnvironmentBinding) { binding.VersionRef = "" }, invalid: true},
		{name: "negative revision", change: func(binding *pb.EnvironmentBinding) { binding.Revision = -1 }, invalid: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binding := &pb.EnvironmentBinding{
				Resource: &pb.ResourceRef{
					Kind: "workload_service",
					Id:   "00000000-0000-4000-8000-000000000001",
				},
				EnvironmentId: "00000000-0000-4000-8000-000000000002",
				VersionRef:    "releases/42",
			}
			if tt.change != nil {
				tt.change(binding)
			}
			assertEnvironmentContractValidation(t, validator, binding, tt.invalid)
		})
	}
}

func assertEnvironmentContractValidation(t *testing.T, validator protovalidate.Validator, message proto.Message, invalid bool) {
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
