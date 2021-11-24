package summary

import (
	"context"
	"math/rand"

	//"sort"
	"time"

	"github.com/tidepool-org/platform/auth"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/data/types/blood/glucose/summary"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/version"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

const (
	AvailableAfterDurationMaximum = 30 * time.Minute
	AvailableAfterDurationMinimum = 15 * time.Minute
	TaskDurationMaximum           = 40 * time.Minute
	WorkerCount                   = 8
)

type Runner struct {
	logger          log.Logger
	versionReporter version.Reporter
	authClient      auth.Client
	dataClient      dataClient.Client
}

func NewRunner(logger log.Logger, versionReporter version.Reporter, authClient auth.Client, dataClient dataClient.Client) (*Runner, error) {
	if logger == nil {
		return nil, errors.New("logger is missing")
	}
	if versionReporter == nil {
		return nil, errors.New("version reporter is missing")
	}
	if authClient == nil {
		return nil, errors.New("auth client is missing")
	}
	if dataClient == nil {
		return nil, errors.New("data client is missing")
	}

	return &Runner{
		logger:          logger,
		versionReporter: versionReporter,
		authClient:      authClient,
		dataClient:      dataClient,
	}, nil
}

func (r *Runner) CanRunTask(tsk *task.Task) bool {
	return tsk != nil && tsk.Type == Type
}

func (r *Runner) GenerateNextTime() time.Duration {
	randTime := time.Duration(rand.Int63n(int64(AvailableAfterDurationMaximum - AvailableAfterDurationMinimum + 1)))
	return AvailableAfterDurationMinimum + randTime
}

func (r *Runner) Run(ctx context.Context, tsk *task.Task) {
	now := time.Now()

	ctx = log.NewContextWithLogger(ctx, r.logger)

	tsk.ClearError()

	if serverSessionToken, sErr := r.authClient.ServerSessionToken(); sErr != nil {
		tsk.AppendError(errors.Wrap(sErr, "unable to get server session token"))
	} else {
		ctx = auth.NewContextWithServerSessionToken(ctx, serverSessionToken)

		if taskRunner, tErr := NewTaskRunner(r, tsk); tErr != nil {
			tsk.AppendError(errors.Wrap(tErr, "unable to create task runner"))
		} else if tErr = taskRunner.Run(ctx); tErr != nil {
			tsk.AppendError(errors.Wrap(tErr, "unable to run task runner"))
		}
	}

	if !tsk.IsFailed() {
		tsk.RepeatAvailableAfter(r.GenerateNextTime())
	}

	if taskDuration := time.Since(now); taskDuration > TaskDurationMaximum {
		r.logger.WithField("taskDuration", taskDuration.Truncate(time.Millisecond).Seconds()).Warn("Task duration exceeds maximum")
	}
}

type TaskRunner struct {
	*Runner
	task      *task.Task
	context   context.Context
	validator structure.Validator
}

func NewTaskRunner(rnnr *Runner, tsk *task.Task) (*TaskRunner, error) {
	if rnnr == nil {
		return nil, errors.New("runner is missing")
	}
	if tsk == nil {
		return nil, errors.New("task is missing")
	}

	return &TaskRunner{
		Runner: rnnr,
		task:   tsk,
	}, nil
}

func (t *TaskRunner) Run(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is missing")
	}

	if len(t.task.Data) == 0 {
		t.task.SetFailed()
		return errors.New("data is missing")
	}

	t.context = ctx
	t.validator = structureValidator.New()

	t.logger.Info("Searching for User Summaries requiring Update")
	summaries, err := t.dataClient.GetAgedSummaries(t.context, nil)
	if err != nil {
		return err
	}

	t.logger.Info("Starting User Summary Update")

	if err := t.UpdateSummaries(summaries); err != nil {
		t.task.SetFailed()
		return err
	}

	return nil
}

func (t *TaskRunner) UpdateSummaries(summaries []*summary.Summary) error {
	sem := semaphore.NewWeighted(WorkerCount)
	eg, c := errgroup.WithContext(t.context)

	for _, summary := range summaries {
		if c.Err() != nil {
			break
		}

		if err := sem.Acquire(t.context, 1); err != nil {
			t.logger.Error("Failed to acquire semaphore")
			break
		}

		// we can't pass arguments to errgroup goroutines
		// we need to explicitly redefine the variables,
		// because we're launching the goroutines in a loop
		summary := summary
		eg.Go(func() error {
			defer sem.Release(1)
			return t.UpdateUserSummary(summary)
		})
	}
	return eg.Wait()
}

func (t *TaskRunner) UpdateUserSummary(summary *summary.Summary) error {
	t.logger.WithField("UserID", summary.UserID).Debug("Updating User Summary")

	// update summary
	_, err := t.dataClient.UpdateSummary(t.context, summary.UserID)
	if err != nil {
		return err
	}

	t.logger.WithField("UserID", summary.UserID).Debug("Finished Updating User Summary")

	return nil
}
