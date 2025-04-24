package test

import (
	"context"
	"sync"
	"time"

	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/task/queue"
)

type CountingRunner struct {
	Type  string
	count int
	mu    *sync.Mutex
}

func NewCountingRunner(typ string) *CountingRunner {
	return &CountingRunner{
		Type:  typ,
		count: 0,
		mu:    &sync.Mutex{},
	}
}
func (c *CountingRunner) GetRunnerType() string {
	return c.Type
}

func (c *CountingRunner) GetRunnerDeadline() time.Time {
	return time.Now().Add(time.Second * 10)
}

func (c *CountingRunner) GetRunnerTimeout() time.Duration {
	return time.Second * 5
}

func (c *CountingRunner) GetRunnerDurationMaximum() time.Duration {
	return time.Second
}

func (c *CountingRunner) Run(ctx context.Context, tsk *task.Task) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.count += 1
	tsk.SetCompleted()
}

func (c *CountingRunner) GetCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

var _ queue.Runner = &CountingRunner{}
