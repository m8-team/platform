package platform

import (
	"errors"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DatabaseURL                                        string
	Brokers                                            []string
	HTTPAddress, AdminAddress, CertFile, KeyFile       string
	Local                                              bool
	DevToken, IntrospectionURL, ClientID, ClientSecret string
	OTLPEndpoint                                       string
	ShutdownTimeout                                    time.Duration
}

// Load is the only environment boundary. The returned value is never mutated.
func Load() (Config, error) {
	c := Config{DatabaseURL: os.Getenv("DATABASE_URL"), Brokers: strings.Split(os.Getenv("KAFKA_BROKERS"), ","), HTTPAddress: os.Getenv("HTTP_ADDRESS"), AdminAddress: os.Getenv("ADMIN_ADDRESS"), CertFile: os.Getenv("TLS_CERT_FILE"), KeyFile: os.Getenv("TLS_KEY_FILE"), DevToken: os.Getenv("DEV_AUTH_TOKEN"), IntrospectionURL: os.Getenv("AUTH_INTROSPECTION_URL"), ClientID: os.Getenv("AUTH_CLIENT_ID"), ClientSecret: os.Getenv("AUTH_CLIENT_SECRET"), OTLPEndpoint: os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"), ShutdownTimeout: 30 * time.Second}
	if value := os.Getenv("LOCAL_INSECURE"); value != "" {
		var err error
		c.Local, err = strconv.ParseBool(value)
		if err != nil {
			return c, errors.New("invalid LOCAL_INSECURE")
		}
	}
	if value := os.Getenv("SHUTDOWN_TIMEOUT"); value != "" {
		var err error
		c.ShutdownTimeout, err = time.ParseDuration(value)
		if err != nil || c.ShutdownTimeout <= 0 || c.ShutdownTimeout > 35*time.Second {
			return c, errors.New("SHUTDOWN_TIMEOUT must be positive and at most 35s for the supplied deployment")
		}
	}
	if c.HTTPAddress == "" {
		c.HTTPAddress = ":8080"
	}
	if c.AdminAddress == "" {
		c.AdminAddress = ":8081"
	}
	if c.DatabaseURL == "" || len(c.Brokers) == 0 || c.Brokers[0] == "" {
		return c, errors.New("DATABASE_URL and KAFKA_BROKERS required")
	}
	if c.Local {
		if len(c.DevToken) < 32 {
			return c, errors.New("local mode requires DEV_AUTH_TOKEN with at least 32 bytes")
		}
	} else {
		if c.CertFile == "" || c.KeyFile == "" {
			return c, errors.New("TLS certificate and key required")
		}
		u, err := url.Parse(c.IntrospectionURL)
		if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || c.ClientID == "" || c.ClientSecret == "" {
			return c, errors.New("HTTPS introspection endpoint and client credentials required")
		}
		db, err := url.Parse(c.DatabaseURL)
		if err != nil || db.Query().Get("sslmode") != "verify-full" {
			return c, errors.New("production database requires sslmode=verify-full")
		}
	}
	return c, nil
}
