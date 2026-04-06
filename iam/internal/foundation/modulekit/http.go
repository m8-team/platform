package modulekit

import "net/http"

type HTTPRoute struct {
	Method  string
	Pattern string
	Handler http.Handler
}

type HTTPRegistrar interface {
	RegisterHTTPRoute(route HTTPRoute)
}

type HTTPRegistry struct {
	routes []HTTPRoute
}

func NewHTTPRegistry() *HTTPRegistry {
	return &HTTPRegistry{routes: make([]HTTPRoute, 0)}
}

func (r *HTTPRegistry) RegisterHTTPRoute(route HTTPRoute) {
	if r == nil {
		return
	}
	r.routes = append(r.routes, route)
}

func (r *HTTPRegistry) Routes() []HTTPRoute {
	if r == nil {
		return nil
	}
	routes := make([]HTTPRoute, len(r.routes))
	copy(routes, r.routes)
	return routes
}
