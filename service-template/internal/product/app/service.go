package app

import (
	"context"
	"crypto/sha256"
	"errors"
	"github.com/m8-team/platform-sdk/fault"
	"github.com/m8-team/product-service/internal/product/domain"
)

type Repository interface {
	Get(context.Context, domain.TenantID, domain.ID) (domain.Product, error)
	Create(context.Context, domain.Product, string, [32]byte) (domain.Product, error)
}
type Actor struct {
	Subject string
	Tenant  domain.TenantID
	Roles   []string
}
type Authorizer interface {
	Check(context.Context, Actor, string) error
}
type Service struct {
	repo  Repository
	auth  Authorizer
	newID func() domain.ID
}

func New(repo Repository, auth Authorizer, newID func() domain.ID) (*Service, error) {
	if repo == nil || auth == nil || newID == nil {
		return nil, errors.New("product dependencies required")
	}
	return &Service{repo, auth, newID}, nil
}
func (s *Service) Get(ctx context.Context, actor Actor, id domain.ID) (domain.Product, error) {
	if err := s.authorize(ctx, actor, "product.read"); err != nil {
		return domain.Product{}, err
	}
	if id == "" {
		return domain.Product{}, fault.New(fault.InvalidArgument, "product id required", nil)
	}
	return s.repo.Get(ctx, actor.Tenant, id)
}
func (s *Service) Create(ctx context.Context, actor Actor, name, key string) (domain.Product, error) {
	if err := s.authorize(ctx, actor, "product.create"); err != nil {
		return domain.Product{}, err
	}
	if len(key) < 16 || len(key) > 128 {
		return domain.Product{}, fault.New(fault.InvalidArgument, "idempotency key must contain 16 to 128 bytes", nil)
	}
	product, err := domain.New(s.newID(), actor.Tenant, name)
	if err != nil {
		return domain.Product{}, fault.New(fault.InvalidArgument, "invalid product", err)
	}
	return s.repo.Create(ctx, product, key, sha256.Sum256([]byte(name)))
}
func (s *Service) authorize(ctx context.Context, actor Actor, operation string) error {
	if actor.Subject == "" || actor.Tenant == "" {
		return fault.New(fault.Unauthenticated, "authentication required", nil)
	}
	return s.auth.Check(ctx, actor, operation)
}
