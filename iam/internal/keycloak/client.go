package keycloak

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/m8platform/platform/iam/internal/config"
)

var ErrNotConfigured = errors.New("keycloak client is not configured")

type Client struct {
	cfg config.KeycloakConfig
}

func NewClient(cfg config.KeycloakConfig) *Client {
	return &Client{cfg: cfg}
}

func (c *Client) CreateConfidentialClient(_ context.Context, tenantID string, clientID string, displayName string, serviceAccountsEnabled bool) (string, error) {
	if c.cfg.BaseURL == "" {
		return "", ErrNotConfigured
	}
	return fmt.Sprintf("%s:%s:%s:%t", tenantID, clientID, displayName, serviceAccountsEnabled), nil
}

func (c *Client) RotateClientSecret(_ context.Context, clientID string) (string, string, error) {
	if c.cfg.BaseURL == "" {
		return "", "", ErrNotConfigured
	}
	secretRef := fmt.Sprintf("vault://keycloak/%s/%s", c.cfg.Realm, clientID)
	return uuid.NewString(), secretRef, nil
}
