package resourcemanager

import (
	"errors"
	"strings"

	"go.uber.org/fx"
)

var ErrEmptyServiceName = errors.New("resource manager service name is empty")

type Config struct {
	ServiceName string
	Debug       bool
}

func Module(cfg Config) fx.Option {
	return fx.Module(
		"resourcemanager",
		fx.Supply(cfg.normalized()),
		fx.Invoke(configureModule),
	)
}

func configureModule(cfg Config) error {
	return cfg.Validate()
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.ServiceName) == "" {
		return ErrEmptyServiceName
	}

	return nil
}

func (c Config) normalized() Config {
	c.ServiceName = strings.TrimSpace(c.ServiceName)
	return c
}
