package queue

import (
	"context"
	"math/rand"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/task/store"
)

type Config struct {
	Workers int
	Delay   time.Duration
}

func NewConfig() *Config {
	return &Config{
		Workers: 1,
		Delay:   60 * time.Second,
	}
}

func (c *Config) Load(configReporter config.Reporter) error {
	if configReporter == nil {
		return errors.New("config reporter is missing")
	}

	if workersString, err := configReporter.Get("workers"); err == nil {
		var workers int64
		workers, err = strconv.ParseInt(workersString, 10, 0)
		if err != nil {
			return errors.New("workers is invalid")
		}
		c.Workers = int(workers)
	}
	if delayString, err := configReporter.Get("delay"); err == nil {
		var delay int64
		delay, err = strconv.ParseInt(delayString, 10, 0)
		if err != nil {
			return errors.New("delay is invalid")
		}
		c.Delay = time.Duration(delay) * time.Second
	}

	return nil
}

func (c *Config) Validate() error {
	if c.Workers < 1 {
		return errors.New("workers is invalid")
	}
	if c.Delay < 0 {
		return errors.New("delay is invalid")
	}

	return nil
}

type Runner interface {

	// The type of tasks that the runner supports.
	GetRunnerType() string

	// The time after which the task manager will forcefully reset the task back to pending
	// and available. This is calculated based upon the current time and a duration significantly
	// longer that the task duration maximum. Normally this would only be used on a task that
	// is in the running state even though it is not running (likely due to a system crash or interruption).
	GetRunnerDeadline() time.Time

	// The duration of a task where the task manager will forcefully cancel the task context to interrupt
	// the task and force completion. This is typically a duration somewhat longer than the task
	// duration maximum.
	GetRunnerTimeout() time.Duration

	// The typical duration maximum of the task after which a warning will be displayed.
	GetRunnerDurationMaximum() time.Duration

	// Execute the specified task within the specified context. The context will be forcefully
	// canceled after a duration specified by GetRunnerTimeout.
	Run(ctx context.Context, tsk *task.Task)
}

type Queue interface {
	RegisterRunner(Runner) error
	Start()
	Stop()
}

type queue struct {
	logger            log.Logger
	store             store.Store
	typeFilter        string
	workers           int
	delay             time.Duration
	runners           map[string]Runner
	workersCancelFunc context.CancelFunc
	workersWaitGroup  sync.WaitGroup
	managerCancelFunc context.CancelFunc
	managerWaitGroup  sync.WaitGroup
	workersAvailable  int
	dispatchChannel   chan *task.Task
	completionChannel chan *task.Task
	timer             *time.Timer
	taskRepository    store.TaskRepository
	iterator          *mongo.Cursor
}

func New(cfg *Config, lgr log.Logger, str store.Store, typeFilter string) (Queue, error) {
	if cfg == nil {
		return nil, errors.New("config is missing")
	}
	if lgr == nil {
		return nil, errors.New("logger is missing")
	}
	if str == nil {
		return nil, errors.New("store is missing")
	}
	if typeFilter == "" {
		return nil, errors.New("type filter is empty")
	}

	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}

	workers := cfg.Workers
	delay := cfg.Delay

	return &queue{
		logger:            lgr,
		store:             str,
		typeFilter:        typeFilter,
		workers:           workers,
		delay:             delay,
		runners:           make(map[string]Runner),
		dispatchChannel:   make(chan *task.Task, workers),
		completionChannel: make(chan *task.Task, workers),
	}, nil
}

func (q *queue) RegisterRunner(runner Runner) error {
	if runner == nil {
		return errors.New("runner is missing")
	}

	q.runners[runner.GetRunnerType()] = runner
	return nil
}

