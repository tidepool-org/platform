package queue

import (
	"context"
	"math/rand"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

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

	if workersString, found := configReporter.Get("workers"); found {
		workers, err := strconv.ParseInt(workersString, 10, 0)
		if err != nil {
			return errors.New("workers is invalid")
		}
		c.Workers = int(workers)
	}
	if delayString, found := configReporter.Get("delay"); found {
		delay, err := strconv.ParseInt(delayString, 10, 0)
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
	CanRunTask(tks *task.Task) bool

	Run(ctx context.Context, tsk *task.Task)
}

type Queue struct {
	logger            log.Logger
	store             store.Store
	workers           int
	delay             time.Duration
	runners           []Runner
	cancelFunc        context.CancelFunc
	waitGroup         sync.WaitGroup
	workersAvailable  int
	dispatchChannel   chan *task.Task
	completionChannel chan *task.Task
	timer             *time.Timer
	session           store.TaskSession
	iterator          store.TaskIterator
}

func New(cfg *Config, lgr log.Logger, str store.Store) (*Queue, error) {
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

	return &Queue{
		logger:            lgr,
		store:             str,
		workers:           workers,
		delay:             delay,
		runners:           []Runner{},
		dispatchChannel:   make(chan *task.Task, workers),
		completionChannel: make(chan *task.Task, workers),
	}, nil
}

func (q *Queue) RegisterRunner(runner Runner) error {
	if runner == nil {
		return errors.New("runner is missing")
	}

	q.runners = append(q.runners, runner)
	return nil
}

func (q *Queue) Start() {
	if q.cancelFunc == nil {
		ctx, cancelFunc := context.WithCancel(log.NewContextWithLogger(context.Background(), q.logger))
		q.cancelFunc = cancelFunc

		q.startWorkers(ctx)
		q.startManager(ctx)
	}
}

func (q *Queue) Stop() {
	if q.cancelFunc != nil {
		q.cancelFunc()
		q.cancelFunc = nil

		q.waitGroup.Wait()
	}
}

func (q *Queue) startWorkers(ctx context.Context) {
	for q.workersAvailable = 0; q.workersAvailable < q.workers; q.workersAvailable++ {
		q.startWorker(ctx)
	}
}

func (q *Queue) startWorker(ctx context.Context) {
	q.waitGroup.Add(1)
	go func() {
		defer q.waitGroup.Done()

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

func (q *Queue) runTask(ctx context.Context, tsk *task.Task) {
	logger := q.logger.WithField("taskId", tsk.ID)

	defer func() {
		if err := recover(); err != nil {
			logger.WithFields(log.Fields{"error": err, "stack": string(debug.Stack())}).Error("Unhandled panic")
			tsk.AppendError(errors.New("unhandled panic"))
		}
	}()

	for _, runner := range q.runners {
		if runner.CanRunTask(tsk) {
			runner.Run(ctx, tsk)
			return
		}
	}

	logger.Error("Runner not found for task")
	tsk.AppendError(errors.New("runner not found for task"))
}

func (q *Queue) startManager(ctx context.Context) {
	q.waitGroup.Add(1)
	go func() {
		defer q.waitGroup.Done()

		q.startTimer(time.Duration(rand.Int63n(int64(q.delay))))
		defer q.stopTimer()

		for {
			select {
			case <-ctx.Done():
				return
			case tsk := <-q.completionChannel:
				q.stopTimer()
				q.completeTask(ctx, tsk)
				q.startTimer(q.dispatchTasks(ctx))
			case <-q.timer.C:
				q.startTimer(q.dispatchTasks(ctx))
			}
		}
	}()
}

func (q *Queue) dispatchTasks(ctx context.Context) time.Duration {
	defer q.stopPendingIterator()
	for q.workersAvailable > 0 {
		iter := q.startPendingIterator(ctx)

		tsk := &task.Task{}
		if iter.Next(tsk) {
			q.dispatchTask(ctx, tsk)
		} else if err := iter.Error(); err != nil {
			q.logger.WithError(err).Error("Failure iterating tasks") // TODO: Only warn after n fallbacks
			return q.delay                                           // TODO: Exponential fallback
		} else {
			return q.delay
		}
	}

	return q.delay
}

func (q *Queue) dispatchTask(ctx context.Context, tsk *task.Task) {
	logger := q.logger.WithField("taskId", tsk.ID)

	ssn := q.store.NewTaskSession()
	defer ssn.Close()

	tsk.State = task.TaskStateRunning
	tsk.RunTime = pointer.Time(time.Now())

	var err error
	tsk, err = ssn.UpdateFromState(ctx, tsk, task.TaskStatePending)
	if err != nil {
		logger.WithError(err).Error("Failure to update state during dispatch task")
		return
	}

	q.workersAvailable--
	q.dispatchChannel <- tsk
}

func (q *Queue) completeTask(ctx context.Context, tsk *task.Task) {
	logger := q.logger.WithField("taskId", tsk.ID)

	q.workersAvailable++

	ssn := q.store.NewTaskSession()
	defer ssn.Close()

	if tsk.RunTime != nil {
		tsk.Duration = pointer.Float64(float64(time.Now().Sub(*tsk.RunTime).Truncate(time.Millisecond)) / float64(time.Second))
	}
	q.computeState(tsk)

	_, err := ssn.UpdateFromState(ctx, tsk, task.TaskStateRunning)
	if err != nil {
		logger.WithError(err).Error("Failure to update state during complete task")
	}
}

func (q *Queue) computeState(tsk *task.Task) {
	switch tsk.State {
	case task.TaskStatePending:
		if tsk.AvailableTime == nil || time.Now().After(*tsk.AvailableTime) {
			tsk.AppendError(errors.New("pending task requires future available time"))
			tsk.State = task.TaskStateFailed
		}
	case task.TaskStateRunning:
		if tsk.HasError() {
			tsk.State = task.TaskStateFailed
		} else {
			tsk.State = task.TaskStateCompleted
		}
	case task.TaskStateFailed, task.TaskStateCompleted:
	default:
		tsk.AppendError(errors.New("unknown state"))
		tsk.State = task.TaskStateFailed
	}
}

func (q *Queue) startTimer(delay time.Duration) {
	if delay > 0 {
		if q.timer == nil {
			q.timer = time.NewTimer(delay)
		} else {
			q.timer.Reset(delay)
		}
	}
}

func (q *Queue) stopTimer() {
	if q.timer != nil {
		if !q.timer.Stop() {
			<-q.timer.C
		}
	}
}

func (q *Queue) startPendingIterator(ctx context.Context) store.TaskIterator {
	if q.session == nil {
		q.session = q.store.NewTaskSession()
	}
	if q.iterator == nil {
		q.iterator = q.session.IteratePending(ctx)
	}
	return q.iterator
}

func (q *Queue) stopPendingIterator() {
	if q.iterator != nil {
		q.iterator.Close()
		q.iterator = nil
	}
	if q.session != nil {
		q.session.Close()
		q.session = nil
	}
}
