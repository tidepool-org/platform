package queue

import (
	"context"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"github.com/tidepool-org/platform/task"
	taskStore "github.com/tidepool-org/platform/task/store"
)

const (
	// WorkersDefault is the default number of workers for the queue.
	WorkersDefault = 5

	// StartManagerDelayDefault is the default upper bound on the randomized delay the manager waits before it begins
	// dispatching tasks, spreading startup across instances.
	StartManagerDelayDefault = 1 * time.Minute

	// DispatchTasksDelayDefault is the default delay, jittered, between polls for pending tasks to dispatch. Completing
	// a task also dispatches immediately, without waiting for the poll.
	DispatchTasksDelayDefault = 1 * time.Minute

	// MonitorTaskDelayDefault is the default interval between checks that each in-flight task's claim is still held
	// (the task exists and its claim token is unchanged); a run whose claim is lost is canceled with ErrClaimLost.
	MonitorTaskDelayDefault = 1 * time.Minute

	// RunnerWatchdogGracePeriodDefault is the extra time beyond the runner timeout that the watchdog waits before
	// reporting a runner as blocked. The runner context is still canceled at the runner timeout; the grace period only
	// gives a cooperative runner time to observe that cancellation and return before the watchdog reports it as
	// non-cooperative.
	RunnerWatchdogGracePeriodDefault = 5 * time.Second

	// UnstickTasksDelayDefault is the default delay, jittered, between attempts to unstick tasks. The first attempt is
	// made after a randomized delay of at most this duration.
	UnstickTasksDelayDefault = 5 * time.Minute

	// UnstickTasksAvailableGracePeriodDefault is the default grace period added to the claim monitor delay and the
	// runner watchdog grace period to form the delay after a task is unstuck before it is made available for
	// re-dispatch. This ensures that any still-running tasks have likely been canceled and exited before being
	// re-dispatched.
	UnstickTasksAvailableGracePeriodDefault = 1 * time.Minute

	// StopWaitTimeoutDefault bounds how long Stop waits for in-flight tasks to observe cancellation and exit before
	// abandoning them (they are recovered by the deadline and unstick mechanism). Kept under the typical Kubernetes
	// termination grace period since we could use it up to twice (once for workers and once for manager) and we still
	// want to leave time for the store to flush any pending writes.
	StopWaitTimeoutDefault = 10 * time.Second

	TaskDeadlineDefault  = 1 * time.Minute
	DurationJitterFactor = 0.2
)

type Config struct {
	Workers                          int
	StartManagerDelay                time.Duration
	DispatchTasksDelay               time.Duration
	MonitorTaskDelay                 time.Duration
	RunnerWatchdogGracePeriod        time.Duration
	UnstickTasksDelay                time.Duration
	UnstickTasksAvailableGracePeriod time.Duration
	StopWaitTimeout                  time.Duration
}

func NewConfig() *Config {
	return &Config{
		Workers:                          WorkersDefault,
		StartManagerDelay:                StartManagerDelayDefault,
		DispatchTasksDelay:               DispatchTasksDelayDefault,
		MonitorTaskDelay:                 MonitorTaskDelayDefault,
		RunnerWatchdogGracePeriod:        RunnerWatchdogGracePeriodDefault,
		UnstickTasksDelay:                UnstickTasksDelayDefault,
		UnstickTasksAvailableGracePeriod: UnstickTasksAvailableGracePeriodDefault,
		StopWaitTimeout:                  StopWaitTimeoutDefault,
	}
}

