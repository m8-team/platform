package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/m8platform/platform/iam/internal/foundation/config"
	"github.com/m8platform/platform/iam/internal/infrastructure/seeder"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	runner, err := seeder.New(cfg.YDB, "testdata/seed")
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

	log.Printf(
		"seed complete: files=%d tenants=%d users=%d memberships=%d groups=%d group_members=%d service_accounts=%d oauth_clients=%d bindings=%d",
		len(report.Files),
		report.Tenants,
		report.Users,
		report.Memberships,
		report.Groups,
		report.GroupMembers,
		report.ServiceAccounts,
		report.OAuthClients,
		report.Bindings,
	)
}
