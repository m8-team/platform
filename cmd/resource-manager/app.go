package main

import (
 "errors"
 "net"
	"github.com/m8-team/platform/internal/platform/health"
	grpcserver "github.com/m8-team/platform/internal/platform/server/grpc"
	"github.com/m8-team/platform/internal/resourcemanager"
	grpcadapter "github.com/m8-team/platform/internal/resourcemanager/adapter/grpc"
	"go.uber.org/fx"
)

func NewApp(cfg Config) *fx.App {
	return fx.New(appOptions(cfg)...)
}

func appOptions(cfg Config) []fx.Option {
 // Until M8 Access is wired, local bypass may only be used on loopback.
 // Never allow a remote caller to acquire tenant-wide administration.
 if cfg.AllowUnauthenticated && (!loopbackOnly(cfg.HTTP.Address)||!loopbackOnly(cfg.GRPC.Address)){
  return []fx.Option{fx.Error(errors.New("M8_RM_ALLOW_UNAUTHENTICATED requires loopback-only HTTP and gRPC listeners"))}
 }
	options := []fx.Option{
		health.FxModule,
		resourceManagerHTTPModule(cfg.HTTP),
		healthHTTPModule(cfg.HealthHTTP),
		grpcserver.Module(cfg.GRPC),
		resourcemanager.Module(resourcemanager.Config{
			ServiceName:          "resource-manager",
			Debug:                cfg.Debug,
			AllowUnauthenticated: cfg.AllowUnauthenticated,
			SoftDeleteRetention:  cfg.SoftDeleteRetention,
			PageTokenKey:         cfg.PageTokenKey,
			YDBDSN: cfg.YDBDSN,
			YDBTablePrefix: cfg.YDBTablePrefix,
			AccessModelRevision: cfg.AccessModelRevision,
		}),
		grpcadapter.Module(),
	}

	if !cfg.Debug {
		options = append([]fx.Option{fx.NopLogger}, options...)
	}

	return options
}

func loopbackOnly(address string)bool{
 host,_,err:=net.SplitHostPort(address)
 if err!=nil{return false}
 return host=="localhost" || (net.ParseIP(host)!=nil && net.ParseIP(host).IsLoopback())
}