func (c *Config) Load(configReporter config.Reporter) error {
	if configReporter == nil {
		return errors.New("config reporter is missing")
	}

	if valueString, err := configReporter.Get("workers"); err == nil {
		if value, parseErr := strconv.ParseInt(valueString, 10, 0); parseErr != nil {
			return errors.New("workers is invalid")
		} else {
			c.Workers = int(value)
		}
	}
	if valueString, err := configReporter.Get("start_manager_delay"); err == nil {
		if value, parseErr := strconv.ParseInt(valueString, 10, 0); parseErr != nil {
			return errors.New("start manager delay is invalid")
		} else {
			c.StartManagerDelay = time.Duration(value) * time.Second
		}
	}
	if valueString, err := configReporter.Get("dispatch_tasks_delay"); err == nil {
		if value, parseErr := strconv.ParseInt(valueString, 10, 0); parseErr != nil {
			return errors.New("dispatch tasks delay is invalid")
		} else {
			c.DispatchTasksDelay = time.Duration(value) * time.Second
		}
	}
	if valueString, err := configReporter.Get("monitor_task_delay"); err == nil {
		if value, parseErr := strconv.ParseInt(valueString, 10, 0); parseErr != nil {
			return errors.New("monitor task delay is invalid")
		} else {
			c.MonitorTaskDelay = time.Duration(value) * time.Second
		}
	}
	if valueString, err := configReporter.Get("runner_watchdog_grace_period"); err == nil {
		if value, parseErr := strconv.ParseInt(valueString, 10, 0); parseErr != nil {
			return errors.New("runner watchdog grace period is invalid")
		} else {
			c.RunnerWatchdogGracePeriod = time.Duration(value) * time.Second
		}
	}
	if valueString, err := configReporter.Get("unstick_tasks_delay"); err == nil {
		if value, parseErr := strconv.ParseInt(valueString, 10, 0); parseErr != nil {
			return errors.New("unstick tasks delay is invalid")
		} else {
			c.UnstickTasksDelay = time.Duration(value) * time.Second
		}
	}
	if valueString, err := configReporter.Get("unstick_tasks_available_grace_period"); err == nil {
		if value, parseErr := strconv.ParseInt(valueString, 10, 0); parseErr != nil {
			return errors.New("unstick tasks available grace period is invalid")
		} else {
			c.UnstickTasksAvailableGracePeriod = time.Duration(value) * time.Second
		}
	}
	if valueString, err := configReporter.Get("stop_wait_timeout"); err == nil {
		if value, parseErr := strconv.ParseInt(valueString, 10, 0); parseErr != nil {
			return errors.New("stop wait timeout is invalid")
		} else {
			c.StopWaitTimeout = time.Duration(value) * time.Second
		}
	}

	return nil
}

func (c *Config) Validate() error {
	if c.Workers < 1 {
		return errors.New("workers is invalid")
	}
	if c.StartManagerDelay <= 0 {
		return errors.New("start manager delay is invalid")
	}
	if c.DispatchTasksDelay <= 0 {
		return errors.New("dispatch tasks delay is invalid")
	}
	if c.MonitorTaskDelay <= 0 {
		return errors.New("monitor task delay is invalid")
	}
	if c.RunnerWatchdogGracePeriod <= 0 {
		return errors.New("runner watchdog grace period is invalid")
	}
	if c.UnstickTasksDelay <= 0 {
		return errors.New("unstick tasks delay is invalid")
	}
	if c.UnstickTasksAvailableGracePeriod <= 0 {
		return errors.New("unstick tasks available grace period is invalid")
	}
	if c.StopWaitTimeout <= 0 {
		return errors.New("stop wait timeout is invalid")
	}
	return nil
}

// The Queue's fields are all immutable after New, except the lifecycle fields, which are guarded by the lifecycle
// mutex, and workersAvailable, which is initialized by Start before the manager exists and thereafter owned exclusively
// by the manager goroutine. The workers and manager therefore read the channels and runners map freely, without
// synchronization.
type Queue struct {
	name              string
	config            *Config
	logger            log.Logger
	repository        taskStore.TaskRepository
	runners           map[string]Runner
	dispatchChannel   chan *task.Task
	completionChannel chan *task.Task
	lifecycleMutex    sync.Mutex
	started           bool
	stopped           bool
	cancelFunc        context.CancelFunc
	workersWaitGroup  sync.WaitGroup
	managerWaitGroup  sync.WaitGroup
	workersAvailable  int
}

