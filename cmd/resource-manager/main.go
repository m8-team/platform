package main

import (
	"fmt"
	"os"

	_ "go.uber.org/automaxprocs"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	NewApp(cfg).Run()
}
