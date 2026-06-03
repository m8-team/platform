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
	Register(check Check) error
	Snapshot(ctx context.Context, kind CheckKind) Snapshot
}

type registry struct {
	mu     sync.RWMutex
	checks map[string]Check
}

func NewRegistry() Registry {
	return &registry{
		checks: make(map[string]Check),
	}
}

func (r *registry) Register(check Check) error {
	if r == nil {
		return ErrRegistryRequired
	}

	registered, err := normalizeCheck(check)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.checks[registered.Name]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateCheck, registered.Name)
	}

	r.checks[registered.Name] = registered
	return nil
}

func (r *registry) Snapshot(ctx context.Context, kind CheckKind) Snapshot {
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
					Target:      Target{Kind: TargetApplication, Name: "platform-health"},
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

func (r *registry) checksForKind(kind CheckKind) []Check {
	r.mu.RLock()
	defer r.mu.RUnlock()

	checks := make([]Check, 0, len(r.checks))
	for _, check := range r.checks {
		if checkHasKind(check, kind) {
			checks = append(checks, copyCheck(check))
		}
	}

	sort.Slice(checks, func(i, j int) bool {
		return checks[i].Name < checks[j].Name
	})

	return checks
}

func RegisterChecks(registry Registry, checks ...Check) error {
	if registry == nil {
		return ErrRegistryRequired
	}

	for _, check := range checks {
		if err := registry.Register(check); err != nil {
			return err
		}
	}

	return nil
}

func normalizeCheck(check Check) (Check, error) {
	check.Name = strings.TrimSpace(check.Name)
	if check.Name == "" {
		return Check{}, ErrCheckNameRequired
	}
	if isNilChecker(check.Checker) {
		return Check{}, fmt.Errorf("%w: %s", ErrCheckCheckerRequired, check.Name)
	}
	if len(check.Kinds) == 0 {
		return Check{}, fmt.Errorf("%w: %s", ErrCheckKindsRequired, check.Name)
	}
	if check.Criticality == "" {
		check.Criticality = CriticalityRequired
	}
	if check.Criticality != CriticalityRequired && check.Criticality != CriticalityOptional {
		return Check{}, fmt.Errorf("%w: %s", ErrInvalidCriticality, check.Criticality)
	}
	if check.Timeout <= 0 {
		check.Timeout = defaultTimeout
	}
	if check.Interval <= 0 {
		check.Interval = defaultInterval
	}

	check.Kinds = append([]CheckKind(nil), check.Kinds...)
	return check, nil
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

func copyCheck(check Check) Check {
	check.Kinds = append([]CheckKind(nil), check.Kinds...)
	return check
}

func checkHasKind(check Check, kind CheckKind) bool {
	for _, candidate := range check.Kinds {
		if candidate == kind {
			return true
		}
	}

	return false
}