func New(name string, cfg *Config, lgr log.Logger, str taskStore.Store, runners ...Runner) (*Queue, error) {
	if name == "" {
		return nil, errors.New("name is missing")
	}
	if cfg == nil {
		return nil, errors.New("config is missing")
	}
	if lgr == nil {
		return nil, errors.New("logger is missing")
	}
	if str == nil {
		return nil, errors.New("store is missing")
	}

	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}

	runnerMap := make(map[string]Runner, len(runners))
	for _, runner := range runners {
		if runner == nil {
			return nil, errors.New("runner is missing")
		}
		if _, ok := runnerMap[runner.GetRunnerType()]; ok {
			return nil, errors.New("runner type already registered")
		}
		if err := validateRunner(runner); err != nil {
			return nil, err
		}
		runnerMap[runner.GetRunnerType()] = runner
	}

	return &Queue{
		name:       name,
		config:     cfg,
		logger:     lgr.WithField("queue", name),
		repository: str.NewTaskRepository(),
		runners:    runnerMap,

		// NOT buffered so a task is only handed off when a worker is ready to receive it. This ensures a dispatched
		// task is never stranded in a buffer during shutdown.
		dispatchChannel: make(chan *task.Task),

		// Buffered so that a worker can complete a task and hand it off to the manager even if the manager is busy
		// dispatching other tasks.
		completionChannel: make(chan *task.Task, cfg.Workers),
	}, nil
}

func (q *Queue) Start() {
	q.lifecycleMutex.Lock()
	defer q.lifecycleMutex.Unlock()

	if q.started || q.stopped {
		return
	}
	q.started = true

	q.logger.Debug("Task queue starting")

	ctx, cancelFunc := context.WithCancel(log.NewContextWithLogger(context.Background(), q.logger))
	q.cancelFunc = cancelFunc

	q.startWorkers(ctx)
	q.startManager(ctx)

	q.logger.Debug("Task queue started")
}

func (q *Queue) Stop() {
	// Hold the mutex for the entire stop, including the waits, so a concurrent Start cannot observe a partially stopped
	// queue.
	q.lifecycleMutex.Lock()
	defer q.lifecycleMutex.Unlock()

	if q.stopped {
		return
	}
	q.stopped = true

	// Never started, so nothing to stop; the stopped flag ensures it never will be.
	if !q.started {
		return
	}

	q.logger.Info("Task queue stopping")

	lgr := q.logger.WithField("stopWaitTimeout", q.config.StopWaitTimeout)

	// Cancel the manager, so it stops dispatching new tasks and begins draining completions from the workers, and the
	// workers, to interrupt any in-flight task.
	q.cancelFunc()

	// Wait for all workers to exit, but only up to a bounded timeout so a runner that does not honor cancellation
	// cannot block shutdown forever. If a worker is still running we must NOT close the channels: a stuck worker that
	// later finishes would panic sending on a closed completion channel. Instead we leave the goroutines orphaned
	// (reaped at process exit); the abandoned task stays running and is recovered by the deadline/unstick mechanism.
	if !waitWithTimeout(&q.workersWaitGroup, q.config.StopWaitTimeout) {
		lgr.Error("Task queue workers did not stop within timeout; abandoning in-flight tasks; will be fixed with UnstickTasks later")
		return
	}

	// All workers have exited, so completion channel can be closed.
	close(q.completionChannel)

	// Wait for manager to exit. This should be prompt now that the completion channel is closed, but bound it too in
	// case a completion write is slow.
	if !waitWithTimeout(&q.managerWaitGroup, q.config.StopWaitTimeout) {
		lgr.Error("Task queue manager did not stop within timeout")
		return
	}

	// Manager has exited, so no further tasks will be dispatched. Because the dispatch channel is not buffered, no
	// dispatched task can be stranded in it; any task the manager could not hand off was reverted to pending during
	// dispatch.
	close(q.dispatchChannel)

	q.logger.Info("Task queue stopped")
}

