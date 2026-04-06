package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	foundationconfig "github.com/m8platform/platform/iam/internal/foundation/config"
	"github.com/m8platform/platform/iam/internal/infrastructure/bootstrap"
)

func main() {
	if err := run(); err != nil {
		log.New(os.Stderr, "", 0).Printf("m8-platform: %v", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := foundationconfig.Load()
	application, err := bootstrap.NewApplication(ctx, cfg)
	if err != nil {
		return fmt.Errorf("build application: %w", err)
	}

	serveErr := application.Serve(ctx)
	closeErr := application.Close(context.Background())

	if serveErr != nil && closeErr != nil {
		return errors.Join(
			fmt.Errorf("serve application: %w", serveErr),
			fmt.Errorf("close application: %w", closeErr),
		)
	}
	if serveErr != nil {
		return fmt.Errorf("serve application: %w", serveErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close application: %w", closeErr)
	}

	return nil
}
