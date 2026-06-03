package main

import (
	"errors"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	cfg, err := loadConfig(emptyLookup)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	if cfg.Debug {
		t.Fatal("Debug = true, want false")
	}
	if cfg.HealthHTTP.Address != defaultHealthHTTPAddress {
		t.Fatalf("HealthHTTP.Address = %q, want %q", cfg.HealthHTTP.Address, defaultHealthHTTPAddress)
	}
}

func TestLoadConfigDebug(t *testing.T) {
	cfg, err := loadConfig(mapLookup(map[string]string{
		envDebug: "true",
	}))
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	if !cfg.Debug {
		t.Fatal("Debug = false, want true")
	}
}

func TestLoadConfigInvalidDebug(t *testing.T) {
	_, err := loadConfig(mapLookup(map[string]string{
		envDebug: "definitely",
	}))
	if !errors.Is(err, ErrInvalidConfigValue) {
		t.Fatalf("loadConfig() error = %v, want %v", err, ErrInvalidConfigValue)
	}
}

func TestLoadConfigHealthHTTPAddress(t *testing.T) {
	cfg, err := loadConfig(mapLookup(map[string]string{
		envHealthHTTPAddress: " 127.0.0.1:9090 ",
	}))
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	if cfg.HealthHTTP.Address != "127.0.0.1:9090" {
		t.Fatalf("HealthHTTP.Address = %q, want 127.0.0.1:9090", cfg.HealthHTTP.Address)
	}
}

func emptyLookup(string) (string, bool) {
	return "", false
}

func mapLookup(values map[string]string) func(string) (string, bool) {
	return func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	}
}
