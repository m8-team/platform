package port

import (
	"context"

	authzentity "github.com/m8platform/platform/iam/internal/entity/authz"
	"github.com/m8platform/platform/iam/internal/usecase/model"
)

type AccessBindingRepository interface {
	ListByResource(ctx context.Context, resource authzentity.ResourceRef) ([]authzentity.AccessBinding, error)
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
