package redis

import (
	"testing"

	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
)

func TestBuildCheckAccessCacheKey(t *testing.T) {
	key := BuildCheckAccessCacheKey(
		&authzv1.SubjectRef{Type: authzv1.SubjectType_SUBJECT_TYPE_USER_ACCOUNT, Id: "user-1"},
		&authzv1.ResourceRef{Type: authzv1.ResourceType_RESOURCE_TYPE_PROJECT, Id: "project-1"},
		"project.read",
		"v1",
	)
	expected := "authz:SUBJECT_TYPE_USER_ACCOUNT:user-1:RESOURCE_TYPE_PROJECT:project-1:project.read:v1"
	if key != expected {
		t.Fatalf("expected %q, got %q", expected, key)
	}
}
