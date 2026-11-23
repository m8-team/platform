// Package spicedb adapts the official Authzed Go SDK to M8 Access ports.
// It does not own grants, authorization schema lifecycle, or tenant identity.
package spicedb

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/m8-team/platform/internal/access/app/ports"
	"github.com/m8-team/platform/internal/access/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	ErrInvalidConfig          = errors.New("invalid SpiceDB gRPC configuration")
	ErrInvalidResponse        = errors.New("invalid SpiceDB response")
	ErrRPCRejected            = errors.New("SpiceDB rejected authorization request")
	ErrSchemaRevisionMismatch = errors.New("SpiceDB adapter model revision mismatch")
	ErrInvalidGraphReference  = errors.New("invalid SpiceDB graph reference")
)

// Config describes the SpiceDB gRPC endpoint and the operator-reviewed M8 model
// revision. ModelRevision is NOT a ZedToken and is not verified by the constructor.
// It must be verified against the deployed schema during model rollout.
type Config struct {
	Endpoint              string // host:port, without scheme or path
	BearerToken           string
	ModelRevision         string
	AllowInsecureLoopback bool // local tests only; production MUST use TLS
}

// Client uses the official authzed-go SDK. The PermissionsServiceClient interface
// permits unit tests without an external SpiceDB process.
type Client struct {
	sdk         *authzed.Client
	permissions pb.PermissionsServiceClient
	revision    domain.ModelRevision
	catalog     ports.PermissionCatalog
}

var _ ports.PermissionEngine = (*Client)(nil)
var _ ports.RelationshipWriter = (*Client)(nil)

// NewClient owns a gRPC connection; callers must Close it during shutdown.
func NewClient(cfg Config, catalog ports.PermissionCatalog) (*Client, error) {
	if catalog == nil || strings.TrimSpace(cfg.BearerToken) == "" || strings.ContainsAny(cfg.BearerToken, "\r\n") {
		return nil, ErrInvalidConfig
	}
	endpoint := strings.TrimSpace(cfg.Endpoint)
	parsed, err := url.Parse("//" + endpoint)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Path != "" || parsed.Hostname() == "" {
		return nil, ErrInvalidConfig
	}
	if _, _, err := net.SplitHostPort(endpoint); err != nil {
		return nil, ErrInvalidConfig
	}
	revision, err := domain.NewModelRevision(cfg.ModelRevision)
	if err != nil {
		return nil, ErrInvalidConfig
	}
	var transport = credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})
	if cfg.AllowInsecureLoopback {
		host := parsed.Hostname()
		ip := net.ParseIP(host)
		if !strings.EqualFold(host, "localhost") && (ip == nil || !ip.IsLoopback()) {
			return nil, ErrInvalidConfig
		}
		transport = insecure.NewCredentials()
	}
	client, err := authzed.NewClient(endpoint,
		grpc.WithTransportCredentials(transport),
		grpc.WithPerRPCCredentials(bearerCredentials{token: cfg.BearerToken, tlsRequired: !cfg.AllowInsecureLoopback}),
	)
	if err != nil {
		return nil, fmt.Errorf("create SpiceDB SDK client: %w", err)
	}
	return &Client{sdk: client, permissions: client.PermissionsServiceClient, revision: revision, catalog: catalog}, nil
}

type bearerCredentials struct {
	token       string
	tlsRequired bool
}

func (b bearerCredentials) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer " + b.token}, nil
}
func (b bearerCredentials) RequireTransportSecurity() bool { return b.tlsRequired }

// NewWithPermissionsClient supports explicit dependency injection in tests and
// composition. The injected service must already be secured and authenticated.
func NewWithPermissionsClient(modelRevision string, catalog ports.PermissionCatalog, service pb.PermissionsServiceClient) (*Client, error) {
	if catalog == nil || service == nil {
		return nil, ErrInvalidConfig
	}
	rev, err := domain.NewModelRevision(modelRevision)
	if err != nil {
		return nil, ErrInvalidConfig
	}
	return &Client{permissions: service, revision: rev, catalog: catalog}, nil
}

func (c *Client) Close() error {
	if c != nil && c.sdk != nil {
		return c.sdk.Close()
	}
	return nil
}
