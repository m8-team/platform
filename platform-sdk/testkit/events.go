// Package testkit contains small explicit fakes, not a mocking framework.
package testkit

import (
	"context"
	"encoding/json"
	"github.com/m8-team/platform-sdk/outbox"
	"io"
	"log/slog"
	"sync"
)

// Recorder implements outbox.Publisher. Snapshots are deep copies so tests cannot
// race with the publisher by mutating event payloads and headers.
type Recorder struct {
	mu     sync.Mutex
	events []outbox.Event
}

func (r *Recorder) Publish(ctx context.Context, event outbox.Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, clone(event))
	return nil
}
func (r *Recorder) Events() []outbox.Event {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]outbox.Event, len(r.events))
	for i, e := range r.events {
		result[i] = clone(e)
	}
	return result
}
func clone(e outbox.Event) outbox.Event {
	e.Payload = append([]byte(nil), e.Payload...)
	headers := map[string]string{}
	for k, v := range e.Headers {
		headers[k] = v
	}
	e.Headers = headers
	return e
}
func Logger(writer io.Writer) *slog.Logger { return slog.New(slog.NewJSONHandler(writer, nil)) }

// JSON is useful for inspecting recorded wire payloads without generics or reflection DI.
func JSON(event outbox.Event, destination any) error {
	return json.Unmarshal(event.Payload, destination)
}
