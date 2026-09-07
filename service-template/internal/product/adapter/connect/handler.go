package connectadapter

import (
	"connectrpc.com/connect"
	"context"
	"github.com/m8-team/platform-sdk/metadata"
	productv1 "github.com/m8-team/product-service/internal/gen/product/v1"
	productapp "github.com/m8-team/product-service/internal/product/app"
	"github.com/m8-team/product-service/internal/product/domain"
)

type Handler struct{ Service *productapp.Service }

func actor(ctx context.Context) productapp.Actor {
	p, _ := metadata.PrincipalFrom(ctx)
	return productapp.Actor{Subject: p.SubjectID, Tenant: domain.TenantID(p.TenantID), Roles: p.Roles}
}
func (h *Handler) GetProduct(ctx context.Context, req *connect.Request[productv1.GetProductRequest]) (*connect.Response[productv1.GetProductResponse], error) {
	p, err := h.Service.Get(ctx, actor(ctx), domain.ID(req.Msg.GetId()))
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&productv1.GetProductResponse{Product: &productv1.Product{Id: string(p.ID), Name: p.Name}}), nil
}
func (h *Handler) CreateProduct(ctx context.Context, req *connect.Request[productv1.CreateProductRequest]) (*connect.Response[productv1.CreateProductResponse], error) {
	p, err := h.Service.Create(ctx, actor(ctx), req.Msg.GetName(), req.Msg.GetIdempotencyKey())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&productv1.CreateProductResponse{Product: &productv1.Product{Id: string(p.ID), Name: p.Name}}), nil
}