func (q *Queue) startWorkers(ctx context.Context) {
	for q.workersAvailable = 0; q.workersAvailable < q.config.Workers; q.workersAvailable++ {
		q.startWorker(log.ContextWithField(ctx, "worker", q.workersAvailable))
	}
	WorkersAvailable.WithLabelValues(q.name).Set(float64(q.workersAvailable))
	WorkersTotal.WithLabelValues(q.name).Set(float64(q.config.Workers))
}

func (q *Queue) startWorker(ctx context.Context) {
	q.workersWaitGroup.Go(func() {
		lgr := log.LoggerFromContext(ctx)

		lgr.Debug("Task queue worker started")

		for {
			if err := q.executeWorker(ctx); err != nil {
				lgr.WithError(err).Debug("Task queue worker stopped")
				return
			}
		}
	})
}

func (q *Queue) executeWorker(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	default:
	}

	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case tsk := <-q.dispatchChannel:
		if tsk != nil {
			q.runTask(ctx, tsk)
			q.completionChannel <- tsk
		}
	}

	return nil
}

func (q *Queue) runTask(ctx context.Context, tsk *task.Task) {
	ctx, lgr := log.ContextAndLoggerWithField(ctx, "task", tsk.LogFields())

	runner, ok := q.runners[tsk.Type]
	if !ok {
		// A whole task type is unprocessable by this queue; surface it distinctly (a configuration problem) rather than
		// letting it blend in with ordinary per-task failures downstream.
		lgr.Error("Runner not found for task type; task cannot be processed")
		tsk.SetFailedWithError(errors.New("runner not found for task type"))
		RunnerNotFoundTotal.WithLabelValues(tsk.Type).Inc()
		return
	}

	// The claim context is canceled with ErrClaimLost by the task claim monitor when the task is deleted or re-claimed
	// mid-run. Claim loss is irreversible for this run (every claim gets a fresh token), so once canceled the outcome
	// can never be persisted.
	claimContext, claimCancel := context.WithCancelCause(ctx)
	defer claimCancel(nil)

	// Clearing the claim token marks the outcome as unpersistable, which completeTask discards. Done in a defer, after
	// the recover below, so a panicking runner is also reconciled correctly.
	defer func() {
		if errors.Is(context.Cause(claimContext), ErrClaimLost) {
			tsk.ClaimToken = nil
		}
	}()

	defer func() {
		if err := recover(); err != nil {
			lgr.WithFields(log.Fields{"error": err, "stack": string(debug.Stack())}).Error("Unhandled panic while running task")
			tsk.AppendError(errors.New("unhandled panic"))
			RunPanicTotal.WithLabelValues(tsk.Type).Inc()
		}
	}()

	// If runner does not respect its own maximum duration, then enforce a context-based timeout. This forces the task
	// to cancel via the context.
	runnerContext, cancel := context.WithTimeoutCause(claimContext, runner.GetRunnerTimeout(), ErrRunnerTimeoutExceeded)
	defer cancel()

	// Watchdog for a runner that ignores cancellation. Go cannot preempt a goroutine, so if the runner blows past its
	// timeout without returning, this worker stays blocked until the process restarts (the task itself is recovered by
	// the deadline/unstick mechanism). We cannot unblock the worker, but we surface the condition via a log and metric
	// so a non-cooperative runner is detectable rather than silent. The grace period lets a cooperative runner that
	// returns promptly after the timeout cancellation avoid being reported as blocked here; such a run is instead
	// logged and counted as "recovered" during reconciliation.
	runnerWatchdog := time.AfterFunc(runner.GetRunnerTimeout()+q.config.RunnerWatchdogGracePeriod, func() {
		lgr.Error("Task runner exceeded timeout without returning; worker is blocked until it returns")
		RunnerTimeoutExceededTotal.WithLabelValues(runner.GetRunnerType(), "blocked").Inc()
	})
	defer runnerWatchdog.Stop()

	// Watch the claim for the duration of the run. The task is owned by the runner while it runs, so the watch gets its
	// own copies of everything it needs rather than reading tsk.
	go func(id string, typ string, claimToken string) {
		if reason := q.monitorTaskForLostClaim(claimContext, id, claimToken); reason != nil {
			log.LoggerFromContext(claimContext).Warnf("Task %s; canceling task run", *reason)
			RunClaimLostTotal.WithLabelValues(typ, *reason).Inc()
			claimCancel(ErrClaimLost)
		}
	}(tsk.ID, tsk.Type, *tsk.ClaimToken)

	// Run the task via the runner
	startTime := time.Now()
	runner.Run(runnerContext, tsk)
	duration := time.Since(startTime).Truncate(time.Millisecond)

	// Immediately stop the runner watchdog
	runnerWatchdog.Stop()

	RunDurationSeconds.WithLabelValues(runner.GetRunnerType()).Observe(duration.Seconds())
	if duration > runner.GetRunnerDurationMaximum() {
		lgr.WithField("duration", duration.Seconds()).Warn("Task duration exceeds maximum")
	}

	// The claim was lost mid-run; any write-back would miss the claim token, so skip state reconciliation. The outcome
	// is discarded during completion.
	if errors.Is(context.Cause(claimContext), ErrClaimLost) {
		return
	}

	// If the runner left the task running, reconcile its state based on why the run ended. A claim lost after the check
	// above matches neither branch; the outcome is discarded during completion. Any other run left running is failed
	// with a missing terminal state error during completion.
	if tsk.State == task.TaskStateRunning {
		if context.Cause(ctx) != nil {
			// The parent (worker) context was canceled by shutdown; make the task available again for retry rather than
			// treating the interruption as a completion.
			tsk.RepeatAvailableAfter(0)
		} else if cause := context.Cause(runnerContext); errors.Is(cause, ErrRunnerTimeoutExceeded) {
			// The runner exceeded its timeout but returned; record the cause so the failure is attributed to the
			// timeout rather than the generic missing terminal state error.
			lgr.Warn("Task runner exceeded timeout; task will be failed")
			tsk.AppendError(cause)
			RunnerTimeoutExceededTotal.WithLabelValues(runner.GetRunnerType(), "recovered").Inc()
		}
	}
}

