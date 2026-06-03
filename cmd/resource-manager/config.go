package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const envDebug = "M8_DEBUG"

var ErrInvalidConfigValue = errors.New("invalid config value")

type Config struct {
	Debug bool
}

func LoadConfig() (Config, error) {
	return loadConfig(os.LookupEnv)
}

func loadConfig(lookup func(string) (string, bool)) (Config, error) {
	debug, err := boolEnv(lookup, envDebug, false)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Debug: debug,
	}, nil
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
