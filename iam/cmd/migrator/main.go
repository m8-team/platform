package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	entries, err := os.ReadDir("migrations")
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fmt.Println(filepath.Join("migrations", entry.Name()))
	}
}
