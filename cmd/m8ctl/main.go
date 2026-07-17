package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/m8platform/platform/internal/installer/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app := cli.New(os.Stdout, os.Stderr)
	os.Exit(app.Run(ctx, os.Args[1:]))
}
