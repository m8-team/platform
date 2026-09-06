package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
)

func TestHierarchyLockCancellationAndIndependentParents(t *testing.T) {
	repository := NewWorkspaceRepository()
	parent := organization.NewID()
	other := organization.NewID()
	entered, release, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	go func() {
		done <- repository.WithOrganizationLock(context.Background(), parent, func(context.Context) error { close(entered); <-release; return nil })
	}()
	<-entered
	defer func() {
		close(release)
		if err := <-done; err != nil {
			t.Error(err)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := repository.WithOrganizationLock(ctx, other, func(context.Context) error { return nil }); err != nil {
		t.Fatalf("unrelated parent blocked: %v", err)
	}
	canceled, stop := context.WithCancel(context.Background())
	waiting := make(chan error, 1)
	go func() {
		waiting <- repository.WithOrganizationLock(canceled, parent, func(context.Context) error { t.Error("canceled waiter entered"); return nil })
	}()
	stop()
	select {
	case err := <-waiting:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("waiter = %v", err)
		}
	case <-ctx.Done():
		t.Fatal("lock wait did not respect cancellation")
	}
}
