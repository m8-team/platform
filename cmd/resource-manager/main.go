package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"

	"github.com/m8platform/platform/internal/bootstrap"
	"github.com/m8platform/platform/internal/frameworks/config"
)

func main() {
	ctx := context.Background()
	app, err := bootstrap.NewApp(ctx, config.Load())
	if err != nil {
		log.Fatalf("bootstrap resource-manager: %v", err)
	}

	listener, err := net.Listen("tcp", app.Config.GRPCAddress)
	if err != nil {
		log.Fatalf("listen grpc: %v", err)
	}

	go func() {
		if err := app.HTTPServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("serve http: %v", err)
		}
	}()

	if err := app.GRPCServer.Serve(listener); err != nil {
		log.Fatalf("serve grpc: %v", err)
	}
}
