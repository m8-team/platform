// Package spicedb adapts the SpiceDB HTTP API to M8 Access application ports.
// It does not own role/grant lifecycles, schema publication or tenant identity.
package spicedb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/domain"
)

var (
	ErrInvalidConfig          = errors.New("invalid SpiceDB adapter configuration")
	ErrInvalidResponse        = errors.New("invalid SpiceDB response")
	ErrSchemaRevisionMismatch = errors.New("SpiceDB adapter model revision mismatch")
	ErrInvalidGraphReference  = errors.New("invalid SpiceDB graph reference")
)

const maxMessageBytes = 1 << 20

// Config is a pinned, independently reviewed graph schema deployment.
// ModelRevision is an M8 schema marker; it is NOT a ZedToken and is NOT
// automatically verified against the live SpiceDB schema.
type Config struct {
	Endpoint              string
	BearerToken           string
	ModelRevision         string
	AllowInsecureLoopback bool // for local tests only; never for deployment
}

// Client implements PermissionEngine and RelationshipWriter using SpiceDB's
// native HTTP gateway. Enable it on SpiceDB with --http-enabled.
type Client struct {
	endpoint   string
	token      string
	revision   domain.ModelRevision
	catalog    ports.PermissionCatalog
	httpClient *http.Client
}

var _ ports.PermissionEngine = (*Client)(nil)
var _ ports.RelationshipWriter = (*Client)(nil)

func NewClient(cfg Config, catalog ports.PermissionCatalog, transport *http.Client) (*Client, error) {
	if catalog == nil {
		return nil, ErrInvalidConfig
	}
	u, err := url.Parse(cfg.Endpoint)
	if err != nil || u == nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return nil, ErrInvalidConfig
	}
	if u.Scheme != "https" {
		if u.Scheme != "http" || !cfg.AllowInsecureLoopback || !isLoopback(u.Hostname()) {
			return nil, ErrInvalidConfig
		}
	}
	if strings.TrimSpace(cfg.BearerToken) == "" || strings.ContainsAny(cfg.BearerToken, "\r\n") {
		return nil, ErrInvalidConfig
	}
	rev, err := domain.NewModelRevision(strings.TrimSpace(cfg.ModelRevision))
	if err != nil {
		return nil, ErrInvalidConfig
	}
	if transport == nil {
		transport = &http.Client{Timeout: 5 * time.Second}
	}
	copyTransport := *transport
	// A redirect must never forward a SpiceDB bearer token to another host.
	copyTransport.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
	return &Client{endpoint: strings.TrimRight(cfg.Endpoint, "/"), token: cfg.BearerToken, revision: rev, catalog: catalog, httpClient: &copyTransport}, nil
}

func isLoopback(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func (c *Client) post(ctx context.Context, path string, payload any, result any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal SpiceDB request: %w", err)
	}
	if len(b) > maxMessageBytes {
		return ErrInvalidGraphReference
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+path, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("build SpiceDB request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("%w: deadline", ports.ErrPermissionEngineTimeout)
		}
		if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
			return context.Canceled
		}
		if timeout, ok := err.(net.Error); ok && timeout.Timeout() {
			return fmt.Errorf("%w: timeout", ports.ErrPermissionEngineTimeout)
		}
		return fmt.Errorf("%w: transport failure", ports.ErrPermissionEngineUnavailable)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable || resp.StatusCode == http.StatusBadGateway || resp.StatusCode == http.StatusGatewayTimeout {
			return fmt.Errorf("%w: HTTP %d", ports.ErrPermissionEngineUnavailable, resp.StatusCode)
		}
		return fmt.Errorf("SpiceDB rejected request: HTTP %d", resp.StatusCode)
	}
	limited := io.LimitReader(resp.Body, maxMessageBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return fmt.Errorf("%w: unreadable response", ErrInvalidResponse)
	}
	if len(body) > maxMessageBytes {
		return fmt.Errorf("%w: oversized response", ErrInvalidResponse)
	}
	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("%w: malformed JSON", ErrInvalidResponse)
	}
	return nil
}
