package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/m8platform/platform/iam/internal/app"
	"github.com/m8platform/platform/iam/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	application, err := app.New(ctx, cfg)
	if err != nil {
		panic(err)
	}
	defer application.Close(context.Background())

	if err := application.Serve(ctx); err != nil {
		panic(err)
	}
}
