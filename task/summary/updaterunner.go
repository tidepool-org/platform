package summary

import (
	"context"
	"math/rand"
	"time"

	"github.com/tidepool-org/platform/page"

	"github.com/tidepool-org/platform/auth"
	dataClient "github.com/tidepool-org/platform/data/client"
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
	UpdateAvailableAfterDurationMaximum = 4 * time.Minute
	UpdateAvailableAfterDurationMinimum = 2 * time.Minute
	UpdateTaskDurationMaximum           = 5 * time.Minute
	UpdateWorkerCount                   = 8
	UpdateWorkerBatchSize               = 500
)

type UpdateRunner struct {
	logger          log.Logger
	versionReporter version.Reporter
	authClient      auth.Client
	dataClient      dataClient.Client
}

func NewUpdateRunner(logger log.Logger, versionReporter version.Reporter, authClient auth.Client, dataClient dataClient.Client) (*UpdateRunner, error) {
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

	return &UpdateRunner{
		logger:          logger,
		versionReporter: versionReporter,
		authClient:      authClient,
		dataClient:      dataClient,
	}, nil
}

func (r *UpdateRunner) CanRunTask(tsk *task.Task) bool {
	return tsk != nil && tsk.Type == UpdateType
}

func (r *UpdateRunner) GenerateNextTime() time.Duration {
	randTime := time.Duration(rand.Int63n(int64(UpdateAvailableAfterDurationMaximum - UpdateAvailableAfterDurationMinimum + 1)))
	return UpdateAvailableAfterDurationMinimum + randTime
}

func (r *UpdateRunner) Run(ctx context.Context, tsk *task.Task) {
	now := time.Now()

	ctx = log.NewContextWithLogger(ctx, r.logger)

	tsk.ClearError()

	if serverSessionToken, sErr := r.authClient.ServerSessionToken(); sErr != nil {
		tsk.AppendError(errors.Wrap(sErr, "unable to get server session token"))
	} else {
		ctx = auth.NewContextWithServerSessionToken(ctx, serverSessionToken)

		if taskRunner, tErr := NewUpdateTaskRunner(r, tsk); tErr != nil {
			tsk.AppendError(errors.Wrap(tErr, "unable to create task runner"))
		} else if tErr = taskRunner.Run(ctx); tErr != nil {
			tsk.AppendError(errors.Wrap(tErr, "unable to run task runner"))
		}
	}

	if !tsk.IsFailed() {
		tsk.RepeatAvailableAfter(r.GenerateNextTime())
	}

	if taskDuration := time.Since(now); taskDuration > UpdateTaskDurationMaximum {
		r.logger.WithField("taskDuration", taskDuration.Truncate(time.Millisecond).Seconds()).Warn("Task duration exceeds maximum")
	}
}

type UpdateTaskRunner struct {
	*UpdateRunner
	task      *task.Task
	context   context.Context
	validator structure.Validator
}

func NewUpdateTaskRunner(rnnr *UpdateRunner, tsk *task.Task) (*UpdateTaskRunner, error) {
	if rnnr == nil {
		return nil, errors.New("runner is missing")
	}
	if tsk == nil {
		return nil, errors.New("task is missing")
	}

	return &UpdateTaskRunner{
		UpdateRunner: rnnr,
		task:         tsk,
	}, nil
}

func (t *UpdateTaskRunner) Run(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is missing")
	}

	t.context = ctx
	t.validator = structureValidator.New()

	t.logger.Info("Searching for User CGM Summaries requiring Update")
	pagination := page.NewPagination()
	pagination.Size = UpdateWorkerBatchSize
	outdatedCGMSummaryUserIDs, err := t.dataClient.GetOutdatedUserIDs(t.context, "cgm", pagination)
	if err != nil {
		return err
	}

	t.logger.Info("Searching for User BGM Summaries requiring Update")
	outdatedBGMSummaryUserIDs, err := t.dataClient.GetOutdatedUserIDs(t.context, "bgm", pagination)
	if err != nil {
		return err
	}

	t.logger.Debug("Starting User CGM Summary Update")
	if err := t.UpdateCGMSummaries(outdatedCGMSummaryUserIDs); err != nil {
		return err
	}
	t.logger.Debug("Finished User CGM Summary Update")

	t.logger.Debug("Starting User BGM Summary Update")
	if err := t.UpdateBGMSummaries(outdatedBGMSummaryUserIDs); err != nil {
		return err
	}
	t.logger.Debug("Finished User BGM Summary Update")

	return nil
}

func (t *UpdateTaskRunner) UpdateCGMSummaries(userIDs []string) error {
	eg, ctx := errgroup.WithContext(t.context)

	eg.Go(func() error {
		sem := semaphore.NewWeighted(UpdateWorkerCount)
		for _, userID := range userIDs {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}

			// we can't pass arguments to errgroup goroutines
			// we need to explicitly redefine the variables,
			// because we're launching the goroutines in a loop
			userID := userID
			eg.Go(func() error {
				defer sem.Release(1)
				return t.UpdateCGMUserSummary(userID)
			})
		}

		return nil
	})
	return eg.Wait()
}

func (t *UpdateTaskRunner) UpdateBGMSummaries(userIDs []string) error {
	eg, ctx := errgroup.WithContext(t.context)

	eg.Go(func() error {
		sem := semaphore.NewWeighted(UpdateWorkerCount)
		for _, userID := range userIDs {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}

			// we can't pass arguments to errgroup goroutines
			// we need to explicitly redefine the variables,
			// because we're launching the goroutines in a loop
			userID := userID
			eg.Go(func() error {
				defer sem.Release(1)
				return t.UpdateBGMUserSummary(userID)
			})
		}

		return nil
	})
	return eg.Wait()
}

func (t *UpdateTaskRunner) UpdateCGMUserSummary(userID string) error {
	t.logger.WithField("UserID", userID).Debug("Updating User CGM Summary")

	// update summary
	_, err := t.dataClient.UpdateCGMSummary(t.context, userID)
	if err != nil {
		return err
	}

	t.logger.WithField("UserID", userID).Debug("Finished Updating User CGM Summary")

	return nil
}

func (t *UpdateTaskRunner) UpdateBGMUserSummary(userID string) error {
	t.logger.WithField("UserID", userID).Debug("Updating User BGM Summary")

	// update summary
	_, err := t.dataClient.UpdateBGMSummary(t.context, userID)
	if err != nil {
		return err
	}

	t.logger.WithField("UserID", userID).Debug("Finished Updating User BGM Summary")

	return nil
}
