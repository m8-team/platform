package main

import (
	"github.com/m8platform/platform/internal/resourcemanager"
	"go.uber.org/fx"
)

func NewApp(cfg Config) *fx.App {
	options := []fx.Option{
		resourcemanager.Module(resourcemanager.Config{
			ServiceName: "resource-manager",
			Debug:       cfg.Debug,
		}),
	}

	if !cfg.Debug {
		options = append([]fx.Option{fx.NopLogger}, options...)
	}

	return fx.New(options...)
}
