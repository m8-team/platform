package authz

import (
	"context"
	"errors"
	"testing"

	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
)

func TestStaticAuthorizerIsDenyByDefault(t *testing.T) {
	request := ports.AuthorizationRequest{Action: ports.ActionListOrganizations}
	for name, authorizer := range map[string]*StaticAuthorizer{
		"constructor": DenyAll(),
		"zero value":  {},
		"nil pointer": nil,
	} {
		t.Run(name, func(t *testing.T) {
			if err := authorizer.Authorize(context.Background(), request); !errors.Is(err, ports.ErrPermissionDenied) {
				t.Fatalf("Authorize() error = %v, want %v", err, ports.ErrPermissionDenied)
			}
		})
	}
}

func TestStaticAuthorizerExplicitAllowAndContext(t *testing.T) {
	request := ports.AuthorizationRequest{Action: ports.ActionListOrganizations}
	if err := AllowAll().Authorize(context.Background(), request); err != nil {
		t.Fatalf("Authorize() error = %v", err)
	}
	if scope, err := AllowAll().ScopeKey(context.Background()); err != nil || scope != allowAllScopeKey {
		t.Fatalf("ScopeKey() = %q, %v", scope, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := AllowAll().Authorize(ctx, request); !errors.Is(err, context.Canceled) {
		t.Fatalf("Authorize(canceled) error = %v, want %v", err, context.Canceled)
	}
}
