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

func New(cfg *Config, lgr log.Logger, str store.Store) (Queue, error) {
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

	workers := cfg.Workers
	delay := cfg.Delay

	return &queue{
		logger:            lgr,
		store:             str,
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

		for {
			select {
			case <-ctx.Done():
				return
			case tsk := <-q.dispatchChannel:
				q.runTask(ctx, tsk)
				q.completionChannel <- tsk
			}
		}
	}()
}

func (q *queue) runTask(ctx context.Context, tsk *task.Task) {
	ctx = log.ContextWithField(ctx, "taskId", tsk.ID)

	defer func() {
		if err := recover(); err != nil {
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
		runner.Run(ctx, tsk)

		if taskDuration := time.Since(startTime); taskDuration > runner.GetRunnerDurationMaximum() {
			log.LoggerFromContext(ctx).WithField("taskDuration", taskDuration.Truncate(time.Millisecond).Seconds()).Warn("Task duration exceeds maximum")
		}
	} else {
		tsk.AppendError(errors.New("runner not found for task"))
		tsk.SetFailed()
	}
}

func (q *queue) startManager(ctx context.Context) {
	q.managerWaitGroup.Add(1)

	go func() {
		defer q.managerWaitGroup.Done()

		q.startTimer(time.Duration(rand.Int63n(int64(q.delay)) + 1))
		defer q.stopTimer()

		// pick a starting random time in a future cycle to ensure multiple daemons don't do this exactly at the same
		// time, it is not an error condition if it does, but could stress the db if the collection gets large
		nextUnstickTime := time.Now().Add(time.Duration(rand.Int63n(int64(q.delay * 15))))

		for {
			if nextUnstickTime.Before(time.Now()) {
				q.unstickTasks(ctx)
				nextUnstickTime = time.Now().Add(q.delay * 15)
			}

			select {
			case <-ctx.Done(): // Drain and complete any interrupted tasks
				for tsk := range q.completionChannel {
					q.completeTask(ctx, tsk)
				}
				return
			case tsk := <-q.completionChannel:
				if tsk != nil {
					q.stopTimer()
					q.completeTask(ctx, tsk)
					q.startTimer(q.dispatchTasks(ctx))
				}
			case <-q.timer.C:
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

func (q *queue) dispatchTasks(ctx context.Context) time.Duration {
	defer q.stopPendingIterator(ctx)
	for q.workersAvailable > 0 {
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
			return q.delay
		}
	}

	return q.delay
}

func (q *queue) dispatchTask(ctx context.Context, tsk *task.Task) {
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

	q.workersAvailable--
	q.dispatchChannel <- tsk
}

func (q *queue) completeTask(ctx context.Context, tsk *task.Task) {
	ctx = log.ContextWithField(ctx, "taskId", tsk.ID)

	q.workersAvailable++

	repository := q.store.NewTaskRepository()

	q.computeState(tsk)

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
		log.LoggerFromContext(ctx).WithError(err).Error("Failure to update state during complete task")
	}

	if tsk.HasError() {
		log.LoggerFromContext(ctx).WithError(tsk.Error.Error).Error("Error occurred while running task")
	}
}

func (q *queue) computeState(tsk *task.Task) {
	switch tsk.State {
	case task.TaskStatePending:
		now := time.Now()
		if tsk.AvailableTime == nil || tsk.AvailableTime.Before(now) {
			tsk.AvailableTime = &now
		} else if time.Now().After(*tsk.AvailableTime) {
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
