package usecase

import (
	"context"
	"testing"

	"github.com/m8platform/platform/iam/internal/module/authz/entity"
	"github.com/m8platform/platform/iam/internal/module/authz/model"
	"github.com/m8platform/platform/iam/internal/shared/principal"
	"github.com/m8platform/platform/iam/internal/shared/resource"
)

type accessBindingRepositoryFake struct {
	bindings []entity.AccessBinding
}

func (f accessBindingRepositoryFake) ListByResource(_ context.Context, _ resource.Ref) ([]entity.AccessBinding, error) {
	return f.bindings, nil
}

type authorizationCheckerFake struct {
	result model.AccessCheckResult
	err    error
}

func (f authorizationCheckerFake) CheckAccess(_ context.Context, _ model.AccessCheckQuery) (model.AccessCheckResult, error) {
	return f.result, f.err
}

type accessDecisionCacheFake struct {
	data map[string]model.AccessCheckResult
}

func (f *accessDecisionCacheFake) GetAccessDecision(_ context.Context, query model.AccessCheckQuery) (model.AccessCheckResult, bool, error) {
	result, ok := f.data[query.Permission]
	return result, ok, nil
}

func (f *accessDecisionCacheFake) SaveAccessDecision(_ context.Context, query model.AccessCheckQuery, result model.AccessCheckResult) error {
	if f.data == nil {
		f.data = make(map[string]model.AccessCheckResult)
	}
	f.data[query.Permission] = result
	return nil
}

type rolePermissionResolverFake struct {
	permissions map[string][]string
}

func (f rolePermissionResolverFake) Permissions(roleID string) []string {
	return f.permissions[roleID]
}

func TestCheckAccessUseCaseExecuteFallsBackToBindingsAndCachesResult(t *testing.T) {
	t.Parallel()

	cache := &accessDecisionCacheFake{}
	useCase := NewCheckAccessUseCase(
		accessBindingRepositoryFake{
			bindings: []entity.AccessBinding{
				{
					ID:     "binding-1",
					RoleID: "tenant-owner",
					Subject: principal.Principal{
						Type: "SUBJECT_TYPE_USER_ACCOUNT",
						ID:   "user-1",
					},
					Resource: resource.Ref{
						Type: "RESOURCE_TYPE_TENANT",
						ID:   "tenant-1",
					},
				},
			},
		},
		nil,
		cache,
		rolePermissionResolverFake{
			permissions: map[string][]string{
				"tenant-owner": {"tenant.manage"},
			},
		},
	)

	query := model.AccessCheckQuery{
		Subject: principal.Principal{
			Type: "SUBJECT_TYPE_USER_ACCOUNT",
			ID:   "user-1",
		},
		Resource: resource.Ref{
			Type: "RESOURCE_TYPE_TENANT",
			ID:   "tenant-1",
		},
		Permission: "tenant.manage",
	}

	result, err := useCase.Execute(context.Background(), query)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision != entity.PermissionDecisionAllow {
		t.Fatalf("expected allow decision, got %s", result.Decision)
	}
	if _, ok := cache.data[query.Permission]; !ok {
		t.Fatal("expected fallback result to be cached")
	}

	cached, err := useCase.Execute(context.Background(), query)
	if err != nil {
		t.Fatalf("unexpected cached error: %v", err)
	}
	if !cached.CacheHit {
		t.Fatal("expected cached result")
	}
}

func TestCheckAccessUseCaseExecuteUsesRuntimeResult(t *testing.T) {
	t.Parallel()

	useCase := NewCheckAccessUseCase(
		accessBindingRepositoryFake{},
		authorizationCheckerFake{
			result: model.AccessCheckResult{
				Decision:   entity.PermissionDecisionConditional,
				Permission: "project.read",
				ZedToken:   "zed-1",
			},
		},
		nil,
		rolePermissionResolverFake{},
	)

	result, err := useCase.Execute(context.Background(), model.AccessCheckQuery{
		Permission: "project.read",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision != entity.PermissionDecisionConditional {
		t.Fatalf("expected conditional decision, got %s", result.Decision)
	}
	if result.ZedToken != "zed-1" {
		t.Fatalf("expected zed token zed-1, got %q", result.ZedToken)
	}
}
