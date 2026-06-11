package health

import "errors"

var ErrModuleRequired = errors.New("health module is required")

type Module interface {
	HealthChecks() []Config
}

func RegisterModuleChecks(registry Registry, module Module) error {
	if module == nil {
		return ErrModuleRequired
	}

	return Register(registry, module.HealthChecks()...)
}
