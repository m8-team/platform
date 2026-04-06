package port

import (
	"context"
	"errors"

	"github.com/m8platform/platform/iam/internal/module/authz/entity"
	"github.com/m8platform/platform/iam/internal/module/authz/model"
	"github.com/m8platform/platform/iam/internal/shared/resource"
)

var ErrAuthorizationUnavailable = errors.New("authorization runtime is unavailable")

type AccessBindingRepository interface {
	ListByResource(ctx context.Context, resource resource.Ref) ([]entity.AccessBinding, error)
}

type AuthorizationChecker interface {
	CheckAccess(ctx context.Context, query model.AccessCheckQuery) (model.AccessCheckResult, error)
}

type AccessDecisionCache interface {
	GetAccessDecision(ctx context.Context, query model.AccessCheckQuery) (model.AccessCheckResult, bool, error)
	SaveAccessDecision(ctx context.Context, query model.AccessCheckQuery, result model.AccessCheckResult) error
}

type RolePermissionResolver interface {
	Permissions(roleID string) []string
}
