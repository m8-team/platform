package api

import (
	"context"
	"errors"

	"github.com/m8platform/platform/iam/internal/module/authz/model"
	authzuc "github.com/m8platform/platform/iam/internal/module/authz/usecase"
)

var ErrFacadeUnavailable = errors.New("authz facade is unavailable")

type Facade interface {
	CheckPermission(ctx context.Context, query model.AccessCheckQuery) (model.AccessCheckResult, error)
}

type Service struct {
	check *authzuc.CheckAccessUseCase
}

func New(check *authzuc.CheckAccessUseCase) *Service {
	return &Service{check: check}
}

func (s *Service) CheckPermission(ctx context.Context, query model.AccessCheckQuery) (model.AccessCheckResult, error) {
	if s == nil || s.check == nil {
		return model.AccessCheckResult{}, ErrFacadeUnavailable
	}
	return s.check.Execute(ctx, query)
}