func (q *queue) Start() {
	backgroundCtx := log.NewContextWithLogger(context.Background(), q.logger)
	if q.workersCancelFunc == nil {
		ctx, workersCancelFunc := context.WithCancel(backgroundCtx)
		q.workersCancelFunc = workersCancelFunc

		q.startWorkers(ctx)
	}
	if q.managerCancelFunc == nil {
		ctx, managerCancelFunc := context.WithCancel(backgroundCtx)
		q.managerCancelFunc = managerCancelFunc

		q.startManager(ctx)
	}
}

func (q *queue) Stop() {
	if q.workersCancelFunc != nil {
		q.workersCancelFunc()
		q.workersCancelFunc = nil
	}
	q.workersWaitGroup.Wait()

	close(q.completionChannel)

	if q.managerCancelFunc != nil {
		q.managerCancelFunc()
		q.managerCancelFunc = nil
	}
	q.managerWaitGroup.Wait()

	close(q.dispatchChannel)
}

func (q *queue) startWorkers(ctx context.Context) {
	for q.workersAvailable = 0; q.workersAvailable < q.workers; q.workersAvailable++ {
		q.startWorker(ctx)
	}
}

func (q *queue) startWorker(ctx context.Context) {
	q.workersWaitGroup.Add(1)
	go func() {
		defer q.workersWaitGroup.Done()

		debugLogger := q.debugLogger()

		for {
			select {
			case <-ctx.Done():
				debugLogger.Warn("Worker context done, exiting")
				return
			case tsk := <-q.dispatchChannel:
				debugLogger.WithField("task", tsk.Fields()).Warn("Running task")
				q.runTask(ctx, tsk)
				taskFields := tsk.Fields()
				debugLogger.WithField("task", taskFields).Warn("Ran task; sending to completion channel")
				q.completionChannel <- tsk
				debugLogger.WithField("task", taskFields).Warn("Sent to completion channel")
			}
		}
	}()
}

func (q *queue) runTask(ctx context.Context, tsk *task.Task) {
	ctx = log.ContextWithField(ctx, "taskId", tsk.ID)

	debugLogger := q.debugLogger()

	defer func() {
		if err := recover(); err != nil {
			debugLogger.WithField("task", tsk.Fields()).Error("Unhandled panic")
			log.LoggerFromContext(ctx).WithFields(log.Fields{"error": err, "stack": string(debug.Stack())}).Error("Unhandled panic")
			tsk.AppendError(errors.New("unhandled panic"))
		}
	}()

	if runner, ok := q.runners[tsk.Type]; ok {

		// If runner does not respect its own maximum duration, then enforce a context-based timeout.
		// This forces the task to cancel via the context.
		ctx, cancel := context.WithTimeout(ctx, runner.GetRunnerTimeout())
		defer cancel()

		startTime := time.Now()

		// Run the task via the runner
		debugLogger.WithField("task", tsk.Fields()).Error("Actual running task")
		runner.Run(ctx, tsk)
		debugLogger.WithField("task", tsk.Fields()).Error("Actual ran task")

		if taskDuration := time.Since(startTime); taskDuration > runner.GetRunnerDurationMaximum() {
			log.LoggerFromContext(ctx).WithField("taskDuration", taskDuration.Truncate(time.Millisecond).Seconds()).Warn("Task duration exceeds maximum")
		}
	} else {
		debugLogger.WithField("task", tsk.Fields()).Error("Runner not found for task")
		tsk.AppendError(errors.New("runner not found for task"))
		tsk.SetFailed()
	}
}

