// Package temporal adapts a native Temporal worker without wrapping workflows.
package temporal

import (
	"context"
	"errors"
	"github.com/m8-team/platform-sdk/lifecycle"
)

// Worker is implemented by worker.Worker from go.temporal.io/sdk/worker.
// Set WorkerStopTimeout on that native worker to fit the application's budget.
type Worker interface {
	Start() error
	Stop()
}

func Component(name string, w Worker) (lifecycle.Component, error) {
	if w == nil {
		return lifecycle.Component{}, errors.New("temporal worker required")
	}
	return lifecycle.Component{Name: name, Start: func(context.Context) error { return w.Start() }, Stop: func(context.Context) error { w.Stop(); return nil }}, nil
}
