package keycloak

import (
	"context"

	legacykeycloak "github.com/m8platform/platform/iam/internal/keycloak"
)

type Client struct {
	client *legacykeycloak.Client
}

func NewClient(client *legacykeycloak.Client) *Client {
	return &Client{client: client}
}

func (c *Client) CreateConfidentialClient(ctx context.Context, tenantID string, clientID string, displayName string, serviceAccountsEnabled bool) (string, error) {
	if c == nil || c.client == nil {
		return "", nil
	}
	return c.client.CreateConfidentialClient(ctx, tenantID, clientID, displayName, serviceAccountsEnabled)
}

func (c *Client) RotateOAuthClientSecret(ctx context.Context, oauthClientID string) (string, error) {
	if c == nil || c.client == nil {
		return "", nil
	}
	_, secretRef, err := c.client.RotateClientSecret(ctx, oauthClientID)
	if err != nil {
		return "", err
	}
	return secretRef, nil
}
