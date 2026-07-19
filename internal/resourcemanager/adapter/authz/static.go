package authz

import (
	"context"
	"fmt"

	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
)

// StaticAuthorizer is intentionally small: DenyAll is the secure production
// default, while AllowAll must only be selected explicitly for local
// development and tests until the Access adapter is wired.
type StaticAuthorizer struct {
	allow bool
}

const allowAllScopeKey = "static:allow-all"

func DenyAll() *StaticAuthorizer {
	return &StaticAuthorizer{}
}

func AllowAll() *StaticAuthorizer {
	return &StaticAuthorizer{allow: true}
}

func (a *StaticAuthorizer) Authorize(ctx context.Context, request ports.AuthorizationRequest) error {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	if a != nil && a.allow {
		return nil
	}

	return fmt.Errorf("%w: action=%s organization_id=%s", ports.ErrPermissionDenied, request.Action, request.OrganizationID)
}

func (a *StaticAuthorizer) ScopeKey(ctx context.Context) (string, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return "", err
		}
	}
	if a == nil || !a.allow {
		return "", ports.ErrPermissionDenied
	}
	return allowAllScopeKey, nil
}
