package postgres

import (
	"context"
)

type HierarchyReader struct {
	Store *Store
}

func (r HierarchyReader) HasActiveWorkspaces(_ context.Context, organizationID string) (bool, error) {
	return false, nil
}
