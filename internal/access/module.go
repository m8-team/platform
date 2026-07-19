package access

import (
	"errors"
	"fmt"
	"strings"

	"github.com/m8-team/platform/internal/access/domain"
	"go.uber.org/fx"
)

var ErrEmptyServiceName = errors.New("access service name is empty")

type Config struct {
	ServiceName     string
	DefaultFailMode domain.FailMode
}

func Module(cfg Config) fx.Option {
	return fx.Module(
		"access",
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
	if !c.DefaultFailMode.WithDefault().IsValid() {
		return fmt.Errorf("%w: %q", domain.ErrInvalidFailMode, c.DefaultFailMode)
	}

	return nil
}

func (c Config) normalized() Config {
	c.ServiceName = strings.TrimSpace(c.ServiceName)
	c.DefaultFailMode = c.DefaultFailMode.WithDefault()
	return c
}
