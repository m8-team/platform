package resourcemanager

import (
	"go.uber.org/fx"
)

type Config struct {
	ServiceName string
	Debug       bool
}

func Module(cfg Config) fx.Option {
	return fx.Module(
		cfg.ServiceName,
	)
}
