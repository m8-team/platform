package event

import "sync/atomic"

// PublishStats counts calls and staged batches for one Publisher. The
// transaction owner observes commit outcomes separately.
type PublishStats struct {
	Attempts, Errors, StagedBatches, StagedEvents, StagedBytes uint64
}

type publishCounters struct {
	attempts, errors, batches, events, bytes atomic.Uint64
}

// Stats returns cumulative counters for the service's metrics exporter.
func (p *Publisher) Stats() PublishStats {
	return PublishStats{
		Attempts: p.stats.attempts.Load(), Errors: p.stats.errors.Load(),
		StagedBatches: p.stats.batches.Load(), StagedEvents: p.stats.events.Load(),
		StagedBytes: p.stats.bytes.Load(),
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
