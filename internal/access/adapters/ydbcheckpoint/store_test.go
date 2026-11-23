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

func TestCheckpointCertification(t *testing.T) {
	for _, tc := range []struct {
		name               string
		exists             bool
		revision, required int64
		fence, lease       uint64
		token              string
		want               bool
	}{
		{"current", true, 5, 5, 3, 3, "opaque", true},
		{"ahead", true, 6, 5, 3, 3, "opaque", true},
		{"behind", true, 4, 5, 3, 3, "opaque", false},
		{"missing row", false, 5, 5, 3, 3, "opaque", false},
		{"missing requirement", true, 5, 0, 3, 3, "opaque", false},
		{"missing token", true, 5, 5, 3, 3, "", false},
		{"unfenced", true, 5, 5, 0, 0, "opaque", false},
		{"lease takeover", true, 5, 5, 3, 4, "opaque", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := row{exists: tc.exists, leaseEpoch: tc.lease, cp: projection.Checkpoint{
				Revision: tc.revision, FenceEpoch: tc.fence, ZedToken: tc.token,
			}}
			if got := r.certifiedAt(tc.required); got != tc.want {
				t.Fatalf("certifiedAt(%d) = %t, want %t", tc.required, got, tc.want)
			}
		})
	}
}
