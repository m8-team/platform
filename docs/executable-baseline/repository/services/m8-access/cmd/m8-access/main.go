package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})
	address := os.Getenv("HTTP_ADDRESS")
	if address == "" {
		address = ":8080"
	}
	fmt.Printf("m8-access listening on %s\n", address)
	if err := http.ListenAndServe(address, mux); err != nil {
		panic(err)
	}
}
