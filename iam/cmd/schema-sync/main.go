package main

import (
	"fmt"
	"os"

	"github.com/m8platform/platform/iam/internal/config"
)

func main() {
	cfg := config.Load()
	payload, err := os.ReadFile(cfg.SpiceDB.SchemaPath)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(payload))
}
