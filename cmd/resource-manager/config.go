package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	grpcserver "github.com/m8-team/platform/internal/platform/server/grpc"
)

const (
	envDebug                = "M8_DEBUG"
	envHTTPAddress          = "M8_HTTP_ADDR"
	envHealthHTTPAddress    = "M8_HEALTH_HTTP_ADDR"
	envGRPCAddress          = "M8_GRPC_ADDR"
	envAllowUnauthenticated = "M8_RM_ALLOW_UNAUTHENTICATED"
	envSoftDeleteRetention  = "M8_RM_SOFT_DELETE_RETENTION"
	envPageTokenKey         = "M8_RM_PAGE_TOKEN_KEY"
)

const (
	defaultHTTPAddress         = ":8080"
	defaultHealthHTTPAddress   = ":8081"
	defaultGRPCAddress         = ":9090"
	defaultSoftDeleteRetention = 30 * 24 * time.Hour
	minimumPageTokenKeyLength  = 32
)

var ErrInvalidConfigValue = errors.New("invalid config value")

type Config struct {
	Debug                bool
	HTTP                 HTTPConfig
	HealthHTTP           HealthHTTPConfig
	GRPC                 grpcserver.Config
	AllowUnauthenticated bool
	SoftDeleteRetention  time.Duration
	PageTokenKey         []byte
}

func LoadConfig() (Config, error) {
	return loadConfig(os.LookupEnv)
}

func loadConfig(lookup func(string) (string, bool)) (Config, error) {
	debug, err := boolEnv(lookup, envDebug, false)
	if err != nil {
		return Config{}, err
	}
	httpAddress := stringEnv(lookup, envHTTPAddress, defaultHTTPAddress)
	healthHTTPAddress := stringEnv(lookup, envHealthHTTPAddress, defaultHealthHTTPAddress)
	grpcAddress := stringEnv(lookup, envGRPCAddress, defaultGRPCAddress)
	allowUnauthenticated, err := boolEnv(lookup, envAllowUnauthenticated, false)
	if err != nil {
		return Config{}, err
	}
	softDeleteRetention, err := durationEnv(lookup, envSoftDeleteRetention, defaultSoftDeleteRetention)
	if err != nil {
		return Config{}, err
	}
	pageTokenKey := []byte(stringEnv(lookup, envPageTokenKey, ""))
	if len(pageTokenKey) > 0 && len(pageTokenKey) < minimumPageTokenKeyLength {
		return Config{}, fmt.Errorf(
			"%w: %s must contain at least %d bytes",
			ErrInvalidConfigValue,
			envPageTokenKey,
			minimumPageTokenKeyLength,
		)
	}

	return Config{
		Debug:                debug,
		HTTP:                 HTTPConfig{Address: httpAddress},
		AllowUnauthenticated: allowUnauthenticated,
		SoftDeleteRetention:  softDeleteRetention,
		PageTokenKey:         pageTokenKey,
		HealthHTTP: HealthHTTPConfig{
			Address: healthHTTPAddress,
		},
		GRPC: grpcserver.Config{Address: grpcAddress},
	}, nil
}

func stringEnv(lookup func(string) (string, bool), name string, defaultValue string) string {
	value, ok := lookup(name)
	if !ok || strings.TrimSpace(value) == "" {
		return defaultValue
	}

	return strings.TrimSpace(value)
}

func boolEnv(lookup func(string) (string, bool), name string, defaultValue bool) (bool, error) {
	value, ok := lookup(name)
	if !ok || strings.TrimSpace(value) == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false, fmt.Errorf("%w: %s=%q must be a boolean", ErrInvalidConfigValue, name, value)
	}

	return parsed, nil
}

func durationEnv(
	lookup func(string) (string, bool),
	name string,
	defaultValue time.Duration,
) (time.Duration, error) {
	value, ok := lookup(name)
	if !ok || strings.TrimSpace(value) == "" {
		return defaultValue, nil
	}

	parsed, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%w: %s=%q must be a positive duration", ErrInvalidConfigValue, name, value)
	}
	return parsed, nil
}