func (q *queue) startManager(ctx context.Context) {
	q.managerWaitGroup.Add(1)

	go func() {
		defer q.managerWaitGroup.Done()

		debugLogger := q.debugLogger()

		q.startTimer(time.Duration(rand.Int63n(int64(q.delay)) + 1))
		defer q.stopTimer()

		// pick a starting random time in a future cycle to ensure multiple daemons don't do this exactly at the same
		// time, it is not an error condition if it does, but could stress the db if the collection gets large
		nextUnstickTime := time.Now().Add(time.Duration(rand.Int63n(int64(q.delay * 15))))

		for {
			if nextUnstickTime.Before(time.Now()) {
				debugLogger.Warn("Unsticking tasks")
				q.unstickTasks(ctx)
				debugLogger.Warn("Unstuck tasks")
				nextUnstickTime = time.Now().Add(q.delay * 15)
			}

			select {
			case <-ctx.Done(): // Drain and complete any interrupted tasks
				debugLogger.Warn("Manager context cancelled")
				for tsk := range q.completionChannel {
					q.completeTask(ctx, tsk)
				}
				return
			case tsk := <-q.completionChannel:
				if tsk != nil {
					debugLogger.WithField("task", tsk.Fields()).Warn("Task received from completion channel")
					q.stopTimer()
					q.completeTask(ctx, tsk)
					q.startTimer(q.dispatchTasks(ctx))
				} else {
					debugLogger.Warn("Null task received from completion channel")
				}
			case <-q.timer.C:
				debugLogger.Warn("Timer fired")
				q.startTimer(q.dispatchTasks(ctx))
			}
		}
	}()
}

func (q *queue) unstickTasks(ctx context.Context) {
	repository := q.store.NewTaskRepository()
	count, err := repository.UnstickTasks(ctx)
	if err != nil {
		q.logger.WithError(err).Error("Failure in unsticking tasks")
	}
	if count > 0 {
		q.logger.WithField("unstickCount", count).Warn("Unstuck tasks")
	}
}

func (q *queue) debugLogger() log.Logger {
	if q.typeFilter == "org.tidepool.oauth.dexcom.fetch" {
		return q.logger.WithField("typeFilter", q.typeFilter)
	} else {
		return logNull.NewLogger()
	}
}

func (q *queue) dispatchTasks(ctx context.Context) time.Duration {
	defer q.stopPendingIterator(ctx)

	debugLogger := q.debugLogger()

	debugLogger.WithField("workersAvailable", q.workersAvailable).Debug("Dispatching tasks")

	for q.workersAvailable > 0 {

		debugLogger.WithField("workersAvailable", q.workersAvailable).Debug("Start pending iterator")
		iter, err := q.startPendingIterator(ctx)
		if err != nil {
			q.logger.WithError(err).Error("Failure starting pending iterator")
			return q.delay
		}

		tsk := &task.Task{}
		if iter.Next(ctx) {
			if err := iter.Decode(tsk); err != nil {
				q.logger.WithError(err).Error("Failure iterating tasks")
				return q.delay
			}
			q.dispatchTask(ctx, tsk)
		} else {
			debugLogger.Debug("Dispatched tasks; no tasks")
			return q.delay
		}
	}

	debugLogger.Debug("Dispatched tasks; no workers")
	return q.delay
}

func (q *queue) dispatchTask(ctx context.Context, tsk *task.Task) {
	debugLogger := q.debugLogger()

	debugLogger.Debug("Dispatching task")

	ctx = log.ContextWithField(ctx, "taskId", tsk.ID)

	repository := q.store.NewTaskRepository()

	tsk.State = task.TaskStateRunning
	tsk.AvailableTime = nil
	tsk.RunTime = pointer.FromAny(time.Now())

	// we don't error here if missing, as the task will be failed during runTask
	if runner, ok := q.runners[tsk.Type]; ok {
		tsk.DeadlineTime = pointer.FromAny(runner.GetRunnerDeadline())
	}

	var err error
	tsk, err = repository.UpdateFromState(context.WithoutCancel(ctx), tsk, task.TaskStatePending)
	if err != nil {
		if errors.Is(err, task.AlreadyClaimedTask) {
			log.LoggerFromContext(ctx).Warnf("Failure to claim task %s (%s) as it is already in progress or is no longer available.", tsk.Name, tsk.ID)
			return
		}

		log.LoggerFromContext(ctx).WithError(err).Error("Failure to update state during dispatch task")
		return
	}

	taskFields := tsk.Fields()

	q.workersAvailable--
	debugLogger.WithField("task", taskFields).WithField("workersAvailable", q.workersAvailable).Debug("Decremented workers available")
	q.dispatchChannel <- tsk
	debugLogger.WithField("task", taskFields).Debug("Dispatched task")
}

