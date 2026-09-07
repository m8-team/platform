package app

import (
	"context"
	"errors"
	"github.com/m8-team/product-service/internal/product/domain"
	"testing"
)

type repo struct{ called bool }

func (r *repo) Get(context.Context, domain.TenantID, domain.ID) (domain.Product, error) {
	r.called = true
	return domain.Product{}, nil
}
func (r *repo) Create(_ context.Context, p domain.Product, _ string, _ [32]byte) (domain.Product, error) {
	r.called = true
	return p, nil
}

type denied struct{}

func (denied) Check(context.Context, Actor, string) error { return errors.New("denied") }
func TestAuthorizationBeforeStorage(t *testing.T) {
	r := &repo{}
	s, err := New(r, denied{}, func() domain.ID { return "id" })
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.Create(context.Background(), Actor{Subject: "s", Tenant: "t"}, "Coffee", "0123456789abcdef")
	if err == nil || r.called {
		t.Fatal("authorization bypass")
	}
}
