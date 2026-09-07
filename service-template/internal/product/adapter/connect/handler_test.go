package connectadapter

import (
	"connectrpc.com/connect"
	"context"
	sdkconnect "github.com/m8-team/platform-sdk/connect"
	"github.com/m8-team/platform-sdk/logging"
	"github.com/m8-team/platform-sdk/telemetry"
	productv1 "github.com/m8-team/product-service/internal/gen/product/v1"
	"github.com/m8-team/product-service/internal/gen/product/v1/productv1connect"
	"github.com/m8-team/product-service/internal/platform"
	productapp "github.com/m8-team/product-service/internal/product/app"
	"github.com/m8-team/product-service/internal/product/domain"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"
)

type repository struct{}

func (repository) Get(_ context.Context, tenant domain.TenantID, id domain.ID) (domain.Product, error) {
	return domain.Product{ID: id, Tenant: tenant, Name: "Coffee"}, nil
}
func (repository) Create(_ context.Context, p domain.Product, _ string, _ [32]byte) (domain.Product, error) {
	return p, nil
}
func TestUnaryWire(t *testing.T) {
	ctx := context.Background()
	providers, err := telemetry.New(ctx, logging.Info{Name: "test"}, telemetry.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer providers.Shutdown(ctx)
	auth := platform.Identity{Local: true, Token: "test-token"}
	interceptor, err := sdkconnect.New(sdkconnect.Config{Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Telemetry: providers, Authentication: auth, Authorization: platform.TransportPolicy{}})
	if err != nil {
		t.Fatal(err)
	}
	service, err := productapp.New(repository{}, platform.Policy{}, func() domain.ID { return "prd_test" })
	if err != nil {
		t.Fatal(err)
	}
	_, handler := productv1connect.NewProductServiceHandler(&Handler{Service: service}, connect.WithInterceptors(interceptor))
	server := httptest.NewServer(handler)
	defer server.Close()
	client := productv1connect.NewProductServiceClient(server.Client(), server.URL, connect.WithHTTPGet())
	get := connect.NewRequest(&productv1.GetProductRequest{Id: "prd_test"})
	if _, err = client.GetProduct(ctx, get); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("unauthenticated: %v", err)
	}
	get.Header().Set("Authorization", "Bearer test-token")
	response, err := client.GetProduct(ctx, get)
	if err != nil {
		t.Fatal(err)
	}
	if response.Msg.GetProduct().GetName() != "Coffee" || response.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing result or correlation")
	}
	create := connect.NewRequest(&productv1.CreateProductRequest{Name: "Coffee", IdempotencyKey: "too-short"})
	create.Header().Set("Authorization", "Bearer test-token")
	if _, err = client.CreateProduct(ctx, create); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("validation: %v", err)
	}
	create.Msg.IdempotencyKey = "1234567890abcdef"
	if _, err = client.CreateProduct(ctx, create); err != nil {
		t.Fatal(err)
	}
}