func (q *queue) completeTask(ctx context.Context, tsk *task.Task) {
	ctx = log.ContextWithField(ctx, "taskId", tsk.ID)

	debugLogger := q.debugLogger()

	debugLogger.WithField("workersAvailable", q.workersAvailable).Debug("Incrementing workers available")
	q.workersAvailable++
	debugLogger.WithField("workersAvailable", q.workersAvailable).Debug("Incremented workers available")

	repository := q.store.NewTaskRepository()

	debugLogger.WithField("task", tsk.Fields()).Debug("Computing state")
	q.computeState(tsk)
	debugLogger.WithField("task", tsk.Fields()).Debug("Computed state")

	if tsk.State != task.TaskStatePending {
		tsk.AvailableTime = nil
	}
	tsk.DeadlineTime = nil
	if tsk.RunTime != nil {
		tsk.Duration = pointer.FromFloat64(time.Since(*tsk.RunTime).Truncate(time.Millisecond).Seconds())
	}

	// Without cancel to ensure task is updated in the database
	_, err := repository.UpdateFromState(context.WithoutCancel(ctx), tsk, task.TaskStateRunning)
	if err != nil {
		debugLogger.Error("Failure to update state during complete task")
		debugLogger.WithError(err).Error("Failure to update state during complete task")
		log.LoggerFromContext(ctx).WithError(err).Error("Failure to update state during complete task")
	}

	if tsk.HasError() {
		debugLogger.Error("Failure to update state during complete task")
		debugLogger.WithError(tsk.Error.Error).Error("Failure to update state during complete task")
		log.LoggerFromContext(ctx).WithError(tsk.Error.Error).Error("Error occurred while running task")
	}
}

func (q *queue) computeState(tsk *task.Task) {
	switch tsk.State {
	case task.TaskStatePending:
		if tsk.AvailableTime == nil || time.Now().After(*tsk.AvailableTime) {
			tsk.AppendError(errors.New("pending task requires future available time"))
			tsk.SetFailed()
		}
	case task.TaskStateRunning:
		if tsk.HasError() {
			tsk.SetFailed()
		} else {
			tsk.SetCompleted()
		}
	case task.TaskStateFailed, task.TaskStateCompleted:
	default:
		tsk.AppendError(errors.New("unknown state"))
		tsk.SetFailed()
	}
}

func (q *queue) startTimer(delay time.Duration) {
	if delay > 0 {
		if q.timer == nil {
			q.timer = time.NewTimer(delay)
		} else {
			q.timer.Reset(delay)
		}
	}
}

func (q *queue) stopTimer() {
	if q.timer != nil {
		if !q.timer.Stop() {
			<-q.timer.C
		}
	}
}

func (q *queue) startPendingIterator(ctx context.Context) (*mongo.Cursor, error) {
	if q.taskRepository == nil {
		q.taskRepository = q.store.NewTaskRepository()
	}
	if q.iterator == nil {
		if iterator, err := q.taskRepository.IteratePending(ctx); err != nil {
			return nil, err
		} else {
			q.iterator = iterator
		}
	}
	return q.iterator, nil
}

func (q *queue) stopPendingIterator(ctx context.Context) {
	if q.iterator != nil {
		if err := q.iterator.Close(context.WithoutCancel(ctx)); err != nil {
			q.logger.WithError(err).Warn("failure closing pending iterator")
		}
		q.iterator = nil
	}
	if q.taskRepository != nil {
		q.taskRepository = nil
	}
}
