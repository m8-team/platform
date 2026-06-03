package http

import (
	"encoding/json"
	nethttp "net/http"
	"time"

	"github.com/m8platform/platform/internal/platform/health"
)

type Handler struct {
	registry health.Registry
}

func NewHandler(registry health.Registry) *Handler {
	return &Handler{registry: registry}
}

func (h *Handler) RegisterRoutes(mux *nethttp.ServeMux) {
	mux.HandleFunc("/livez", h.handle(health.CheckKindLiveness))
	mux.HandleFunc("/readyz", h.handle(health.CheckKindReadiness))
	mux.HandleFunc("/startupz", h.handle(health.CheckKindStartup))
	mux.HandleFunc("/healthz", h.handle(health.CheckKindDeep))
}

func (h *Handler) handle(kind health.CheckKind) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != nethttp.MethodGet {
			w.Header().Set("Allow", nethttp.MethodGet)
			w.WriteHeader(nethttp.StatusMethodNotAllowed)
			return
		}

		snapshot := h.snapshot(r, kind)
		w.WriteHeader(statusCode(snapshot.Status))

		_ = json.NewEncoder(w).Encode(snapshot)
	}
}

func (h *Handler) snapshot(r *nethttp.Request, kind health.CheckKind) health.Snapshot {
	if h.registry != nil {
		return h.registry.Snapshot(r.Context(), kind)
	}

	return health.Snapshot{
		Status:    health.StatusUnhealthy,
		Kind:      kind,
		CheckedAt: time.Now().UTC(),
		Results: []health.Result{
			{
				Name:        "health-registry",
				Status:      health.StatusUnhealthy,
				Message:     "health registry is not configured",
				Error:       health.ErrRegistryRequired.Error(),
				CheckedAt:   time.Now().UTC(),
				Target:      health.Target{Kind: health.TargetApplication, Name: "platform-health"},
				Criticality: health.CriticalityRequired,
			},
		},
	}
}

func statusCode(status health.Status) int {
	switch status {
	case health.StatusHealthy, health.StatusDegraded:
		return nethttp.StatusOK
	case health.StatusUnhealthy, health.StatusUnknown:
		return nethttp.StatusServiceUnavailable
	default:
		return nethttp.StatusServiceUnavailable
	}
}
