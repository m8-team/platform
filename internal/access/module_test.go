package access

import (
	"errors"
	"testing"

	"github.com/m8platform/platform/internal/access/domain"
	"go.uber.org/fx"
)

func TestModuleBuildsWithValidConfig(t *testing.T) {
	app := fx.New(
		Module(Config{
			ServiceName:     "m8-access",
			DefaultFailMode: domain.FailModeDeny,
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
			ServiceName: " m8-access ",
		}),
		fx.Invoke(func(config Config) {
			got = config
		}),
		fx.NopLogger,
	)
	if err := app.Err(); err != nil {
		t.Fatalf("Module() error = %v", err)
	}

	if got.ServiceName != "m8-access" {
		t.Fatalf("ServiceName = %q, want %q", got.ServiceName, "m8-access")
	}
	if got.DefaultFailMode != domain.FailModeDeny {
		t.Fatalf("DefaultFailMode = %q, want %q", got.DefaultFailMode, domain.FailModeDeny)
	}
}
