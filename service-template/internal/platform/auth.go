package platform

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	sdkconnect "github.com/m8-team/platform-sdk/connect"
	"github.com/m8-team/platform-sdk/fault"
	productapp "github.com/m8-team/product-service/internal/product/app"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Identity adapts RFC 7662 introspection with company tenant_id/roles extensions.
// No token is logged or cached. Client timeout and HTTPS are set at composition.
type Identity struct {
	Local                                   bool
	Token, Endpoint, ClientID, ClientSecret string
	Client                                  *http.Client
}

func (i Identity) Authenticate(ctx context.Context, headers http.Header) (sdkconnect.Principal, error) {
	raw := headers.Get("Authorization")
	if !strings.HasPrefix(raw, "Bearer ") || len(raw) > 8192 {
		return sdkconnect.Principal{}, fault.New(fault.Unauthenticated, "invalid credentials", nil)
	}
	token := strings.TrimPrefix(raw, "Bearer ")
	if i.Local {
		if subtle.ConstantTimeCompare([]byte(token), []byte(i.Token)) != 1 {
			return sdkconnect.Principal{}, fault.New(fault.Unauthenticated, "invalid credentials", nil)
		}
		return sdkconnect.Principal{SubjectID: "local-user", TenantID: "local-tenant", Roles: []string{"product.read", "product.create"}}, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, i.Endpoint, strings.NewReader(url.Values{"token": {token}, "token_type_hint": {"access_token"}}.Encode()))
	if err != nil {
		return sdkconnect.Principal{}, errors.New("construct introspection request")
	}
	req.SetBasicAuth(i.ClientID, i.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := i.Client.Do(req)
	if err != nil {
		return sdkconnect.Principal{}, fault.New(fault.Unavailable, "identity provider unavailable", nil)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return sdkconnect.Principal{}, fault.New(fault.Unavailable, "identity provider unavailable", nil)
	}
	var claims struct {
		Active  bool     `json:"active"`
		Subject string   `json:"sub"`
		Tenant  string   `json:"tenant_id"`
		Roles   []string `json:"roles"`
	}
	if err = json.NewDecoder(io.LimitReader(response.Body, 64<<10)).Decode(&claims); err != nil {
		return sdkconnect.Principal{}, fault.New(fault.Unavailable, "invalid identity provider response", nil)
	}
	if !claims.Active || claims.Subject == "" || claims.Tenant == "" {
		return sdkconnect.Principal{}, fault.New(fault.Unauthenticated, "invalid credentials", nil)
	}
	return sdkconnect.Principal{SubjectID: claims.Subject, TenantID: claims.Tenant, Roles: claims.Roles}, nil
}

type Policy struct{}

func (Policy) Check(_ context.Context, actor productapp.Actor, operation string) error {
	for _, role := range actor.Roles {
		if role == operation {
			return nil
		}
	}
	return fault.New(fault.PermissionDenied, "permission denied", nil)
}

type TransportPolicy struct{}

func (TransportPolicy) Check(ctx context.Context, req sdkconnect.CheckRequest) error {
	operation := ""
	switch req.Procedure {
	case "/product.v1.ProductService/GetProduct":
		operation = "product.read"
	case "/product.v1.ProductService/CreateProduct":
		operation = "product.create"
	default:
		return fault.New(fault.PermissionDenied, "permission denied", nil)
	}
	return (Policy{}).Check(ctx, productapp.Actor{Roles: req.Principal.Roles}, operation)
}
