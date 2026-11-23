package ydbcheckpoint

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/access/app/projection"
)

func TestRejectsUnconfiguredDistributedStore(t *testing.T) {
	cases := []struct {
		prefix, owner string
		ttl           time.Duration
	}{
		{"bad-prefix", "owner", time.Second * 10},
		{"/m8", "", time.Second * 10},
		{"/m8", "owner", time.Second},
	}
	for _, tc := range cases {
		if _, err := New(&sql.DB{}, tc.prefix, tc.owner, tc.ttl); !errors.Is(err, ErrInvalidStore) {
			t.Fatalf("bad configuration allowed: %+v %v", tc, err)
		}
	}
}
func TestDistributedStoreRejectsUnfencedLegacyWrites(t *testing.T) {
	s, _ := New(&sql.DB{}, "/m8", "worker-1", 10*time.Second)
	err := s.WithLocked(context.Background(), projection.Key{}, nil)
	if !errors.Is(err, ErrLeaseLost) {
		t.Fatalf("unfenced write allowed: %v", err)
	}
}
