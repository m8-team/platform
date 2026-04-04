package ydb

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/m8platform/platform/iam/internal/config"
	"github.com/m8platform/platform/iam/internal/core"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
)

var ErrNotConfigured = errors.New("ydb document store is not configured")

type Client struct {
	driver       *ydb.Driver
	databaseName string
}

func Open(ctx context.Context, cfg config.YDBConfig) (*Client, error) {
	if cfg.DSN == "" {
		return &Client{}, nil
	}
	driver, err := ydb.Open(ctx, cfg.DSN)
	if err != nil {
		return nil, err
	}
	return &Client{driver: driver, databaseName: driver.Name()}, nil
}

func (c *Client) Close(ctx context.Context) error {
	if c == nil || c.driver == nil {
		return nil
	}
	return c.driver.Close(ctx)
}

func (c *Client) GetDocument(ctx context.Context, table string, id string) (core.StoredDocument, error) {
	if c.driver == nil {
		return core.StoredDocument{}, ErrNotConfigured
	}

	documents, err := c.queryDocuments(ctx, table, `
DECLARE $id AS Utf8;

SELECT id, tenant_id, payload, created_at, updated_at
FROM `+quoteTable(table)+`
WHERE id = $id
ORDER BY tenant_id, id
LIMIT 2;
`, query.WithParameters(
		ydb.ParamsBuilder().
			Param("$id").Text(id).
			Build(),
	))
	if err != nil {
		return core.StoredDocument{}, err
	}
	switch len(documents) {
	case 0:
		return core.StoredDocument{}, core.ErrNotFound
	case 1:
		return documents[0], nil
	default:
		return core.StoredDocument{}, fmt.Errorf("document %q is not unique in %s", id, table)
	}
}

func (c *Client) UpsertDocument(ctx context.Context, table string, doc core.StoredDocument) error {
	if c.driver == nil {
		return ErrNotConfigured
	}

	createdAt := doc.CreatedAt.UTC()
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	updatedAt := doc.UpdatedAt.UTC()
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}

	sql := `
DECLARE $id AS Utf8;
DECLARE $tenant_id AS Utf8;
DECLARE $payload AS JsonDocument;
DECLARE $created_at AS Timestamp;
DECLARE $updated_at AS Timestamp;

UPSERT INTO ` + quoteTable(table) + ` (id, tenant_id, payload, created_at, updated_at)
VALUES ($id, $tenant_id, $payload, $created_at, $updated_at);
`
	return c.driver.Query().Exec(ctx, c.prefixedSQL(sql),
		query.WithParameters(
			ydb.ParamsBuilder().
				Param("$id").Text(doc.ID).
				Param("$tenant_id").Text(doc.TenantID).
				Param("$payload").JSONDocumentFromBytes(doc.Payload).
				Param("$created_at").Timestamp(createdAt).
				Param("$updated_at").Timestamp(updatedAt).
				Build(),
		),
	)
}

func (c *Client) DeleteDocument(ctx context.Context, table string, id string) error {
	if c.driver == nil {
		return ErrNotConfigured
	}

	document, err := c.GetDocument(ctx, table, id)
	if err != nil {
		return err
	}

	sql := `
DECLARE $id AS Utf8;
DECLARE $tenant_id AS Utf8;

DELETE FROM ` + quoteTable(table) + `
WHERE tenant_id = $tenant_id AND id = $id;
`
	return c.driver.Query().Exec(ctx, c.prefixedSQL(sql),
		query.WithParameters(
			ydb.ParamsBuilder().
				Param("$id").Text(document.ID).
				Param("$tenant_id").Text(document.TenantID).
				Build(),
		),
	)
}

func (c *Client) ListDocuments(ctx context.Context, table string, tenantID string, offset int, limit int) ([]core.StoredDocument, string, error) {
	if c.driver == nil {
		return nil, "", ErrNotConfigured
	}
	if limit <= 0 {
		limit = core.DefaultPageSize
	}

	sql := `
SELECT id, tenant_id, payload, created_at, updated_at
FROM ` + quoteTable(table)
	options := make([]query.ExecuteOption, 0, 1)
	if tenantID != "" {
		sql += `
WHERE tenant_id = $tenant_id`
		options = append(options, query.WithParameters(
			ydb.ParamsBuilder().
				Param("$tenant_id").Text(tenantID).
				Build(),
		))
	}
	sql += `
ORDER BY tenant_id, id;
`

	documents, err := c.queryDocuments(ctx, table, sql, options...)
	if err != nil {
		return nil, "", err
	}
	if offset >= len(documents) {
		return []core.StoredDocument{}, "", nil
	}

	end := offset + limit
	if end > len(documents) {
		end = len(documents)
	}
	next := ""
	if end < len(documents) {
		next = core.EncodePageToken(end)
	}
	return documents[offset:end], next, nil
}

func (c *Client) queryDocuments(ctx context.Context, table string, sql string, options ...query.ExecuteOption) ([]core.StoredDocument, error) {
	resultSet, err := c.driver.Query().QueryResultSet(ctx, c.prefixedSQL(sql), options...)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resultSet.Close(ctx)
	}()

	documents := make([]core.StoredDocument, 0)
	for row, rowErr := range resultSet.Rows(ctx) {
		if rowErr != nil {
			return nil, rowErr
		}

		var (
			id        string
			tenantID  string
			payload   string
			createdAt time.Time
			updatedAt time.Time
		)
		if err := row.ScanNamed(
			query.Named("id", &id),
			query.Named("tenant_id", &tenantID),
			query.Named("payload", &payload),
			query.Named("created_at", &createdAt),
			query.Named("updated_at", &updatedAt),
		); err != nil {
			return nil, fmt.Errorf("%s: %w", table, err)
		}

		documents = append(documents, core.StoredDocument{
			ID:        id,
			TenantID:  tenantID,
			Payload:   []byte(payload),
			CreatedAt: createdAt.UTC(),
			UpdatedAt: updatedAt.UTC(),
		})
	}
	return documents, nil
}

func (c *Client) prefixedSQL(sql string) string {
	return fmt.Sprintf("PRAGMA TablePathPrefix(%q);\n\n%s", c.databaseName, strings.TrimSpace(sql))
}

func quoteTable(table string) string {
	return "`" + strings.ReplaceAll(table, "`", "") + "`"
}
