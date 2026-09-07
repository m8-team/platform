package app

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/m8-team/platform-sdk/lifecycle"
	"net"
	"net/http"
	"time"
)

type HTTPConfig struct {
	Address, CertFile, KeyFile string
	InsecureLocal, Streaming   bool
	MaxBodyBytes               int64
}

// HTTP leaves routing and Connect options to the caller. Streaming is opt-in:
// whole-stream read/write deadlines are disabled; handlers need per-message budgets.
func HTTP(name string, cfg HTTPConfig, handler http.Handler) (lifecycle.Component, error) {
	if cfg.Address == "" || handler == nil {
		return lifecycle.Component{}, errors.New("HTTP address and handler required")
	}
	if !cfg.InsecureLocal && (cfg.CertFile == "" || cfg.KeyFile == "") {
		return lifecycle.Component{}, errors.New("TLS certificate/key required")
	}
	if cfg.MaxBodyBytes <= 0 {
		cfg.MaxBodyBytes = 1 << 20
	}
	handler = http.MaxBytesHandler(handler, cfg.MaxBodyBytes)
	server := &http.Server{Addr: cfg.Address, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 32 << 10, TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12}}
	if cfg.Streaming {
		server.ReadTimeout = 0
		server.WriteTimeout = 0
	}
	var listener net.Listener
	closeListener := func() error {
		if listener == nil {
			return nil
		}
		err := listener.Close()
		if errors.Is(err, net.ErrClosed) {
			return nil
		}
		return err
	}
	return lifecycle.Component{Name: name, Start: func(ctx context.Context) error {
		if !cfg.InsecureLocal {
			cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
			if err != nil {
				return err
			}
			server.TLSConfig.Certificates = []tls.Certificate{cert}
		}
		var err error
		listener, err = (&net.ListenConfig{}).Listen(ctx, "tcp", cfg.Address)
		return err
	}, Run: func(context.Context) error {
		var err error
		if cfg.InsecureLocal {
			err = server.Serve(listener)
		} else {
			err = server.ServeTLS(listener, "", "")
		}
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}, Stop: func(ctx context.Context) error {
		err := server.Shutdown(ctx)
		if err != nil {
			return errors.Join(err, server.Close(), closeListener())
		}
		// A later component may fail before Serve ever registered this listener.
		return closeListener()
	}, Force: func() error { return errors.Join(server.Close(), closeListener()) }}, nil
}
