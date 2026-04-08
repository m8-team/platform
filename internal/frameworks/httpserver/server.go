package httpserver

import (
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

func New(address string, mux *runtime.ServeMux) *http.Server {
	return &http.Server{
		Addr:    address,
		Handler: mux,
	}
}
