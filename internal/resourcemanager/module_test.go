package resourcemanager

import (
	"errors"
	"testing"

	"github.com/m8-team/platform/internal/resourcemanager/app/usecase"
	"go.uber.org/fx"
)

func TestModuleBuildsWithValidConfig(t *testing.T) {
	app := fx.New(
		Module(Config{
			ServiceName: "resource-manager",
			Debug:       true,
		}),
		fx.Invoke(func(*usecase.OrganizationService) {}),
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

func TestModuleRejectsInvalidOrganizationConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr error
	}{
		{
			name: "negative retention",
			config: Config{
				ServiceName:         "resource-manager",
				SoftDeleteRetention: -1,
			},
			wantErr: usecase.ErrInvalidSoftDeleteRetention,
		},
		{
			name: "short page token key",
			config: Config{
				ServiceName:  "resource-manager",
				PageTokenKey: []byte("short"),
			},
			wantErr: usecase.ErrInvalidPageTokenKey,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := fx.New(Module(test.config), fx.NopLogger)
			if !errors.Is(app.Err(), test.wantErr) {
				t.Fatalf("Module() error = %v, want %v", app.Err(), test.wantErr)
			}
		})
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
