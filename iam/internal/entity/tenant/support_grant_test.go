package tenant

import (
	"errors"
	"testing"
	"time"

	authzentity "github.com/m8platform/platform/iam/internal/entity/authz"
)

func TestNewSupportGrant(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 4, 14, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		params  NewSupportGrantParams
		wantErr error
	}{
		{
			name: "valid",
			params: NewSupportGrantParams{
				TenantID: "tenant-1",
				Subject:  authzentity.SubjectRef{Type: "SUBJECT_TYPE_USER_ACCOUNT", ID: "user-1"},
				Resource: authzentity.ResourceRef{Type: "RESOURCE_TYPE_PROJECT", ID: "project-1"},
				RoleID:   "support-operator",
				TTL:      5 * time.Minute,
				Reason:   "incident triage",
				Now:      now,
			},
		},
		{
			name: "missing tenant",
			params: NewSupportGrantParams{
				Subject: authzentity.SubjectRef{Type: "SUBJECT_TYPE_USER_ACCOUNT", ID: "user-1"},
				Resource: authzentity.ResourceRef{
					Type: "RESOURCE_TYPE_PROJECT",
					ID:   "project-1",
				},
				RoleID: "support-operator",
				TTL:    5 * time.Minute,
				Reason: "incident triage",
				Now:    now,
			},
			wantErr: ErrSupportGrantTenantRequired,
		},
		{
			name: "non-positive ttl",
			params: NewSupportGrantParams{
				TenantID: "tenant-1",
				Subject:  authzentity.SubjectRef{Type: "SUBJECT_TYPE_USER_ACCOUNT", ID: "user-1"},
				Resource: authzentity.ResourceRef{Type: "RESOURCE_TYPE_PROJECT", ID: "project-1"},
				RoleID:   "support-operator",
				Reason:   "incident triage",
				Now:      now,
			},
			wantErr: ErrSupportGrantTTLRequired,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			grant, err := NewSupportGrant(tt.params)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if grant.ID == "" {
				t.Fatal("expected generated support grant id")
			}
			if grant.Status != SupportGrantStatusPendingApproval {
				t.Fatalf("expected pending status, got %s", grant.Status)
			}
		})
	}
}

func TestSupportGrantApproveAndRevoke(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 4, 14, 0, 0, 0, time.UTC)
	grant, err := NewSupportGrant(NewSupportGrantParams{
		TenantID: "tenant-1",
		Subject:  authzentity.SubjectRef{Type: "SUBJECT_TYPE_USER_ACCOUNT", ID: "user-1"},
		Resource: authzentity.ResourceRef{Type: "RESOURCE_TYPE_PROJECT", ID: "project-1"},
		RoleID:   "support-operator",
		TTL:      10 * time.Minute,
		Reason:   "incident triage",
		Now:      now,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	approved := grant.Approve("TICKET-1", now.Add(time.Minute))
	if approved.Status != SupportGrantStatusActive {
		t.Fatalf("expected active status, got %s", approved.Status)
	}
	if approved.ExpiresAt.Sub(approved.ApprovedAt) != 10*time.Minute {
		t.Fatalf("expected expiry to match ttl, got %s", approved.ExpiresAt.Sub(approved.ApprovedAt))
	}

	revoked := approved.Revoke(now.Add(2 * time.Minute))
	if revoked.Status != SupportGrantStatusRevoked {
		t.Fatalf("expected revoked status, got %s", revoked.Status)
	}
}
