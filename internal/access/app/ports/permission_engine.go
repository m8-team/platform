package ports

import (
	"context"
	"errors"

	"github.com/m8platform/platform/internal/access/domain"
)

var (
	ErrPermissionEngineTimeout     = errors.New("permission engine timeout")
	ErrPermissionEngineUnavailable = errors.New("permission engine unavailable")
)

type PermissionEngine interface {
	CheckPermission(
		ctx context.Context,
		request domain.CheckPermissionRequest,
	) (domain.PermissionDecision, error)
}
