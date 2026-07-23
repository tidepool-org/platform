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

func (c *CountingRunner) GetRunnerDeadline() time.Duration {
	return time.Second * 10
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

type SleepRunner struct {
	Type     string
	Deadline time.Duration
	Timeout  time.Duration
	Maximum  time.Duration
	Sleep    time.Duration
}

func NewSleepRunner(typ string, deadline time.Duration, timeout time.Duration, maximum time.Duration, sleep time.Duration) *SleepRunner {
	return &SleepRunner{
		Type:     typ,
		Deadline: deadline,
		Timeout:  timeout,
		Maximum:  maximum,
		Sleep:    sleep,
	}
}

func (s *SleepRunner) GetRunnerType() string {
	return s.Type
}

func (s *SleepRunner) GetRunnerDeadline() time.Duration {
	return s.Deadline
}

func (s *SleepRunner) GetRunnerTimeout() time.Duration {
	return s.Timeout
}

func (s *SleepRunner) GetRunnerDurationMaximum() time.Duration {
	return s.Maximum
}

func (s *SleepRunner) Run(ctx context.Context, tsk *task.Task) {
	select {
	case <-ctx.Done():
		tsk.AppendError(context.Cause(ctx))
		tsk.SetCompleted()
	case <-time.After(s.Sleep):
		tsk.SetCompleted()
	}
}

var _ queue.Runner = &SleepRunner{}

type StubRunner struct {
	Type            string
	Stub            func(ctx context.Context, tsk *task.Task)
	deadline        *time.Duration
	timeout         *time.Duration
	durationMaximum *time.Duration
}

func NewStubRunner(typ string) *StubRunner {
	return &StubRunner{
		Type: typ,
	}
}

func (s *StubRunner) GetRunnerType() string {
	return s.Type
}

func (s *StubRunner) GetRunnerDeadline() time.Duration {
	if s.deadline != nil {
		return *s.deadline
	} else {
		return s.GetRunnerDurationMaximum() * 4
	}
}

func (s *StubRunner) GetRunnerTimeout() time.Duration {
	if s.timeout != nil {
		return *s.timeout
	} else {
		return s.GetRunnerDurationMaximum() * 2
	}
}

func (s *StubRunner) GetRunnerDurationMaximum() time.Duration {
	if s.durationMaximum != nil {
		return *s.durationMaximum
	} else {
		return time.Minute
	}
}

func (s *StubRunner) Run(ctx context.Context, tsk *task.Task) {
	if s.Stub != nil {
		s.Stub(ctx, tsk)
	}
}

func (s *StubRunner) WithStub(stub func(ctx context.Context, tsk *task.Task)) *StubRunner {
	s.Stub = stub
	return s
}

func (s *StubRunner) WithDeadline(deadline time.Duration) *StubRunner {
	s.deadline = &deadline
	return s
}

func (s *StubRunner) WithTimeout(timeout time.Duration) *StubRunner {
	s.timeout = &timeout
	return s
}

func (s *StubRunner) WithDurationMaximum(durationMaximum time.Duration) *StubRunner {
	s.durationMaximum = &durationMaximum
	return s
}

var _ queue.Runner = &StubRunner{}
