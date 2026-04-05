package grpc

import (
	"context"
	"errors"
	"testing"

	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
	foundationconfig "github.com/m8platform/platform/iam/internal/foundation/config"
	"github.com/m8platform/platform/iam/internal/spicedb"
	"github.com/m8platform/platform/iam/internal/testutil"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestSetBindingsFallbackCheckAndExplain(t *testing.T) {
	store := testutil.NewFakeStore()
	publisher := &testutil.FakePublisher{}
	service := NewServer(store, testutil.NewFakeCache(), publisher, nil, zap.NewNop(), foundationconfig.Load().Redis.PolicyVersion, foundationconfig.Load().Topics, nil, nil, nil)

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

func TestCheckAccessLogsExpectedSpiceDBFallbackOnceAsInfo(t *testing.T) {
	store := testutil.NewFakeStore()
	publisher := &testutil.FakePublisher{}
	runtime := &testutil.FakeAuthorizationRuntime{Err: spicedb.ErrNotImplemented}
	core, observed := observer.New(zapcore.InfoLevel)
	service := NewServer(store, nil, publisher, runtime, zap.New(core), foundationconfig.Load().Redis.PolicyVersion, foundationconfig.Load().Topics, nil, nil, nil)

	binding := &authzv1.AccessBinding{
		BindingId: "bind-expected-fallback",
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

	for i := 0; i < 2; i++ {
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
	}

	entries := observed.FilterMessage("spicedb runtime unavailable; using fallback evaluator").All()
	if len(entries) != 1 {
		t.Fatalf("expected one info fallback log, got %d", len(entries))
	}
	if entries[0].Level != zapcore.InfoLevel {
		t.Fatalf("expected info level, got %s", entries[0].Level.String())
	}
	if warnings := observed.FilterLevelExact(zapcore.WarnLevel).All(); len(warnings) != 0 {
		t.Fatalf("expected no warn logs, got %d", len(warnings))
	}
}

func TestCheckAccessSkipsCacheWhenRuntimeIsConfigured(t *testing.T) {
	store := testutil.NewFakeStore()
	publisher := &testutil.FakePublisher{}
	cache := testutil.NewFakeCache()
	runtime := &testutil.FakeAuthorizationRuntime{
		Result: &authzv1.AccessCheckResult{
			Decision:   authzv1.PermissionDecision_PERMISSION_DECISION_ALLOW,
			Permission: "project.write",
			ZedToken:   "zed-1",
		},
	}
	service := NewServer(store, cache, publisher, runtime, zap.NewNop(), foundationconfig.Load().Redis.PolicyVersion, foundationconfig.Load().Topics, nil, nil, nil)

	req := &authzv1.CheckAccessRequest{
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
		Permission: "project.write",
	}

	first, err := service.CheckAccess(context.Background(), req)
	if err != nil {
		t.Fatalf("CheckAccess returned error: %v", err)
	}
	if first.GetDecision() != authzv1.PermissionDecision_PERMISSION_DECISION_ALLOW {
		t.Fatalf("expected first decision ALLOW, got %s", first.GetDecision().String())
	}
	if first.GetCacheHit() {
		t.Fatal("expected runtime-backed result to skip cache")
	}

	runtime.Result = &authzv1.AccessCheckResult{
		Decision:   authzv1.PermissionDecision_PERMISSION_DECISION_DENY,
		Permission: "project.write",
		ZedToken:   "zed-2",
	}
	second, err := service.CheckAccess(context.Background(), req)
	if err != nil {
		t.Fatalf("CheckAccess returned error: %v", err)
	}
	if second.GetDecision() != authzv1.PermissionDecision_PERMISSION_DECISION_DENY {
		t.Fatalf("expected second decision DENY, got %s", second.GetDecision().String())
	}
	if second.GetCacheHit() {
		t.Fatal("expected second runtime-backed result to skip cache")
	}
}

func TestSetAccessBindingsReplacesExistingSnapshotInStore(t *testing.T) {
	store := testutil.NewFakeStore()
	publisher := &testutil.FakePublisher{}
	service := NewServer(store, nil, publisher, nil, zap.NewNop(), foundationconfig.Load().Redis.PolicyVersion, foundationconfig.Load().Topics, nil, nil, nil)

	resource := &authzv1.ResourceRef{
		Type:     authzv1.ResourceType_RESOURCE_TYPE_PROJECT,
		Id:       "project-1",
		TenantId: "tenant-1",
	}
	initial := &authzv1.AccessBinding{
		BindingId: "bind-old",
		Subject: &authzv1.SubjectRef{
			Type:     authzv1.SubjectType_SUBJECT_TYPE_USER_ACCOUNT,
			Id:       "user-old",
			TenantId: "tenant-1",
		},
		Resource: resource,
		RoleId:   "project-viewer",
		Reason:   "old",
	}
	replacement := &authzv1.AccessBinding{
		BindingId: "bind-new",
		Subject: &authzv1.SubjectRef{
			Type:     authzv1.SubjectType_SUBJECT_TYPE_USER_ACCOUNT,
			Id:       "user-new",
			TenantId: "tenant-1",
		},
		Resource: resource,
		RoleId:   "project-editor",
		Reason:   "new",
	}

	if _, err := service.SetAccessBindings(context.Background(), &authzv1.SetAccessBindingsRequest{
		RequestId:       "request-old",
		Resource:        resource,
		DesiredBindings: []*authzv1.AccessBinding{initial},
		Reason:          "seed",
		PerformedBy:     "admin-1",
	}); err != nil {
		t.Fatalf("initial SetAccessBindings returned error: %v", err)
	}
	if _, err := service.SetAccessBindings(context.Background(), &authzv1.SetAccessBindingsRequest{
		RequestId:       "request-new",
		Resource:        resource,
		DesiredBindings: []*authzv1.AccessBinding{replacement},
		Reason:          "replace",
		PerformedBy:     "admin-1",
	}); err != nil {
		t.Fatalf("replacement SetAccessBindings returned error: %v", err)
	}

	bindings, err := ListBindingsForResource(context.Background(), store, resource)
	if err != nil {
		t.Fatalf("ListBindingsForResource returned error: %v", err)
	}
	if len(bindings) != 1 {
		t.Fatalf("expected one binding after replace, got %d", len(bindings))
	}
	if bindings[0].GetBindingId() != replacement.GetBindingId() {
		t.Fatalf("expected replacement binding %s, got %s", replacement.GetBindingId(), bindings[0].GetBindingId())
	}
}

func TestSetAccessBindingsRollsBackOnSpiceDBSyncFailure(t *testing.T) {
	store := testutil.NewFakeStore()
	publisher := &testutil.FakePublisher{}
	runtime := &testutil.FakeAuthorizationRuntime{SyncErr: errors.New("spicedb unavailable")}
	service := NewServer(store, nil, publisher, runtime, zap.NewNop(), foundationconfig.Load().Redis.PolicyVersion, foundationconfig.Load().Topics, nil, nil, nil)

	resource := &authzv1.ResourceRef{
		Type:     authzv1.ResourceType_RESOURCE_TYPE_PROJECT,
		Id:       "project-1",
		TenantId: "tenant-1",
	}
	previous := &authzv1.AccessBinding{
		BindingId: "bind-old",
		Subject: &authzv1.SubjectRef{
			Type:     authzv1.SubjectType_SUBJECT_TYPE_USER_ACCOUNT,
			Id:       "user-old",
			TenantId: "tenant-1",
		},
		Resource: resource,
		RoleId:   "project-viewer",
		Reason:   "old",
	}
	if _, err := service.SetAccessBindings(context.Background(), &authzv1.SetAccessBindingsRequest{
		RequestId:       "request-old",
		Resource:        resource,
		DesiredBindings: []*authzv1.AccessBinding{previous},
		Reason:          "seed",
		PerformedBy:     "admin-1",
	}); err == nil {
		t.Fatal("expected initial SetAccessBindings to fail because runtime sync is failing")
	}

	runtime.SyncErr = nil
	if _, err := service.SetAccessBindings(context.Background(), &authzv1.SetAccessBindingsRequest{
		RequestId:       "request-old-2",
		Resource:        resource,
		DesiredBindings: []*authzv1.AccessBinding{previous},
		Reason:          "seed",
		PerformedBy:     "admin-1",
	}); err != nil {
		t.Fatalf("stable SetAccessBindings returned error: %v", err)
	}

	runtime.SyncErr = errors.New("spicedb unavailable")
	replacement := &authzv1.AccessBinding{
		BindingId: "bind-new",
		Subject: &authzv1.SubjectRef{
			Type:     authzv1.SubjectType_SUBJECT_TYPE_USER_ACCOUNT,
			Id:       "user-new",
			TenantId: "tenant-1",
		},
		Resource: resource,
		RoleId:   "project-editor",
		Reason:   "new",
	}
	if _, err := service.SetAccessBindings(context.Background(), &authzv1.SetAccessBindingsRequest{
		RequestId:       "request-new",
		Resource:        resource,
		DesiredBindings: []*authzv1.AccessBinding{replacement},
		Reason:          "replace",
		PerformedBy:     "admin-1",
	}); err == nil {
		t.Fatal("expected SetAccessBindings to fail on unexpected SpiceDB sync error")
	}

	bindings, err := ListBindingsForResource(context.Background(), store, resource)
	if err != nil {
		t.Fatalf("ListBindingsForResource returned error: %v", err)
	}
	if len(bindings) != 1 {
		t.Fatalf("expected rollback to preserve one binding, got %d", len(bindings))
	}
	if bindings[0].GetBindingId() != previous.GetBindingId() {
		t.Fatalf("expected rollback to preserve %s, got %s", previous.GetBindingId(), bindings[0].GetBindingId())
	}
}
