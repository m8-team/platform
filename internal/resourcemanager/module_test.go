package resourcemanager

import (
	"errors"
	"testing"

	"go.uber.org/fx"
)

func TestModuleBuildsWithValidConfig(t *testing.T) {
	app := fx.New(
		Module(Config{
			ServiceName: "resource-manager",
			Debug:       true,
		}),
		fx.NopLogger,
	)
	if err := app.Err(); err != nil {
		t.Fatalf("Module() error = %v", err)
	}
}

func TestModuleRejectsEmptyServiceName(t *testing.T) {
	app := fx.New(
		Module(Config{
			ServiceName: " ",
		}),
		fx.NopLogger,
	)
	if !errors.Is(app.Err(), ErrEmptyServiceName) {
		t.Fatalf("Module() error = %v, want %v", app.Err(), ErrEmptyServiceName)
	}
}

func TestModuleSuppliesNormalizedConfig(t *testing.T) {
	var got Config

	app := fx.New(
		Module(Config{
			ServiceName: " resource-manager ",
			Debug:       true,
		}),
		fx.Invoke(func(config Config) {
			got = config
		}),
		fx.NopLogger,
	)
	if err := app.Err(); err != nil {
		t.Fatalf("Module() error = %v", err)
	}

	if got.ServiceName != "resource-manager" {
		t.Fatalf("ServiceName = %q, want %q", got.ServiceName, "resource-manager")
	}
	if !got.Debug {
		t.Fatal("Debug = false, want true")
	}
}
