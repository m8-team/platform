// Command event-reconciler synchronizes configured StatefulSet slots with YDB.
package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/m8-team/platform/internal/platform/event"
	"github.com/m8-team/platform/internal/platform/kubernetes"
	"github.com/ydb-platform/ydb-go-sdk/v3"
)

const serviceAccount = "/var/run/secrets/kubernetes.io/serviceaccount"

type deployment struct {
	Namespace, StatefulSet, UID string
	// Slots includes every ordinal that could have received ownership.
	Slots int
	// Predecessors are processed before this incarnation can receive ownership.
	Predecessors []incarnation
}

type incarnation struct {
	UID   string
	Slots int
}

func main() {
	config := flag.String("config", "/etc/m8-event/replicas.json", "trusted reconciliation configuration")
	once := flag.Bool("once", false, "reconcile once and exit")
	flag.Parse()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, *config, *once); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("event reconciliation stopped", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, config string, once bool) error {
	data, err := os.ReadFile(config)
	if err != nil {
		return fmt.Errorf("read reconciliation configuration: %w", err)
	}
	var plan deployment
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&plan); err != nil {
		return fmt.Errorf("decode reconciliation configuration: %w", err)
	}
	if err := decoder.Decode(new(json.RawMessage)); err != io.EOF {
		return errors.New("unexpected trailing reconciliation configuration")
	}
	if err := plan.validate(); err != nil {
		return err
	}
	dsn := os.Getenv("YDB_DSN")
	endpoint, err := url.Parse(dsn)
	if err != nil || endpoint.Scheme != "grpcs" || endpoint.Hostname() == "" || endpoint.User != nil ||
		endpoint.RawQuery != "" || endpoint.Fragment != "" || !databasePath.MatchString(endpoint.Path) {
		return errors.New("YDB_DSN must be a grpcs endpoint with an absolute database path")
	}
	token := tokenFile(os.Getenv("YDB_TOKEN_FILE"))
	if _, err := token.Token(ctx); err != nil {
		return err
	}
	pem, err := os.ReadFile(serviceAccount + "/ca.crt")
	if err != nil {
		return fmt.Errorf("read Kubernetes CA: %w", err)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(pem) {
		return errors.New("invalid Kubernetes CA")
	}
	transport := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: kubernetesAuth{transport}, Timeout: 2 * time.Second}
	startup, cancel := context.WithTimeout(ctx, 15*time.Second)
	db, err := ydb.Open(startup, dsn, ydb.WithCredentials(token))
	cancel()
	if err != nil {
		return fmt.Errorf("connect reconciliation database: %w", err)
	}
	defer func() {
		closeCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		_ = db.Close(closeCtx)
	}()
	owner, err := event.NewTransactor(db, 8*time.Second)
	if err != nil {
		return err
	}

	owners := path.Join(endpoint.Path, "replica_owners")
	for {
		before := owner.Stats()
		err := plan.reconcile(ctx, owner, client, "https://kubernetes.default.svc", owners)
		after := owner.Stats()
		if before.OwnershipHandoffs != after.OwnershipHandoffs || before.OwnershipRetirements != after.OwnershipRetirements {
			slog.Info("event ownership changed", "handoffs", after.OwnershipHandoffs-before.OwnershipHandoffs,
				"retirements", after.OwnershipRetirements-before.OwnershipRetirements)
		}
		if once {
			return err
		}
		if err != nil && ctx.Err() == nil {
			slog.Error("event replica reconciliation failed", "error", err)
		}
		timer := time.NewTimer(2 * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

var (
	name         = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
	databasePath = regexp.MustCompile(`^(/[a-zA-Z0-9_][a-zA-Z0-9_.-]*)+$`)
)

func (d deployment) validate() error {
	if !name.MatchString(d.Namespace) || len(d.Namespace) > 63 ||
		!name.MatchString(d.StatefulSet) || len(d.StatefulSet) > 63 {
		return errors.New("invalid StatefulSet or namespace name")
	}
	seen := make(map[string]bool)
	for _, item := range append([]incarnation{{d.UID, d.Slots}}, d.Predecessors...) {
		id, err := uuid.Parse(item.UID)
		if err != nil || id == uuid.Nil || id.String() != item.UID || item.Slots <= 0 || seen[item.UID] {
			return errors.New("invalid or duplicate StatefulSet incarnation")
		}
		seen[item.UID] = true
	}
	return nil
}

func (d deployment) reconcile(ctx context.Context, owner *event.Transactor, client *http.Client, server, table string) error {
	reconcile := func(item incarnation, predecessor bool) error {
		var result error
		for ordinal := range item.Slots {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			slot := event.ReplicaSlot{
				Namespace: d.Namespace, StatefulSetUID: item.UID, PodName: fmt.Sprintf("%s-%d", d.StatefulSet, ordinal), TablePath: table,
			}
			err := slot.Reconcile(ctx, owner, func(ctx context.Context) (string, bool, error) {
				uid, retired, err := kubernetes.ObserveReplica(ctx, client, server, d.Namespace, d.StatefulSet, item.UID, slot.PodName)
				if err == nil && predecessor && !retired {
					return "", false, errors.New("predecessor StatefulSet is still active")
				}
				return uid, retired, err
			})
			if err != nil {
				result = errors.Join(result, fmt.Errorf("reconcile %s: %w", slot.PodName, err))
			}
		}
		return result
	}
	for _, previous := range d.Predecessors {
		if err := reconcile(previous, true); err != nil {
			return err
		}
	}
	return reconcile(incarnation{d.UID, d.Slots}, false)
}

type tokenFile string

func (f tokenFile) Token(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	data, err := os.ReadFile(string(f))
	if err != nil || strings.TrimSpace(string(data)) == "" {
		return "", errors.New("credential token file unavailable or empty")
	}
	return strings.TrimSpace(string(data)), nil
}

type kubernetesAuth struct{ http.RoundTripper }

func (a kubernetesAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := tokenFile(serviceAccount + "/token").Token(req.Context())
	if err != nil {
		return nil, err
	}
	copy := req.Clone(req.Context())
	copy.Header.Set("Authorization", "Bearer "+token)
	return a.RoundTripper.RoundTrip(copy)
}
