package support

import (
	"context"
	"testing"
	"time"

	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
	supportv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/support/v1"
	"github.com/m8platform/platform/iam/internal/config"
	"github.com/m8platform/platform/iam/internal/testutil"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestSupportGrantLifecycle(t *testing.T) {
	store := testutil.NewFakeStore()
	publisher := &testutil.FakePublisher{}
	workflows := &testutil.FakeWorkflowStarter{}
	service := NewService(store, publisher, workflows, zap.NewNop(), config.Load())

	grant, err := service.GrantTemporaryAccess(context.Background(), &supportv1.GrantTemporaryAccessRequest{
		RequestId: "grant-req-1234",
		TenantId:  "tenant-1",
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
		RoleId:      "support-operator",
		Ttl:         durationpb.New(5 * time.Minute),
		Reason:      "support session",
		RequestedBy: "support-admin",
	})
	if err != nil {
		t.Fatalf("GrantTemporaryAccess returned error: %v", err)
	}
	if grant.GetStatus() != supportv1.SupportGrantStatus_SUPPORT_GRANT_STATUS_PENDING_APPROVAL {
		t.Fatalf("expected pending grant, got %s", grant.GetStatus().String())
	}

	approved, err := service.ApproveTemporaryAccess(context.Background(), &supportv1.ApproveTemporaryAccessRequest{
		SupportGrantId: grant.GetSupportGrantId(),
		ApprovalTicket: "TICKET-1",
		Reason:         "approved",
		ApprovedBy:     "lead-1",
	})
	if err != nil {
		t.Fatalf("ApproveTemporaryAccess returned error: %v", err)
	}
	if approved.GetStatus() != supportv1.SupportGrantStatus_SUPPORT_GRANT_STATUS_ACTIVE {
		t.Fatalf("expected active grant, got %s", approved.GetStatus().String())
	}
	if approved.GetExpiresAt().AsTime().Sub(approved.GetApprovedAt().AsTime()) < 5*time.Minute {
		t.Fatal("expected expires_at to be at least ttl after approval")
	}
	if len(workflows.Calls) != 1 {
		t.Fatalf("expected 1 workflow start, got %d", len(workflows.Calls))
	}
}
