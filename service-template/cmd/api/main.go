package main

import (
	"connectrpc.com/connect"
	"context"
	"crypto/rand"
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/m8-team/platform-sdk/app"
	sdkconnect "github.com/m8-team/platform-sdk/connect"
	"github.com/m8-team/platform-sdk/health"
	"github.com/m8-team/platform-sdk/kafka"
	"github.com/m8-team/platform-sdk/lifecycle"
	"github.com/m8-team/platform-sdk/logging"
	"github.com/m8-team/platform-sdk/outbox"
	"github.com/m8-team/platform-sdk/telemetry"
	"github.com/m8-team/product-service/internal/gen/product/v1/productv1connect"
	"github.com/m8-team/product-service/internal/platform"
	connectadapter "github.com/m8-team/product-service/internal/product/adapter/connect"
	"github.com/m8-team/product-service/internal/product/adapter/postgres"
	productapp "github.com/m8-team/product-service/internal/product/app"
	"github.com/m8-team/product-service/internal/product/domain"
	"github.com/twmb/franz-go/pkg/kgo"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var version = "dev" // immutable build metadata supplied by -ldflags, never a dependency container.

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "product-service:", err)
		os.Exit(1)
	}
}
func run() (result error) {
	handedOff := false
	cfg, err := platform.Load()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	info := logging.Info{Name: "product-service", Version: version, Namespace: "company", Environment: "production", InstanceID: rand.Text()}
	if cfg.Local {
		info.Environment = "local"
	}
	logger := logging.New(os.Stdout, info, logging.Config{Text: cfg.Local})
	providers, err := telemetry.New(ctx, info, telemetry.Config{Endpoint: cfg.OTLPEndpoint, Insecure: cfg.Local, SampleRatio: 0.1})
	if err != nil {
		return err
	}
	// Constructors allocate resources before app.Run takes ownership. These defers
	// also cover composition failures; Close/Shutdown are safe to call again.
	defer func() {
		if handedOff {
			return
		}
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		result = errors.Join(result, providers.Shutdown(cleanup))
	}()
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return errors.New("open PostgreSQL configuration failed")
	}
	defer func() {
		if !handedOff {
			result = errors.Join(result, db.Close())
		}
	}()
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	kafkaOptions := []kgo.Opt{kgo.SeedBrokers(cfg.Brokers...), kgo.ClientID(info.Name), kgo.RecordRetries(2), kgo.RecordDeliveryTimeout(8 * time.Second), kgo.ProducerBatchCompression(kgo.SnappyCompression())}
	if !cfg.Local {
		kafkaOptions = append(kafkaOptions, kgo.DialTLSConfig(&tls.Config{MinVersion: tls.VersionTLS12}))
	}
	producer, err := kafka.NewProducer(providers, kafkaOptions...)
	if err != nil {
		return err
	}
	defer func() {
		if !handedOff {
			producer.Client.Close()
		}
	}()
	identity := platform.Identity{Local: cfg.Local, Token: cfg.DevToken, Endpoint: cfg.IntrospectionURL, ClientID: cfg.ClientID, ClientSecret: cfg.ClientSecret, Client: &http.Client{Timeout: 3 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
	interceptor, err := sdkconnect.New(sdkconnect.Config{Logger: logger, Telemetry: providers, Authentication: identity, Authorization: platform.TransportPolicy{}})
	if err != nil {
		return err
	}
	repository := postgres.New(db, providers.Tracer, providers.Propagator)
	products, err := productapp.New(repository, platform.Policy{}, func() domain.ID { return domain.ID("prd_" + rand.Text()) })
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	path, handler := productv1connect.NewProductServiceHandler(&connectadapter.Handler{Service: products}, connect.WithInterceptors(interceptor), connect.WithReadMaxBytes(1<<20))
	mux.Handle(path, handler) // GetProduct also supports Connect's HTTP GET protocol.
	public, err := app.HTTP("public-http", app.HTTPConfig{Address: cfg.HTTPAddress, CertFile: cfg.CertFile, KeyFile: cfg.KeyFile, InsecureLocal: cfg.Local}, mux)
	if err != nil {
		return err
	}
	checks := health.New()
	if err = checks.Register("postgres", db.PingContext); err != nil {
		return err
	}
	adminMux := http.NewServeMux()
	adminMux.Handle("/health/", checks.Handler())
	adminMux.Handle("/metrics", providers.Handler)
	// Private management port: no Kubernetes Service exposure; NetworkPolicy must
	// restrict scrape/probe access. Public API always requires TLS outside local mode.
	admin, err := app.HTTP("management-http", app.HTTPConfig{Address: cfg.AdminAddress, InsecureLocal: true}, adminMux)
	if err != nil {
		return err
	}
	worker, err := outbox.New(db, producer, logger, outbox.Config{PollInterval: time.Second, PublishTimeout: 10 * time.Second, Lease: 30 * time.Second, MaxAttempts: 20, Meter: providers.Meter})
	if err != nil {
		return err
	}
	handedOff = true
	return app.Run(ctx, app.Config{Logger: logger, Health: checks, StartupTimeout: 15 * time.Second, ShutdownTimeout: cfg.ShutdownTimeout},
		lifecycle.Component{Name: "postgres", Stop: func(context.Context) error { return db.Close() }},
		lifecycle.Component{Name: "kafka", Stop: func(context.Context) error { producer.Client.Close(); return nil }},
		lifecycle.Component{Name: "telemetry", Stop: providers.Shutdown},
		lifecycle.Component{Name: "dependency-probes", Start: func(ctx context.Context) error {
			if err := db.PingContext(ctx); err != nil {
				return fmt.Errorf("postgres startup: %w", err)
			}
			if err := producer.Client.Ping(ctx); err != nil {
				return fmt.Errorf("kafka startup: %w", err)
			}
			return nil
		}},
		admin, public, lifecycle.Component{Name: "outbox", Run: worker.Run},
	)
}
