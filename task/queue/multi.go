package queue

import (
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/task/store"
)

// MultiQueue creates a queue per registered runner
type MultiQueue struct {
	queues map[string]Queue
	cfg    *Config
	lgr    log.Logger
	str    store.Store
}

func NewMultiQueue(cfg *Config, lgr log.Logger, str store.Store) (Queue, error) {
	return &MultiQueue{
		queues: make(map[string]Queue),
		cfg:    cfg,
		lgr:    lgr,
		str:    str,
	}, nil
}

func (m *MultiQueue) RegisterRunner(runner Runner) error {
	typ := runner.GetRunnerType()
	if _, ok := m.queues[typ]; ok {
		return errors.New("runner of the same type is already registered")
	}

	str := m.str.WithTypeFilter(typ)
	q, err := New(m.cfg, m.lgr, str)
	if err != nil {
		return err
	}
	if err := q.RegisterRunner(runner); err != nil {
		return err
	}

	m.queues[typ] = q
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

func (m *MultiQueue) GetQueues() map[string]Queue {
	return m.queues
}

var _ Queue = &MultiQueue{}
