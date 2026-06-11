package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	envDebug             = "M8_DEBUG"
	envHealthHTTPAddress = "M8_HEALTH_HTTP_ADDR"
)

const defaultHealthHTTPAddress = ":8080"

var ErrInvalidConfigValue = errors.New("invalid config value")

type Config struct {
	Debug      bool
	HealthHTTP HealthHTTPConfig
}

func LoadConfig() (Config, error) {
	return loadConfig(os.LookupEnv)
}

func loadConfig(lookup func(string) (string, bool)) (Config, error) {
	debug, err := boolEnv(lookup, envDebug, false)
	if err != nil {
		return Config{}, err
	}
	healthHTTPAddress := stringEnv(lookup, envHealthHTTPAddress, defaultHealthHTTPAddress)

	return Config{
		Debug: debug,
		HealthHTTP: HealthHTTPConfig{
			Address: healthHTTPAddress,
		},
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
