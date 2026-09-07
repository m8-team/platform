package health

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type Check func(context.Context) error
type Registry struct {
	started, ready, broken atomic.Bool
	mu                     sync.RWMutex
	checks                 map[string]Check
	Timeout                time.Duration
}

func New() *Registry { return &Registry{checks: make(map[string]Check), Timeout: time.Second} }

// Register is intended for composition before serving; duplicate names fail.
func (r *Registry) Register(name string, check Check) error {
	if name == "" || check == nil {
		return errors.New("check name and function required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.checks[name]; ok {
		return errors.New("duplicate health check")
	}
	r.checks[name] = check
	return nil
}
func (r *Registry) Started() { r.started.Store(true); r.ready.Store(true) }
func (r *Registry) Drain()   { r.ready.Store(false) }
func (r *Registry) Broken()  { r.broken.Store(true); r.Drain() }
func (r *Registry) Ready(ctx context.Context) bool {
	if !r.started.Load() || !r.ready.Load() {
		return false
	}
	r.mu.RLock()
	checks := make([]Check, 0, len(r.checks))
	for _, c := range r.checks {
		checks = append(checks, c)
	}
	r.mu.RUnlock()
	ctx, cancel := context.WithTimeout(ctx, r.Timeout)
	defer cancel()
	for _, check := range checks {
		if err := check(ctx); err != nil {
			return false
		}
	}
	return ctx.Err() == nil && r.ready.Load()
}
func (r *Registry) Handler() http.Handler {
	mux := http.NewServeMux()
	reply := func(w http.ResponseWriter, ok bool) {
		w.Header().Set("Cache-Control", "no-store")
		if !ok {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, _ *http.Request) { reply(w, !r.broken.Load()) })
	mux.HandleFunc("GET /health/startup", func(w http.ResponseWriter, _ *http.Request) { reply(w, r.started.Load()) })
	mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, q *http.Request) { reply(w, r.Ready(q.Context())) })
	return mux
}
