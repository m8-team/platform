package main

import "testing"

func TestNewAppBuilds(t *testing.T) {
	app := NewApp(Config{})
	if err := app.Err(); err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}
}
