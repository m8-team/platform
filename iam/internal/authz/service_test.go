package authz

import (
	"context"
	"testing"

	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
	"github.com/m8platform/platform/iam/internal/config"
	"github.com/m8platform/platform/iam/internal/testutil"
	"go.uber.org/zap"
)

func TestSetBindingsFallbackCheckAndExplain(t *testing.T) {
	store := testutil.NewFakeStore()
	publisher := &testutil.FakePublisher{}
	cache := testutil.NewFakeCache()
	service := NewService(store, cache, publisher, nil, zap.NewNop(), config.Load())

	binding := &authzv1.AccessBinding{
		BindingId: "bind-1",
		Subject: &authzv1.SubjectRef{
			Type:     authzv1.SubjectType_SUBJECT_TYPE_USER_ACCOUNT,
			Id:       "user-1",
			TenantId: "tenant-1",
		},
		Resource: &authzv1.ResourceRef{
			Type:     authzv1.ResourceType_RESOURCE_TYPE_PROJECT,
			Id:       "project-1",
			TenantId: "tenant-1",
		},
		RoleId: "project-editor",
		Reason: "seed",
	}

	if _, err := service.SetAccessBindings(context.Background(), &authzv1.SetAccessBindingsRequest{
		RequestId:       "request-12345",
		Resource:        binding.GetResource(),
		DesiredBindings: []*authzv1.AccessBinding{binding},
		Reason:          "seed",
		PerformedBy:     "admin-1",
	}); err != nil {
		t.Fatalf("SetAccessBindings returned error: %v", err)
	}

	check, err := service.CheckAccess(context.Background(), &authzv1.CheckAccessRequest{
		Subject:    binding.GetSubject(),
		Resource:   binding.GetResource(),
		Permission: "project.write",
	})
	if err != nil {
		t.Fatalf("CheckAccess returned error: %v", err)
	}
	if check.GetDecision() != authzv1.PermissionDecision_PERMISSION_DECISION_ALLOW {
		t.Fatalf("expected ALLOW decision, got %s", check.GetDecision().String())
	}

	explain, err := service.ExplainAccess(context.Background(), &authzv1.ExplainAccessRequest{
		Subject:    binding.GetSubject(),
		Resource:   binding.GetResource(),
		Permission: "project.write",
	})
	if err != nil {
		t.Fatalf("ExplainAccess returned error: %v", err)
	}
	if explain.GetDecision() != authzv1.PermissionDecision_PERMISSION_DECISION_ALLOW {
		t.Fatalf("expected ALLOW explain decision, got %s", explain.GetDecision().String())
	}
	if len(explain.GetPathIds()) != 1 || explain.GetPathIds()[0] != "bind-1" {
		t.Fatalf("unexpected explain paths: %#v", explain.GetPathIds())
	}
	if len(publisher.Topics) == 0 {
		t.Fatal("expected domain event publication")
	}
}
