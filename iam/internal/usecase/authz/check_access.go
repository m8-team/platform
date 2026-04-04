package authz

import (
	"context"
	"strings"

	authzentity "github.com/m8platform/platform/iam/internal/entity/authz"
	"github.com/m8platform/platform/iam/internal/usecase/model"
	"github.com/m8platform/platform/iam/internal/usecase/port"
)

type CheckAccessUseCase struct {
	bindings port.AccessBindingRepository
	checker  port.AuthorizationChecker
	cache    port.AccessDecisionCache
	roles    port.RolePermissionResolver
}

func NewCheckAccessUseCase(
	bindings port.AccessBindingRepository,
	checker port.AuthorizationChecker,
	cache port.AccessDecisionCache,
	roles port.RolePermissionResolver,
) *CheckAccessUseCase {
	return &CheckAccessUseCase{
		bindings: bindings,
		checker:  checker,
		cache:    cache,
		roles:    roles,
	}
}

func (u *CheckAccessUseCase) Execute(ctx context.Context, query model.AccessCheckQuery) (model.AccessCheckResult, error) {
	if u.cache != nil && u.checker == nil {
		if cached, ok, err := u.cache.GetAccessDecision(ctx, query); err == nil && ok {
			cached.CacheHit = true
			return cached, nil
		}
	}

	if u.checker != nil {
		if runtimeResult, err := u.checker.CheckAccess(ctx, query); err == nil {
			return runtimeResult, nil
		}
	}

	bindings, err := u.bindings.ListByResource(ctx, query.Resource)
	if err != nil {
		return model.AccessCheckResult{}, err
	}

	decision := authzentity.PermissionDecisionDeny
	permission := strings.TrimSpace(query.Permission)
	for _, binding := range bindings {
		if binding.Grants(query.Subject, permission, u.roles.Permissions(binding.RoleID)) {
			decision = authzentity.PermissionDecisionAllow
			break
		}
	}

	result := model.AccessCheckResult{
		Decision:   decision,
		Permission: permission,
		ZedToken:   "fallback",
	}
	if u.cache != nil && u.checker == nil {
		_ = u.cache.SaveAccessDecision(ctx, query, result)
	}
	return result, nil
}
