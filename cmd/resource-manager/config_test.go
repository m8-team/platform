package main

import (
	"errors"
	"testing"
	"time"
)

func TestLoadConfigDefaults(t *testing.T) {
	cfg, err := loadConfig(emptyLookup)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	if cfg.Debug {
		t.Fatal("Debug = true, want false")
	}
	if cfg.HTTP.Address != defaultHTTPAddress {
		t.Fatalf("HTTP.Address = %q, want %q", cfg.HTTP.Address, defaultHTTPAddress)
	}
	if cfg.HealthHTTP.Address != defaultHealthHTTPAddress {
		t.Fatalf("HealthHTTP.Address = %q, want %q", cfg.HealthHTTP.Address, defaultHealthHTTPAddress)
	}
	if cfg.GRPC.Address != defaultGRPCAddress {
		t.Fatalf("GRPC.Address = %q, want %q", cfg.GRPC.Address, defaultGRPCAddress)
	}
	if cfg.AllowUnauthenticated {
		t.Fatal("AllowUnauthenticated = true, want false")
	}
	if cfg.SoftDeleteRetention != defaultSoftDeleteRetention {
		t.Fatalf("SoftDeleteRetention = %s, want %s", cfg.SoftDeleteRetention, defaultSoftDeleteRetention)
	}
	if len(cfg.PageTokenKey) != 0 {
		t.Fatalf("PageTokenKey len = %d, want 0", len(cfg.PageTokenKey))
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
		envHTTPAddress:       " 127.0.0.1:8088 ",
		envHealthHTTPAddress: " 127.0.0.1:9090 ",
	}))
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	if cfg.HealthHTTP.Address != "127.0.0.1:9090" {
		t.Fatalf("HealthHTTP.Address = %q, want 127.0.0.1:9090", cfg.HealthHTTP.Address)
	}
	if cfg.HTTP.Address != "127.0.0.1:8088" {
		t.Fatalf("HTTP.Address = %q, want 127.0.0.1:8088", cfg.HTTP.Address)
	}
}

func TestLoadConfigOrganizationService(t *testing.T) {
	key := "01234567890123456789012345678901"
	cfg, err := loadConfig(mapLookup(map[string]string{
		envGRPCAddress:          " 127.0.0.1:9191 ",
		envAllowUnauthenticated: "true",
		envSoftDeleteRetention:  "48h",
		envPageTokenKey:         key,
	}))
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	if cfg.GRPC.Address != "127.0.0.1:9191" {
		t.Fatalf("GRPC.Address = %q, want 127.0.0.1:9191", cfg.GRPC.Address)
	}
	if !cfg.AllowUnauthenticated {
		t.Fatal("AllowUnauthenticated = false, want true")
	}
	if cfg.SoftDeleteRetention != 48*time.Hour {
		t.Fatalf("SoftDeleteRetention = %s, want 48h", cfg.SoftDeleteRetention)
	}
	if string(cfg.PageTokenKey) != key {
		t.Fatal("PageTokenKey does not match")
	}
}

func TestLoadConfigRejectsInvalidOrganizationServiceConfig(t *testing.T) {
	tests := []map[string]string{
		{envAllowUnauthenticated: "sometimes"},
		{envSoftDeleteRetention: "never"},
		{envSoftDeleteRetention: "0s"},
		{envPageTokenKey: "too-short"},
	}
	for _, values := range tests {
		if _, err := loadConfig(mapLookup(values)); !errors.Is(err, ErrInvalidConfigValue) {
			t.Fatalf("loadConfig(%v) error = %v, want %v", values, err, ErrInvalidConfigValue)
		}
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
