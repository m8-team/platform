package api

import (
	"context"
	"errors"

	tenantentity "github.com/m8platform/platform/iam/internal/entity/tenant"
	"github.com/m8platform/platform/iam/internal/usecase/model"
	tenantuc "github.com/m8platform/platform/iam/internal/usecase/tenant"
)

var ErrFacadeUnavailable = errors.New("tenant facade is unavailable")

type Facade interface {
	GrantSupportAccess(ctx context.Context, cmd model.GrantSupportAccessCommand) (model.SupportGrantResult, error)
	ApproveSupportAccess(ctx context.Context, cmd model.ApproveSupportAccessCommand) (model.SupportGrantResult, error)
	RevokeSupportAccess(ctx context.Context, cmd model.RevokeSupportAccessCommand) (model.SupportGrantResult, error)
	ListSupportGrants(ctx context.Context, query model.ListSupportGrantsQuery) ([]tenantentity.SupportGrant, string, error)
}

type Service struct {
	useCase *tenantuc.SupportAccessUseCase
}

func New(useCase *tenantuc.SupportAccessUseCase) *Service {
	return &Service{useCase: useCase}
}

func (s *Service) GrantSupportAccess(ctx context.Context, cmd model.GrantSupportAccessCommand) (model.SupportGrantResult, error) {
	if s == nil || s.useCase == nil {
		return model.SupportGrantResult{}, ErrFacadeUnavailable
	}
	return s.useCase.Grant(ctx, cmd)
}

func (s *Service) ApproveSupportAccess(ctx context.Context, cmd model.ApproveSupportAccessCommand) (model.SupportGrantResult, error) {
	if s == nil || s.useCase == nil {
		return model.SupportGrantResult{}, ErrFacadeUnavailable
	}
	return s.useCase.Approve(ctx, cmd)
}

func (s *Service) RevokeSupportAccess(ctx context.Context, cmd model.RevokeSupportAccessCommand) (model.SupportGrantResult, error) {
	if s == nil || s.useCase == nil {
		return model.SupportGrantResult{}, ErrFacadeUnavailable
	}
	return s.useCase.Revoke(ctx, cmd)
}

func (s *Service) ListSupportGrants(ctx context.Context, query model.ListSupportGrantsQuery) ([]tenantentity.SupportGrant, string, error) {
	if s == nil || s.useCase == nil {
		return nil, "", ErrFacadeUnavailable
	}
	return s.useCase.List(ctx, query)
}
