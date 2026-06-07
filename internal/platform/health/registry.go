package health

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	defaultTimeout  = 2 * time.Second
	defaultInterval = 10 * time.Second
)

var (
	ErrRegistryRequired     = errors.New("health registry is required")
	ErrCheckNameRequired    = errors.New("health check name is required")
	ErrCheckCheckerRequired = errors.New("health check checker is required")
	ErrCheckKindsRequired   = errors.New("health check kind is required")
	ErrDuplicateCheck       = errors.New("health check already registered")
	ErrInvalidCriticality   = errors.New("invalid health check criticality")
)

type Registry interface {
	Register(registration Config) error
	Snapshot(ctx context.Context, kind Kind) Snapshot
}

type registry struct {
	mu     sync.RWMutex
	checks map[string]Config
}

func NewRegistry() Registry {
	return &registry{
		checks: make(map[string]Config),
	}
}

func (r *registry) Register(registration Config) error {
	if r == nil {
		return ErrRegistryRequired
	}

	registered, err := normalizeCheck(registration)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.checks[registered.Spec.Name]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateCheck, registered.Spec.Name)
	}

	r.checks[registered.Spec.Name] = registered
	return nil
}

func (r *registry) Snapshot(ctx context.Context, kind Kind) Snapshot {
	if r == nil {
		return Snapshot{
			Status:    StatusUnhealthy,
			Kind:      kind,
			CheckedAt: time.Now().UTC(),
			Results: []Result{
				{
					Name:        "health-registry",
					Status:      StatusUnhealthy,
					Message:     "health registry is not configured",
					Error:       ErrRegistryRequired.Error(),
					CheckedAt:   time.Now().UTC(),
					Target:      Target{Kind: TargetKindApplication, Name: "platform-health"},
					Criticality: CriticalityRequired,
				},
			},
		}
	}

	checks := r.checksForKind(kind)
	results := runChecks(ctx, checks)

	return Snapshot{
		Status:    Aggregate(results),
		Kind:      kind,
		CheckedAt: time.Now().UTC(),
		Results:   results,
	}
}

func (r *registry) checksForKind(kind Kind) []Config {
	r.mu.RLock()
	defer r.mu.RUnlock()

	checks := make([]Config, 0, len(r.checks))
	for _, registration := range r.checks {
		if checkHasKind(registration.Spec, kind) {
			checks = append(checks, copyCheck(registration))
		}
	}

	sort.Slice(checks, func(i, j int) bool {
		return checks[i].Spec.Name < checks[j].Spec.Name
	})

	return checks
}

func Register(registry Registry, registrations ...Config) error {
	if registry == nil {
		return ErrRegistryRequired
	}

	for _, registration := range registrations {
		if err := registry.Register(registration); err != nil {
			return err
		}
	}

	return nil
}

func normalizeCheck(registration Config) (Config, error) {
	spec := registration.Spec
	spec.Name = strings.TrimSpace(spec.Name)
	if spec.Name == "" {
		return Config{}, ErrCheckNameRequired
	}
	if isNilChecker(registration.Checker) {
		return Config{}, fmt.Errorf("%w: %s", ErrCheckCheckerRequired, spec.Name)
	}
	if len(spec.Kinds) == 0 {
		return Config{}, fmt.Errorf("%w: %s", ErrCheckKindsRequired, spec.Name)
	}
	if spec.Criticality == "" {
		spec.Criticality = CriticalityRequired
	}
	if spec.Criticality != CriticalityRequired && spec.Criticality != CriticalityOptional {
		return Config{}, fmt.Errorf("%w: %s", ErrInvalidCriticality, spec.Criticality)
	}
	if spec.Timeout <= 0 {
		spec.Timeout = defaultTimeout
	}
	if spec.Interval <= 0 {
		spec.Interval = defaultInterval
	}

	spec.Kinds = append([]Kind(nil), spec.Kinds...)
	registration.Spec = spec
	return registration, nil
}

func isNilChecker(checker Checker) bool {
	if checker == nil {
		return true
	}

	value := reflect.ValueOf(checker)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func copyCheck(registration Config) Config {
	registration.Spec.Kinds = append([]Kind(nil), registration.Spec.Kinds...)
	return registration
}

func checkHasKind(spec Spec, kind Kind) bool {
	for _, candidate := range spec.Kinds {
		if candidate == kind {
			return true
		}
	}

	return false
}
