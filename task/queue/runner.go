package queue

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/task"
)

// Runner processes one task type on the queue. This comment is the complete contract; no other
// queue code need be consulted to implement it.
//
// Registration: one Runner per task type, registered at queue construction (duplicate types
// fail construction). The queue owns the single instance for the queue's lifetime.
//
// Concurrency: the queue runs multiple workers (configurable; 5 by default) that share the one Runner instance,
// so Run is called concurrently from multiple goroutines — any state shared across calls must be
// synchronized. Each call's *task.Task is distinct and call-owned: don't retain it past Run or
// touch another call's task. The Get* methods are also called concurrently and must be stable
// (in practice, constants).
//
// Duration contract, validated at construction (violations fail construction):
//
//	GetRunnerDeadline() > GetRunnerTimeout() > GetRunnerDurationMaximum() > 0
//
// Run's obligation: the task arrives running and already claimed under a unique per-run state
// lock — a token the queue set when it claimed the task; a later write-back to the task only lands
// while that same lock is still in place (see the lost-write note under Delivery). Before
// returning, Run MUST move it out of running by calling exactly one terminal helper on tsk:
//
//   - tsk.SetCompleted()          — success; done.
//   - tsk.SetFailed()             — permanent failure; done.
//   - tsk.SetFailedWithError(e)   — permanent failure with error; done.
//   - tsk.RepeatAvailableAfter(d) — reschedule after duration d.
//   - tsk.RepeatAvailableAt(t)    — reschedule at time t.
//
// The task arrives with AvailableTime cleared, so reschedule via the Repeat* helpers (which set
// both pending state and a new available time); a reschedule time in the past is clamped to now
// (run ASAP). Setting pending without an available time is a bug the queue only papers over with
// a warning. If Run returns with the task still running, the
// queue: reverts it to pending if interrupted by shutdown (parent context canceled); else fails
// it, recording the timeout as cause if the context was canceled for timeout, otherwise a
// "runner failed to set state" error.
//
// Context: canceled (with cause) after GetRunnerTimeout and on shutdown — cooperative runners
// select on ctx.Done() and return promptly. To tell the two apart, inspect context.Cause(ctx): a
// timeout cancels with ErrRunnerTimeout (errors.Is(context.Cause(ctx), ErrRunnerTimeout)), a
// shutdown with context.Canceled. The queue cannot preempt a runner that ignores cancellation; it
// stays blocked until it returns or the process restarts, recovered by the deadline/unstick
// mechanism (a periodic sweep that resets tasks still running past their deadline — see
// GetRunnerDeadline). The logger is on the context (log.LoggerFromContext(ctx)). Wrap must-complete writes
// in context.WithoutCancel so cancellation doesn't abandon them.
//
// Delivery is at-least-once, so Run must tolerate re-execution of the same logical task: a runner
// that overruns its deadline is unstuck and re-dispatched (possibly while the original run is still
// in flight), and a crash or kill mid-run re-runs the task from the start. Conversely a final write
// can be lost — if the task was unstuck and re-claimed while Run was still working, the queue's
// write-back on return no longer matches the state lock and is dropped (logged and counted, not
// applied). Design terminal effects to be idempotent and safe to repeat.
//
// Errors: accumulate with tsk.AppendError(err), reset with tsk.ClearError() (commonly at run
// start). Setting an error does not change state — pair it with a terminal helper. Persisted with
// the task; logged at error level if failed, warning if rescheduled. A panic is recovered and an
// "unhandled panic" error attached; the task is then failed unless the runner had already moved it
// out of running before panicking (e.g. a task completed before the panic stays completed, with
// the error attached) or a concurrent shutdown reverts it to pending instead.
//
// Persistence: on return the queue persists only the terminal state, the available time (written
// only when rescheduling; a terminal task's available time was already cleared when the task was
// claimed), tsk.Data (mutate the map in place, or assign a new map if it arrived nil; the map
// is written as a whole-field replacement, so its final contents are exactly what is stored — keys
// dropped from the map are gone, and a nil map unsets the field), and the accumulated error —
// mutate only these. Run time and duration are recorded by the queue; other fields are
// queue-managed, don't set them. To persist Data mid-run, mutate tsk
// and let the queue write it back on return; a Runner that instead updates its OWN task via
// task.Client must replace tsk with the returned task or the queue's write-back clobbers it (see
// task.Client.UpdateTask).
type Runner interface {
	// GetRunnerType returns the task type this runner processes. Unique across a queue's runners;
	// must match task.Task.Type of the tasks it handles.
	GetRunnerType() string

	// GetRunnerDeadline returns the duration, measured from when the task starts running, after
	// which a still-running task is forcibly reset to pending/available by the unstick mechanism,
	// recovering tasks orphaned by crashes, kills, or a runner that never returns. The reset is not
	// immediate: unstick runs on a periodic sweep, so a task is recovered on the first sweep after
	// its deadline elapses. Must exceed GetRunnerTimeout so a task still within its timeout is never
	// unstuck and re-claimed mid-run.
	GetRunnerDeadline() time.Duration

	// GetRunnerTimeout returns the hard cap on a run; at elapse the Run context is canceled so a
	// cooperative runner can return. Must exceed GetRunnerDurationMaximum.
	GetRunnerTimeout() time.Duration

	// GetRunnerDurationMaximum returns the expected upper bound of a normal run. Advisory: exceeding
	// it logs a warning but does not interrupt the run. Must be positive.
	GetRunnerDurationMaximum() time.Duration

	// Run executes tsk within ctx. Before returning it must move tsk out of running (completed,
	// failed, or rescheduled via Repeat*); a task left running is failed by the queue, unless
	// interrupted by shutdown, in which case it is reverted to pending. ctx is canceled after
	// GetRunnerTimeout and on shutdown. Called concurrently for distinct tasks; must be concurrency-safe.
	Run(ctx context.Context, tsk *task.Task)
}
