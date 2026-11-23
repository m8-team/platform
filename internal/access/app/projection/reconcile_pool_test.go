package projection

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

func TestReconcilePoolBoundsConcurrencyAndDeduplicates(t *testing.T) {
	var running atomic.Int64
	var peak atomic.Int64
	var mu sync.Mutex
	seen := make(map[Key]int)
	keys := make([]Key, 0, 49)
	for i := 1; i <= 48; i++ {
		keys = append(keys, testKey(t, fmt.Sprintf("%08x-1111-4111-8111-111111111111", i)))
	}
	keys = append(keys, keys[0])
	err := reconcileKeys(context.Background(), keys, 4, func(ctx context.Context, key Key) error {
		current := running.Add(1)
		for {
			high := peak.Load()
			if current <= high || peak.CompareAndSwap(high, current) {
				break
			}
		}
		defer running.Add(-1)
		mu.Lock()
		seen[key]++
		mu.Unlock()
		return ctx.Err()
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(seen) != 48 || peak.Load() > 4 {
		t.Fatalf("work was duplicated or concurrency unbounded: keys=%d max=%d", len(seen), peak.Load())
	}
	for key, count := range seen {
		if count != 1 {
			t.Fatalf("key %v handled %d times", key, count)
		}
	}
}

func TestReconcilePoolFailsClosed(t *testing.T) {
	key := testKey(t, "11111111-1111-4111-8111-111111111111")
	for _, size := range []int{0, -1, MaxReconcileWorkers + 1} {
		if err := reconcileKeys(context.Background(), []Key{key}, size, func(context.Context, Key) error { return nil }); !errors.Is(err, ErrInvalidReconcilePool) {
			t.Fatalf("workers=%d: %v", size, err)
		}
	}
	called := false
	if err := reconcileKeys(context.Background(), []Key{{}}, 1, func(context.Context, Key) error {
		called = true
		return nil
	}); err == nil || called {
		t.Fatalf("invalid key had side effects: %v", err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := reconcileKeys(cancelled, []Key{key}, 2, func(context.Context, Key) error { return nil }); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation ignored: %v", err)
	}
	wanted := errors.New("projection failed")
	if err := reconcileKeys(context.Background(), []Key{key}, 1, func(context.Context, Key) error { return wanted }); !errors.Is(err, wanted) {
		t.Fatalf("error lost across workers: %v", err)
	}
}