// monitorTaskForLostClaim polls until the run's claim on the task is lost - the task is gone (deleted) or holds a
// different claim token (unstuck and possibly re-claimed) - and then returns the reason for the claim loss. If the
// context is canceled before the claim is lost, returns nil.
func (q *Queue) monitorTaskForLostClaim(ctx context.Context, id string, runningClaimToken string) *string {
	ticker := time.NewTicker(q.config.MonitorTaskDelay)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if storeClaimToken, exists, err := q.repository.GetTaskClaimToken(ctx, id); err != nil {
				if context.Cause(ctx) == nil {
					log.LoggerFromContext(ctx).WithError(err).Error("Unable to get task claim token")
				}
			} else if !exists {
				return pointer.From("deleted")
			} else if storeClaimToken == nil || *storeClaimToken != runningClaimToken {
				return pointer.From("reclaimed")
			}
		}
	}
}

func (q *Queue) startManager(ctx context.Context) {
	q.managerWaitGroup.Go(func() {
		lgr := log.LoggerFromContext(ctx)

		lgr.Debug("Task queue manager started")

		lgr.Debug("Task queue start manager delay initiated")

		// Start at a random future time to help prevent thundering herd problem
		select {
		case <-ctx.Done():
			lgr.WithError(context.Cause(ctx)).Debug("Task queue manager stopped before dispatching tasks")
			return
		case <-time.After(randomDuration(q.config.StartManagerDelay)):
			lgr.Debug("Task queue start manager delay complete")
		}

		// Start at a random future time to help prevent thundering herd problem
		unstickTasksTime := time.Now().Add(randomDuration(q.config.UnstickTasksDelay))

		for {
			if err := q.executeManager(ctx); err != nil {
				lgr.WithError(err).Debug("Task queue manager stopping")

				// Complete any remaining tasks concurrently so the total drain time is bounded by a single slow
				// completion write rather than the sum across all workers' tasks, keeping shutdown within the stop wait
				// timeout.
				var completionWaitGroup sync.WaitGroup
				for tsk := range q.completionChannel {
					if tsk != nil {
						completionWaitGroup.Go(func() { q.completeTask(ctx, tsk) })
					}
				}
				completionWaitGroup.Wait()

				lgr.WithError(err).Debug("Task queue manager stopped")
				return
			}

			if unstickTasksTime.Before(time.Now()) {
				q.unstickTasks(ctx)
				unstickTasksTime = time.Now().Add(durationWithJitter(q.config.UnstickTasksDelay))
			}
		}
	})
}

