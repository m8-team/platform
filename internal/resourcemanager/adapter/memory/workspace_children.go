package memory

import (
	"context"

	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
)

// WorkspaceChildren is the local-development hierarchy projection. The
// current process does not host WorkspaceService yet, so it has no children.
// Production composition must replace this adapter with the authoritative
// workspace repository/projection.
type WorkspaceChildren struct{}

func NewWorkspaceChildren() *WorkspaceChildren {
	return &WorkspaceChildren{}
}

func (*WorkspaceChildren) HasNonDeleted(ctx context.Context, _ organization.ID) (bool, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return false, err
		}
	}
	return false, nil
}
