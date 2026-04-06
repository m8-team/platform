package app

import (
	"context"
	"errors"
)

type Server interface {
	Serve() error
	Shutdown(ctx context.Context) error
}

type CloseFunc func(context.Context) error

type Application struct {
	server   Server
	closeFns []CloseFunc
}

func New(server Server, closeFns ...CloseFunc) *Application {
	return &Application{
		server:   server,
		closeFns: append([]CloseFunc(nil), closeFns...),
	}
}

func (a *Application) Serve(ctx context.Context) error {
	if a == nil || a.server == nil {
		return nil
	}
	go func() {
		<-ctx.Done()
		_ = a.server.Shutdown(context.Background())
	}()
	err := a.server.Serve()
	if err != nil && errors.Is(ctx.Err(), context.Canceled) {
		return nil
	}
	return err
}

func (a *Application) Close(ctx context.Context) error {
	if a == nil {
		return nil
	}
	var closeErr error
	for i := len(a.closeFns) - 1; i >= 0; i-- {
		if a.closeFns[i] == nil {
			continue
		}
		if err := a.closeFns[i](ctx); err != nil {
			closeErr = errors.Join(closeErr, err)
		}
	}
	return closeErr
}