func (q *Queue) executeManager(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	default:
	}

	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case tsk := <-q.completionChannel:
		if tsk != nil {
			q.completeTask(ctx, tsk)
			q.workersAvailable++
			WorkersAvailable.WithLabelValues(q.name).Set(float64(q.workersAvailable))
			q.dispatchTasks(ctx)
		}
	case <-time.After(durationWithJitter(q.config.DispatchTasksDelay)):
		q.dispatchTasks(ctx)
	}

	return nil
}

func (q *Queue) unstickTasks(ctx context.Context) {
	// Delay availability of unstuck tasks to ensure that any still-running tasks have likely been canceled and exited
	// before being re-dispatched.
	availabilityDelay := q.config.MonitorTaskDelay + q.config.RunnerWatchdogGracePeriod + q.config.UnstickTasksAvailableGracePeriod
	ids, err := q.repository.UnstickTasks(ctx, availabilityDelay)
	if count := len(ids); count > 0 {
		log.LoggerFromContext(ctx).WithFields(log.Fields{"count": count, "ids": ids}).Info("Unstuck tasks")
	}

	// Log error unless context was canceled
	if err != nil {
		if context.Cause(ctx) == nil {
			log.LoggerFromContext(ctx).WithError(err).Error("Unable to unstick tasks")
		}
	}
}

func (q *Queue) dispatchTasks(ctx context.Context) {
	if q.workersAvailable < 1 {
		return
	}

	lgr := log.LoggerFromContext(ctx)

	// Iterate across all pending tasks
	cursor, err := q.repository.IteratePending(ctx)
	if err != nil {
		if context.Cause(ctx) == nil {
			lgr.WithError(err).Error("Unable to open task iterator")
		}
		return
	}
	defer storeStructuredMongo.CloseCursor(ctx, cursor)

	// Loop until no more workers available or no more pending tasks
	for q.workersAvailable > 0 && cursor.Next(ctx) {
		tsk := &task.Task{}
		if err = cursor.Decode(tsk); err != nil {
			lgr.WithError(err).Error("Unable to decode task")
		} else if err = q.dispatchTask(ctx, tsk); err != nil {
			lgr.WithError(err).Error("Unable to dispatch task")
			return
		}
	}

	// Log cursor error unless context was canceled
	if err := cursor.Err(); err != nil {
		if context.Cause(ctx) == nil {
			lgr.WithError(err).Error("Unable to iterate tasks")
		}
	}
}

func (q *Queue) dispatchTask(ctx context.Context, tsk *task.Task) error {
	ctx, lgr := log.ContextAndLoggerWithField(ctx, "task", tsk.LogFields())

	// we don't error here if missing, as the task will be failed during runTask and the error persisted to the database
	// when the task completes
	var deadline time.Duration
	if runner, ok := q.runners[tsk.Type]; ok {
		deadline = runner.GetRunnerDeadline()
	} else {
		deadline = TaskDeadlineDefault
	}

	// StartTask completes regardless of context cancellation, so its outcome is definitive: a non-nil startedTask means
	// the claim committed with a known claim token.
	startedTask, err := q.repository.StartTask(ctx, tsk.ID, tsk.Revision, deadline)
	if err != nil {
		return errors.Wrap(err, "unable to start task")
	} else if startedTask == nil {
		lgr.Info("Task no longer available to start")
		return nil
	}

	// Hand the task off to a worker. If the queue is shutting down before a worker can receive it, revert the task to
	// pending rather than blocking the manager. The revert uses the started task claim token so it reliably matches.
	select {
	case <-ctx.Done():
		if err := q.repository.StopTask(ctx, startedTask.ID, startedTask.Revision, startedTask.ClaimToken, task.TaskStatePending, nil, nil); err != nil {
			return errors.Wrap(err, "unable to revert task to pending")
		}
	case q.dispatchChannel <- startedTask:
		q.workersAvailable--
		WorkersAvailable.WithLabelValues(q.name).Set(float64(q.workersAvailable))
	}

	return nil
}

