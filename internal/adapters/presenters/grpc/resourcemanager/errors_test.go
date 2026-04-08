package grpcpresenter

import (
	"testing"

	organizationentity "github.com/m8platform/platform/internal/entities/resourcemanager/organization"
	workspaceentity "github.com/m8platform/platform/internal/entities/resourcemanager/workspace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestPresentErrorMapsCodes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		err  error
		code codes.Code
	}{
		{name: "not found", err: organizationentity.ErrNotFound, code: codes.NotFound},
		{name: "etag mismatch", err: workspaceentity.ErrETagMismatch, code: codes.Aborted},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := PresentError(tc.err)
			if status.Code(got) != tc.code {
				t.Fatalf("expected code %s, got %s", tc.code, status.Code(got))
			}
		})
	}
}
