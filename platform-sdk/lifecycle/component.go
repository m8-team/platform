// Package lifecycle describes explicitly ordered resources and runtime tasks.
package lifecycle

import "context"

// Components are registered in dependency order and stopped in reverse order.
// Start must roll back its own partial initialization on failure. Run must stop
// on cancellation; Stop must honor its deadline. Force closes owned blocking IO
// and must return immediately. Start's context is valid for initialization only.
type Component struct {
	Name  string
	Start func(context.Context) error
	Run   func(context.Context) error
	Stop  func(context.Context) error
	Force func() error
}
