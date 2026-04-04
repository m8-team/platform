package spicedb

import (
	"context"
	"errors"

	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
	"github.com/m8platform/platform/iam/internal/config"
)

var ErrNotConfigured = errors.New("spicedb endpoint is not configured")
var ErrNotImplemented = errors.New("spicedb runtime integration is pending")

type Client struct {
	cfg config.SpiceDBConfig
}

func NewClient(cfg config.SpiceDBConfig) *Client {
	return &Client{cfg: cfg}
}

func (c *Client) Check(_ context.Context, req *authzv1.CheckAccessRequest) (*authzv1.AccessCheckResult, error) {
	if c.cfg.Endpoint == "" {
		return nil, ErrNotConfigured
	}
	return nil, ErrNotImplemented
}

func (c *Client) WriteBindings(_ context.Context, _ []*authzv1.AccessBinding) error {
	if c.cfg.Endpoint == "" {
		return ErrNotConfigured
	}
	return ErrNotImplemented
}
