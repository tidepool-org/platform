package queue

import (
	"maps"
	"sync"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/task/store"
)

// MultiQueue runs a queue per runner provided at construction, each processing only tasks of
// that runner's type. Like Queue, it is single-use. The queues map is immutable after
// NewMultiQueue, so no synchronization is required.
type MultiQueue struct {
	queues map[string]*Queue
}

func NewMultiQueue(cfg *Config, lgr log.Logger, str store.Store, runners ...Runner) (*MultiQueue, error) {
	if cfg == nil {
		return nil, errors.New("config is missing")
	}
	if lgr == nil {
		return nil, errors.New("logger is missing")
	}
	if str == nil {
		return nil, errors.New("store is missing")
	}

	queues := make(map[string]*Queue, len(runners))
	for _, runner := range runners {
		if runner == nil {
			return nil, errors.New("runner is missing")
		}

		typ := runner.GetRunnerType()
		if _, ok := queues[typ]; ok {
			return nil, errors.New("runner type already registered")
		}

		q, err := New(typ, cfg, lgr, str.WithTypeFilter(typ), runner)
		if err != nil {
			return nil, err
		}

		queues[typ] = q
	}

	return &MultiQueue{
		queues: queues,
	}, nil
}

func (m *MultiQueue) Start() {
	for _, q := range m.queues {
		q.Start()
	}
}

func (m *MultiQueue) Stop() {
	// Stop the queues concurrently so total shutdown latency is bounded by a single
	// queue's stop timeout rather than the sum across every queue.
	var waitGroup sync.WaitGroup
	for _, q := range m.queues {
		waitGroup.Go(q.Stop)
	}
	waitGroup.Wait()
}

// GetQueues returns a copy of the type-to-queue map so callers cannot mutate the
// internal map. The queue values are shared references, not copies.
func (m *MultiQueue) GetQueues() map[string]*Queue {
	return maps.Clone(m.queues)
}
