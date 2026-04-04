package identity

import (
	"errors"
	"testing"
	"time"
)

func TestNewServiceAccount(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 4, 10, 11, 12, 0, time.UTC)
	tests := []struct {
		name    string
		params  NewServiceAccountParams
		wantErr error
	}{
		{
			name: "valid",
			params: NewServiceAccountParams{
				ID:          "sa-1",
				TenantID:    "tenant-1",
				DisplayName: "Platform Bot",
				Description: "build automation",
				Now:         now,
			},
		},
		{
			name: "missing id",
			params: NewServiceAccountParams{
				TenantID:    "tenant-1",
				DisplayName: "Platform Bot",
				Now:         now,
			},
			wantErr: ErrServiceAccountIDRequired,
		},
		{
			name: "missing tenant",
			params: NewServiceAccountParams{
				ID:          "sa-1",
				DisplayName: "Platform Bot",
				Now:         now,
			},
			wantErr: ErrTenantIDRequired,
		},
		{
			name: "missing display name",
			params: NewServiceAccountParams{
				ID:       "sa-1",
				TenantID: "tenant-1",
				Now:      now,
			},
			wantErr: ErrServiceAccountDisplayNameRequired,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			account, err := NewServiceAccount(tt.params)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if account.ID != tt.params.ID {
				t.Fatalf("expected id %q, got %q", tt.params.ID, account.ID)
			}
			if account.OperationID == "" {
				t.Fatal("expected operation id to be generated")
			}
			if !account.CreatedAt.Equal(now) || !account.UpdatedAt.Equal(now) {
				t.Fatalf("expected timestamps to equal %s", now)
			}
		})
	}
}
