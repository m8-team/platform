package event

import (
	"context"
	"sync/atomic"
	"time"
)

// TimingObservers connects real observations to service-owned histograms.
// Handlers can be called concurrently and must return promptly.
type TimingObservers struct {
	LaneWait            func(context.Context, time.Duration)
	TransactionDuration func(context.Context, time.Duration)
}

// PublishStats counts calls and staged batches for one Publisher. The
// transaction owner observes commit outcomes separately.
type PublishStats struct {
	Attempts, Errors, StagedBatches, StagedEvents, StagedBytes uint64
	OwnershipRejected, LaneAcquireErrors                       uint64
	LaneWait                                                   time.Duration
	LaneInUse, LaneCapacity                                    int
}

type publishCounters struct {
	attempts, errors, batches, events, bytes atomic.Uint64
}

// Stats returns cumulative counters for the service's metrics exporter.
func (p *Publisher) Stats() PublishStats {
	return PublishStats{
		Attempts: p.stats.attempts.Load(), Errors: p.stats.errors.Load(),
		StagedBatches: p.stats.batches.Load(), StagedEvents: p.stats.events.Load(),
		StagedBytes:       p.stats.bytes.Load(),
		OwnershipRejected: p.ownershipRejected.Load(),
		LaneAcquireErrors: p.laneErrors.Load(),
		LaneWait:          time.Duration(p.laneWait.Load()),
		LaneInUse:         len(p.lane),
		LaneCapacity:      cap(p.lane),
	}
}

func (c *publishCounters) record(err error, events, size int) {
	c.attempts.Add(1)
	if err != nil {
		c.errors.Add(1)
		return
	}
	if events > 0 {
		c.batches.Add(1)
		c.events.Add(uint64(events))
		c.bytes.Add(uint64(size))
	}
}

// TransactionStats counts SDK attempts and final outcomes of DoTx invocations.
// Failed includes unknown outcomes. CommitUnknown counts explicit YDB
// UNDETERMINED commit responses and is a lower bound on ambiguous commits.
type TransactionStats struct {
	Attempts, Retries, Committed, Failed, CommitUnknown      uint64
	Duration                                                 time.Duration
	OwnershipHandoffs, OwnershipRetirements, CommitConflicts uint64
}

type transactionCounters struct {
	attempts, retries, committed, failed, unknown, duration atomic.Uint64
	handoffs, retirements, commitConflicts                  atomic.Uint64
}

// Stats returns cumulative counters for the service's metrics exporter.
func (t *Transactor) Stats() TransactionStats {
	c := &t.counters
	return TransactionStats{
		Attempts: c.attempts.Load(), Retries: c.retries.Load(),
		Committed: c.committed.Load(), Failed: c.failed.Load(),
		CommitUnknown: c.unknown.Load(), Duration: time.Duration(c.duration.Load()),
		OwnershipHandoffs:    c.handoffs.Load(),
		OwnershipRetirements: c.retirements.Load(),
		CommitConflicts:      c.commitConflicts.Load(),
	}
}

func (c *transactionCounters) record(err error, duration time.Duration) {
	c.duration.Add(uint64(duration))
	if err == nil {
		c.committed.Add(1)
		return
	}
	c.failed.Add(1)
}
