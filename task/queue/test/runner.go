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

type PanicRunner struct {
	Type string
}

func NewPanicRunner(typ string) *PanicRunner {
	return &PanicRunner{
		Type: typ,
	}
}

func (p *PanicRunner) GetRunnerType() string {
	return p.Type
}

func (p *PanicRunner) GetRunnerDeadline() time.Duration {
	return p.GetRunnerDurationMaximum() * 5
}

func (p *PanicRunner) GetRunnerTimeout() time.Duration {
	return p.GetRunnerDurationMaximum() * 3
}

func (p *PanicRunner) GetRunnerDurationMaximum() time.Duration {
	return time.Second
}

func (p *PanicRunner) Run(ctx context.Context, tsk *task.Task) {
	panic("panic test")
}

var _ queue.Runner = &PanicRunner{}

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

type RepeatRunner struct {
	Type string
}

func NewRepeatRunner(typ string) *RepeatRunner {
	return &RepeatRunner{
		Type: typ,
	}
}

func (r *RepeatRunner) GetRunnerType() string {
	return r.Type
}

func (r *RepeatRunner) GetRunnerDeadline() time.Duration {
	return r.GetRunnerDurationMaximum() * 5
}

func (r *RepeatRunner) GetRunnerTimeout() time.Duration {
	return r.GetRunnerDurationMaximum() * 3
}

func (r *RepeatRunner) GetRunnerDurationMaximum() time.Duration {
	return time.Minute
}

func (r *RepeatRunner) Run(ctx context.Context, tsk *task.Task) {
	tsk.State = task.TaskStatePending
}

var _ queue.Runner = &RepeatRunner{}

type HangingRunner struct {
	Type    string
	Timeout time.Duration
}

func NewHangingRunner(typ string, timeout time.Duration) *HangingRunner {
	return &HangingRunner{
		Type:    typ,
		Timeout: timeout,
	}
}

func (h *HangingRunner) GetRunnerType() string {
	return h.Type
}

func (h *HangingRunner) GetRunnerDeadline() time.Duration {
	return h.Timeout * 5
}

func (h *HangingRunner) GetRunnerTimeout() time.Duration {
	return h.Timeout
}

func (h *HangingRunner) GetRunnerDurationMaximum() time.Duration {
	return h.Timeout / 2
}

// Run blocks until its context is canceled and returns without setting a terminal
// state, leaving the task running to exercise shutdown interruption and timeout handling.
func (h *HangingRunner) Run(ctx context.Context, tsk *task.Task) {
	<-ctx.Done()
}

var _ queue.Runner = &HangingRunner{}

type BlockingRunner struct {
	Type            string
	DurationMaximum time.Duration
}

func NewBlockingRunner(typ string, durationMaximum time.Duration) *BlockingRunner {
	return &BlockingRunner{
		Type:            typ,
		DurationMaximum: durationMaximum,
	}
}

func (b *BlockingRunner) GetRunnerType() string {
	return b.Type
}

func (b *BlockingRunner) GetRunnerDeadline() time.Duration {
	return b.GetRunnerDurationMaximum() * 5
}

func (b *BlockingRunner) GetRunnerTimeout() time.Duration {
	return b.GetRunnerDurationMaximum() * 3
}

func (b *BlockingRunner) GetRunnerDurationMaximum() time.Duration {
	return b.DurationMaximum
}

// Run blocks forever, ignoring context cancellation entirely, to simulate a runner that
// does not honor shutdown. Used to verify Stop returns within its timeout regardless.
func (b *BlockingRunner) Run(ctx context.Context, tsk *task.Task) {
	select {}
}

var _ queue.Runner = &BlockingRunner{}

type CallbackRunner struct {
	Type     string
	Callback func(ctx context.Context, tsk *task.Task)
}

func NewCallbackRunner(typ string, callback func(ctx context.Context, tsk *task.Task)) *CallbackRunner {
	return &CallbackRunner{
		Type:     typ,
		Callback: callback,
	}
}

func (c *CallbackRunner) GetRunnerType() string {
	return c.Type
}

func (c *CallbackRunner) GetRunnerDeadline() time.Duration {
	return c.GetRunnerDurationMaximum() * 5
}

func (c *CallbackRunner) GetRunnerTimeout() time.Duration {
	return c.GetRunnerDurationMaximum() * 3
}

func (c *CallbackRunner) GetRunnerDurationMaximum() time.Duration {
	return time.Minute
}

// Run invokes the callback, letting a test inject arbitrary behavior (such as updating the
// task while it is running) into the task run.
func (c *CallbackRunner) Run(ctx context.Context, tsk *task.Task) {
	if c.Callback != nil {
		c.Callback(ctx, tsk)
	}
}

var _ queue.Runner = &CallbackRunner{}
