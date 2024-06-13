package summary

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/page"

	"errors"

	"github.com/tidepool-org/platform/auth"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/version"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

const (
	DefaultUpdateAvailableAfterDurationMinimum = 20 * time.Second
	DefaultUpdateAvailableAfterDurationMaximum = 30 * time.Second
	UpdateTaskDurationMaximum                  = 2 * time.Minute
	DefaultUpdateWorkerBatchSize               = 500
	UpdateWorkerCount                          = 10
	UpdateType                                 = "org.tidepool.summary.update"

	IterLimit = 4
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

func (r *UpdateRunner) GetRunnerType() string {
	return UpdateType
}

func (r *UpdateRunner) GetRunnerDeadline() time.Time {
	return time.Now().Add(UpdateTaskDurationMaximum * 3)
}

func (r *UpdateRunner) GetRunnerMaximumDuration() time.Duration {
	return UpdateTaskDurationMaximum
}

func (r *UpdateRunner) GenerateNextTime(interval MinuteRange) time.Duration {
	Min := time.Duration(interval.Min) * time.Second
	Max := time.Duration(interval.Max) * time.Second

	randTime := time.Duration(rand.Int63n(int64(Max - Min + 1)))
	return Min + randTime
}

func (r *UpdateRunner) GetConfig(tsk *task.Task) TaskConfiguration {
	var config TaskConfiguration
	var valid bool
	if raw, ok := tsk.Data["config"]; ok {
		// this is abuse of marshal/unmarshal, this was done with interface{} target when loading the task,
		// but we require something more specific at this point
		bs, _ := bson.Marshal(raw)
		unmarshalError := bson.Unmarshal(bs, &config)
		if unmarshalError != nil {
			r.logger.WithField("unmarshalError", unmarshalError).Warn("Task configuration invalid, falling back to defaults.")
		} else {
			if configErr := ValidateConfig(config, true); configErr != nil {
				r.logger.WithField("validationError", configErr).Warn("Task configuration invalid, falling back to defaults.")
			} else {
				valid = true
			}
		}
	}

	if !valid {
		config = NewDefaultUpdateConfig()

		if tsk.Data == nil {
			tsk.Data = make(map[string]interface{})
		}
		tsk.Data["config"] = config
	}

	return config
}

func (r *UpdateRunner) Run(ctx context.Context, tsk *task.Task) bool {
	now := time.Now()

	ctx = log.NewContextWithLogger(ctx, r.logger)
	ctx = auth.NewContextWithServerSessionTokenProvider(ctx, r.authClient)

	tsk.ClearError()

	config := r.GetConfig(tsk)

	if taskRunner, tErr := NewUpdateTaskRunner(r, tsk); tErr != nil {
		tsk.AppendError(fmt.Errorf("unable to create task runner: %w", tErr))
	} else if tErr = taskRunner.Run(ctx, *config.Batch); tErr != nil {
		tsk.AppendError(fmt.Errorf("unable to run task runner: %w", tErr))
	}

	if !tsk.IsFailed() {
		tsk.RepeatAvailableAfter(r.GenerateNextTime(config.Interval))
	}

	if taskDuration := time.Since(now); taskDuration > UpdateTaskDurationMaximum {
		r.logger.WithField("taskDuration", taskDuration.Truncate(time.Millisecond).Seconds()).Warn("Task duration exceeds maximum")
	}

	return true
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

func (t *UpdateTaskRunner) Run(ctx context.Context, batch int) error {
	if ctx == nil {
		return errors.New("context is missing")
	}

	t.context = ctx
	t.validator = structureValidator.New(log.LoggerFromContext(ctx))
	targetTime := time.Now().UTC().Add(-1 * time.Minute)

	pagination := page.NewPagination()
	pagination.Size = batch

	t.logger.Debug("Starting User CGM Summary Update")
	iCount := 0
	// this loop is a bit odd looking, we are iterating until the end of the previous loop is past the target
	// this avoids a round trip, and allows the default time zero value to work as a starter
	for {
		t.logger.Info("Searching for User CGM Summaries requiring Update")
		outdatedCGM, err := t.dataClient.GetOutdatedUserIDs(t.context, "cgm", pagination)
		if err != nil {
			return err
		}

		if err = t.UpdateCGMSummaries(outdatedCGM.UserIds); err != nil {
			return err
		}

		if iCount > IterLimit {
			t.logger.Warn("Exiting CGM batch loop early, too many iterations")
			break
		}

		if outdatedCGM.End.After(targetTime) || outdatedCGM.End.IsZero() {
			// we are sufficiently caught up
			break
		}

		iCount++
	}
	t.logger.Debug("Finished User CGM Summary Update")

	t.logger.Debug("Starting User BGM Summary Update")
	iCount = 0
	for {
		t.logger.Info("Searching for User BGM Summaries requiring Update")
		outdatedBGM, err := t.dataClient.GetOutdatedUserIDs(t.context, "bgm", pagination)
		if err != nil {
			return err
		}

		if err = t.UpdateBGMSummaries(outdatedBGM.UserIds); err != nil {
			return err
		}

		if iCount > IterLimit {
			t.logger.Warn("Exiting BGM batch loop early, too many iterations")
			break
		}

		if outdatedBGM.End.After(targetTime) || outdatedBGM.End.IsZero() {
			// we are sufficiently caught up
			break
		}

		iCount++
	}
	t.logger.Debug("Finished User BGM Summary Update")

	return nil
}

func (t *UpdateTaskRunner) UpdateCGMSummaries(outdatedUserIds []string) error {
	eg, ctx := errgroup.WithContext(t.context)

	eg.Go(func() error {
		sem := semaphore.NewWeighted(UpdateWorkerCount)
		for _, userID := range outdatedUserIds {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}

			// we can't pass arguments to errgroup goroutines
			// we need to explicitly redefine the variables,
			// because we're launching the goroutines in a loop
			userID := userID
			eg.Go(func() error {
				defer sem.Release(1)
				t.logger.WithField("UserID", userID).Debug("Updating User CGM Summary")

				// update summary
				_, err := t.dataClient.UpdateCGMSummary(t.context, userID)
				if err != nil {
					return err
				}

				t.logger.WithField("UserID", userID).Debug("Finished Updating User CGM Summary")

				return nil
			})
		}

		return nil
	})
	return eg.Wait()
}

func (t *UpdateTaskRunner) UpdateBGMSummaries(outdatedUserIds []string) error {
	eg, ctx := errgroup.WithContext(t.context)

	eg.Go(func() error {
		sem := semaphore.NewWeighted(UpdateWorkerCount)
		for _, userID := range outdatedUserIds {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}

			// we can't pass arguments to errgroup goroutines
			// we need to explicitly redefine the variables,
			// because we're launching the goroutines in a loop
			userID := userID
			eg.Go(func() error {
				defer sem.Release(1)
				t.logger.WithField("UserID", userID).Debug("Updating User BGM Summary")

				// update summary
				_, err := t.dataClient.UpdateBGMSummary(t.context, userID)
				if err != nil {
					return err
				}

				t.logger.WithField("UserID", userID).Debug("Finished Updating User BGM Summary")

				return nil
			})
		}

		return nil
	})
	return eg.Wait()
}
