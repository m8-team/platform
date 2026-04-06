package api

import (
	"context"
	"errors"

	"github.com/m8platform/platform/iam/internal/module/iam/model"
	identityuc "github.com/m8platform/platform/iam/internal/module/iam/usecase"
)

var ErrFacadeUnavailable = errors.New("iam facade is unavailable")

type Facade interface {
	CreateServiceAccount(ctx context.Context, cmd model.CreateServiceAccountCommand) (model.CreateServiceAccountResult, error)
	RotateOAuthClientSecret(ctx context.Context, cmd model.RotateOAuthClientSecretCommand) (model.RotateOAuthClientSecretResult, error)
}

type Service struct {
	create *identityuc.CreateServiceAccountUseCase
	rotate *identityuc.RotateOAuthClientSecretUseCase
}

func New(create *identityuc.CreateServiceAccountUseCase, rotate *identityuc.RotateOAuthClientSecretUseCase) *Service {
	return &Service{
		create: create,
		rotate: rotate,
	}
}

func (s *Service) CreateServiceAccount(ctx context.Context, cmd model.CreateServiceAccountCommand) (model.CreateServiceAccountResult, error) {
	if s == nil || s.create == nil {
		return model.CreateServiceAccountResult{}, ErrFacadeUnavailable
	}
	return s.create.Execute(ctx, cmd)
}

func (s *Service) RotateOAuthClientSecret(ctx context.Context, cmd model.RotateOAuthClientSecretCommand) (model.RotateOAuthClientSecretResult, error) {
	if s == nil || s.rotate == nil {
		return model.RotateOAuthClientSecretResult{}, ErrFacadeUnavailable
	}
	return s.rotate.Execute(ctx, cmd)
}
