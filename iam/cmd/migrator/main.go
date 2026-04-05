package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/m8platform/platform/iam/internal/foundation/config"
	"github.com/m8platform/platform/iam/internal/migrator"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	runner, err := migrator.New(cfg.YDB, "migrations")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = runner.Close(context.Background())
	}()

	report, err := runner.Run(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, item := range report.Items {
		log.Printf("%s: %s", item.Name, item.Status)
	}
	log.Printf("migrations complete: applied=%d backfilled=%d skipped=%d", report.Applied, report.Backfilled, report.Skipped)
}
