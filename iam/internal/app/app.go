package app

import (
	"context"

	foundationapp "github.com/m8platform/platform/iam/internal/foundation/app"
	foundationconfig "github.com/m8platform/platform/iam/internal/foundation/config"
	"github.com/m8platform/platform/iam/internal/infrastructure/bootstrap"
)

type Application = foundationapp.Application

func New(ctx context.Context, cfg foundationconfig.Config) (*Application, error) {
	return bootstrap.NewApplication(ctx, cfg)
}