// completeTask persists the task's completion. It deliberately does not touch workersAvailable (the caller accounts for
// the freed worker where relevant), so it is safe to call concurrently during the manager's shutdown drain.
func (q *Queue) completeTask(ctx context.Context, tsk *task.Task) {
	ctx, lgr := log.ContextAndLoggerWithField(ctx, "task", tsk.LogFields())

	// The claim was lost mid-run (task deleted or re-claimed): any write-back would miss the claim token, so discard
	// the outcome rather than logging a spurious lost completion.
	if tsk.ClaimToken == nil {
		lgr.Warn("Task claim lost; task run outcome discarded")
		return
	}

	var duration *time.Duration
	if tsk.RunTime != nil {
		// Clamp to a zero minimum: the run time round-trips through the database without a monotonic reading, so a
		// backwards wall clock step could yield a negative elapsed time, which StopTask would reject, losing the
		// completion.
		duration = pointer.From(max(time.Since(*tsk.RunTime), 0))
	}

	q.computeState(ctx, tsk)

	// computeState has already settled the terminal state. A failed task is a genuine problem worth an error; a task
	// that errored but will run again (e.g. reverted to pending for retry) is an expected, recoverable outcome, so log
	// it at warning to avoid error-level noise from routine retries.
	if err := tsk.GetError(); tsk.State == task.TaskStateFailed {
		lgr.WithError(err).Error("Task failed while running")
	} else if err != nil {
		lgr.WithError(err).Warn("Error occurred while running task that did not fail")
	}

	// Data and Error use non-nil wrappers so that a task whose data or error was cleared during the run has the
	// corresponding field unset in the database, rather than left stale (a nil wrapper means "leave unchanged" to
	// StopTask).
	update := &task.TaskUpdate{
		Data:          pointer.From(tsk.Data),
		AvailableTime: tsk.AvailableTime,
		Error:         &errors.Serializable{Error: tsk.GetError()},
	}
	if err := q.repository.StopTask(ctx, tsk.ID, tsk.Revision, tsk.ClaimToken, tsk.State, duration, update); err != nil {
		lgr.WithError(err).Error("Unable to complete task")
	}
}

func (q *Queue) computeState(ctx context.Context, tsk *task.Task) {
	switch tsk.State {
	case task.TaskStatePending:
		now := time.Now().UTC()
		if tsk.AvailableTime == nil {
			log.LoggerFromContext(ctx).Warn("Available time missing for pending task")
			tsk.AvailableTime = pointer.FromTime(now)
		} else if tsk.AvailableTime.Before(now) {
			if tsk.AvailableTime.Before(now.Add(-time.Minute)) { // Allow some leeway to prevent spurious warnings
				log.LoggerFromContext(ctx).Warn("Available time significantly before now for pending task")
			}
			tsk.AvailableTime = pointer.FromTime(now)
		}
	case task.TaskStateRunning:
		// The runner returned without moving the task out of the running state, violating the runner contract; fail the
		// task rather than guessing whether it succeeded.
		if tsk.HasError() {
			tsk.SetFailed()
		} else {
			tsk.SetFailedWithError(errors.New("runner failed to set state"))
		}
	case task.TaskStateFailed, task.TaskStateCompleted:
		tsk.AvailableTime = nil
	default:
		tsk.SetFailedWithError(errors.New("unknown task state"))
	}
}

