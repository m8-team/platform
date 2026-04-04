package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/m8platform/platform/iam/internal/config"
	"github.com/m8platform/platform/iam/internal/spicedb"
	ydbstore "github.com/m8platform/platform/iam/internal/storage/ydb"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	client := spicedb.NewClient(cfg.SpiceDB)
	defer func() {
		_ = client.Close()
	}()

	if err := client.ApplySchemaFile(ctx, cfg.SpiceDB.SchemaPath); err != nil {
		log.Fatal(err)
	}
	log.Printf("spicedb schema applied: %s", cfg.SpiceDB.SchemaPath)

	if cfg.YDB.DSN == "" {
		return
	}

	store, err := ydbstore.Open(ctx, cfg.YDB)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = store.Close(context.Background())
	}()

	report, err := client.SyncSnapshot(ctx, store)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf(
		"spicedb snapshot synced: group_members=%d resources=%d bindings=%d relationships=%d",
		report.GroupMembers,
		report.Resources,
		report.Bindings,
		report.Relationships,
	)
}
