package queue

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/task/store"
)

// MultiQueue creates a queue per registered runner
type MultiQueue struct {
	queues []Queue
	cfg    *Config
	lgr    log.Logger
	str    store.Store
}

func NewMultiQueue(cfg *Config, lgr log.Logger, str store.Store) (Queue, error) {
	return &MultiQueue{
		cfg: cfg,
		lgr: lgr,
		str: str,
	}, nil
}

func (m *MultiQueue) RegisterRunner(runner Runner) error {
	str := m.str.WithTypeFilter(runner.GetRunnerType())
	q, err := New(m.cfg, m.lgr, str)
	if err != nil {
		return err
	}
	if err := q.RegisterRunner(runner); err != nil {
		return err
	}

	m.queues = append(m.queues, q)
	return nil
}

func (m *MultiQueue) Start() {
	for _, q := range m.queues {
		q := q
		q.Start()
	}
}

func (m *MultiQueue) Stop() {
	for _, q := range m.queues {
		q := q
		q.Stop()
	}
}

var _ Queue = &MultiQueue{}