// validateRunner enforces the runner duration contract, deadline > timeout > duration maximum > 0, so that a
// misconfigured runner fails queue construction rather than causing subtle runtime misbehavior (a non-positive deadline
// prevents the task from ever starting, while a deadline that does not exceed the timeout allows the unstick mechanism
// to reset a task that is still running).
func validateRunner(runner Runner) error {
	if durationMaximum := runner.GetRunnerDurationMaximum(); durationMaximum <= 0 {
		return errors.New("runner duration maximum is invalid")
	} else if timeout := runner.GetRunnerTimeout(); timeout <= durationMaximum {
		return errors.New("runner timeout is invalid")
	} else if runner.GetRunnerDeadline() <= timeout {
		return errors.New("runner deadline is invalid")
	} else {
		return nil
	}
}

// waitWithTimeout waits for the wait group to complete, returning true if it completed within the timeout and false
// otherwise. On timeout the internal waiter goroutine is left running until the wait group eventually completes (or the
// process exits).
func waitWithTimeout(waitGroup *sync.WaitGroup, timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		waitGroup.Wait()
		close(done)
	}()

	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}

func randomDuration(duration time.Duration) time.Duration {
	if duration <= 0 {
		return 0
	}
	return time.Duration(crypto.RandomInt64N(int64(duration)))
}

func durationWithJitter(duration time.Duration) time.Duration {
	if duration <= 0 {
		return 0
	}
	jitter := time.Duration(float64(duration) * DurationJitterFactor)
	return duration + (randomDuration(jitter*2) - jitter)
}

var (
	// WorkersTotal reports the configured number of workers, per queue.
	WorkersTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tidepool_task_workers_total",
		Help: "The configured number of task queue workers, sorted by queue",
	}, []string{"queue"})

	// WorkersAvailable reports the number of available workers, per queue. A value pinned at zero indicates a saturated
	// queue or workers wedged in non-cooperative runners.
	WorkersAvailable = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tidepool_task_workers_available",
		Help: "The number of available task queue workers, sorted by queue",
	}, []string{"queue"})

	// RunnerNotFoundTotal counts task runs for which no runner is registered for the task's type, sorted by type. A
	// non-zero value indicates pending tasks of a type this queue cannot process (a configuration problem); such tasks
	// are failed immediately.
	RunnerNotFoundTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tidepool_task_runner_not_found_total",
		Help: "The total number of task runs with no registered runner for the task type, sorted by type",
	}, []string{"type"})

	// RunDurationSeconds observes how long each task run took, sorted by type. Use it to track run latency percentiles
	// and to alert when durations approach the runner timeout.
	RunDurationSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tidepool_task_run_duration_seconds",
		Help:    "The duration of task runs, in seconds, sorted by type",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15),
	}, []string{"type"})

	// RunnerTimeoutExceededTotal counts task runs that exceeded the runner timeout, sorted by type and disposition,
	// where disposition can be:
	//   - "blocked" - for runs the watchdog caught still running past the grace period (the worker is wedged until the
	//     runner eventually returns or the process restarts, so this is the severe case)
	//   - "recovered" - for runs that exceeded the timeout, but returned once their context was canceled (the timeout
	//     mechanism worked as designed)
	RunnerTimeoutExceededTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tidepool_task_runner_timeout_exceeded_total",
		Help: "The total number of task runs that exceeded the runner timeout, sorted by type and disposition",
	}, []string{"type", "disposition"})

	// RunClaimLostTotal counts task runs canceled because the run's claim was lost, sorted by type and reason, where
	// reason can be:
	//   - "deleted" - the task document was deleted mid-run
	//   - "reclaimed" - the stored claim token no longer matches (the task was unstuck and possibly re-claimed
	//     elsewhere)
	RunClaimLostTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tidepool_task_run_claim_lost_total",
		Help: "The total number of task runs canceled because the task claim was lost, sorted by type and reason",
	}, []string{"type", "reason"})

	// RunPanicTotal counts task runs that panicked and were recovered, sorted by type. The task is then failed, unless
	// the runner had already moved it out of running before panicking.
	RunPanicTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tidepool_task_run_panic_total",
		Help: "The total number of task runs that panicked, sorted by type",
	}, []string{"type"})
)
