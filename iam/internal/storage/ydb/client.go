package ydb

import (
	"context"
	"errors"

	"github.com/m8platform/platform/iam/internal/config"
	"github.com/m8platform/platform/iam/internal/core"
	"github.com/ydb-platform/ydb-go-sdk/v3"
)

var ErrNotImplemented = errors.New("ydb document store implementation is pending")

type Client struct {
	driver *ydb.Driver
}

func Open(ctx context.Context, cfg config.YDBConfig) (*Client, error) {
	if cfg.DSN == "" {
		return &Client{}, nil
	}
	driver, err := ydb.Open(ctx, cfg.DSN)
	if err != nil {
		return nil, err
	}
	return &Client{driver: driver}, nil
}

func (c *Client) Close(ctx context.Context) error {
	if c == nil || c.driver == nil {
		return nil
	}
	return c.driver.Close(ctx)
}

func (c *Client) GetDocument(_ context.Context, _ string, _ string) (core.StoredDocument, error) {
	return core.StoredDocument{}, ErrNotImplemented
}

func (c *Client) UpsertDocument(_ context.Context, _ string, _ core.StoredDocument) error {
	return ErrNotImplemented
}

func (c *Client) DeleteDocument(_ context.Context, _ string, _ string) error {
	return ErrNotImplemented
}

func (c *Client) ListDocuments(_ context.Context, _ string, _ string, _ int, _ int) ([]core.StoredDocument, string, error) {
	return nil, "", ErrNotImplemented
}
