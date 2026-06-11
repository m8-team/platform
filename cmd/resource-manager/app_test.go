package main

import "testing"

func TestNewAppBuilds(t *testing.T) {
	app := NewApp(Config{
		HealthHTTP: HealthHTTPConfig{Address: "127.0.0.1:0"},
	})
	if err := app.Err(); err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}
}
