package summary

import (
	"context"
	"math/rand"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/tidepool-org/platform/auth"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/version"
)

const (
	DefaultMigrationAvailableAfterDurationMaximum = 5 * time.Minute
	DefaultMigrationAvailableAfterDurationMinimum = 5 * time.Minute
	MigrationTaskDurationMaximum                  = 4 * time.Minute
	DefaultMigrationWorkerBatchSize               = 500
	MigrationWorkerCount                          = 1
	MigrationType                                 = "org.tidepool.summary.migrate"
)

type MigrationRunner struct {
	logger          log.Logger
	versionReporter version.Reporter
	authClient      auth.Client
	dataClient      dataClient.Client
}

func NewMigrationRunner(logger log.Logger, versionReporter version.Reporter, authClient auth.Client, dataClient dataClient.Client) (*MigrationRunner, error) {
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

	return &MigrationRunner{
		logger:          logger,
		versionReporter: versionReporter,
		authClient:      authClient,
		dataClient:      dataClient,
	}, nil
}

func (r *MigrationRunner) GetRunnerType() string {
	return MigrationType
}

func (r *MigrationRunner) GetRunnerDeadline() time.Time {
	return time.Now().Add(MigrationTaskDurationMaximum * 3)
}

func (r *MigrationRunner) GetRunnerMaximumDuration() time.Duration {
	return MigrationTaskDurationMaximum
}

func (r *MigrationRunner) GenerateNextTime(interval MinuteRange) time.Duration {
	Min := time.Duration(interval.Min) * time.Minute
	Max := time.Duration(interval.Max) * time.Minute

	randTime := time.Duration(rand.Int63n(int64(Max - Min + 1)))
	return Min + randTime
}

func (r *MigrationRunner) GetConfig(tsk *task.Task) TaskConfiguration {
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
		config = NewDefaultMigrationConfig()

		if tsk.Data == nil {
			tsk.Data = make(map[string]interface{})
		}
		tsk.Data["config"] = config
	}

	return config
}

func (r *MigrationRunner) Run(ctx context.Context, tsk *task.Task) bool {
	now := time.Now()

	ctx = log.NewContextWithLogger(ctx, r.logger)

	tsk.ClearError()

	config := r.GetConfig(tsk)

	if serverSessionToken, sErr := r.authClient.ServerSessionToken(); sErr != nil {
		tsk.AppendError(errors.Wrap(sErr, "unable to get server session token"))
	} else {
		ctx = auth.NewContextWithServerSessionToken(ctx, serverSessionToken)

		if taskRunner, tErr := NewMigrationTaskRunner(r, tsk); tErr != nil {
			tsk.AppendError(errors.Wrap(tErr, "unable to create task runner"))
		} else if tErr = taskRunner.Run(ctx, *config.Batch); tErr != nil {
			tsk.AppendError(errors.Wrap(tErr, "unable to run task runner"))
		}
	}

	if !tsk.IsFailed() {
		tsk.RepeatAvailableAfter(r.GenerateNextTime(config.Interval))
	}

	if taskDuration := time.Since(now); taskDuration > UpdateTaskDurationMaximum {
		r.logger.WithField("taskDuration", taskDuration.Truncate(time.Millisecond).Seconds()).Warn("Task duration exceeds maximum")
	}

	return true
}

type MigrationTaskRunner struct {
	*MigrationRunner
	task      *task.Task
	context   context.Context
	validator structure.Validator
}

func NewMigrationTaskRunner(rnnr *MigrationRunner, tsk *task.Task) (*MigrationTaskRunner, error) {
	if rnnr == nil {
		return nil, errors.New("runner is missing")
	}
	if tsk == nil {
		return nil, errors.New("task is missing")
	}

	return &MigrationTaskRunner{
		MigrationRunner: rnnr,
		task:            tsk,
	}, nil
}

func (t *MigrationTaskRunner) Run(ctx context.Context, batch int) error {
	if ctx == nil {
		return errors.New("context is missing")
	}

	t.context = ctx
	t.validator = structureValidator.New()

	pagination := page.NewPagination()
	pagination.Size = batch

	t.logger.Info("Searching for User CGM Summaries requiring Update")
	outdatedCGMSummaryUserIDs, err := t.dataClient.GetMigratableUserIDs(t.context, "cgm", pagination)
	if err != nil {
		return err
	}

	t.logger.Info("Searching for User BGM Summaries requiring Update")
	outdatedBGMSummaryUserIDs, err := t.dataClient.GetMigratableUserIDs(t.context, "bgm", pagination)
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

func (t *MigrationTaskRunner) UpdateCGMSummaries(userIDs []string) error {
	eg, ctx := errgroup.WithContext(t.context)

	eg.Go(func() error {
		sem := semaphore.NewWeighted(MigrationWorkerCount)
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

func (t *MigrationTaskRunner) UpdateBGMSummaries(userIDs []string) error {
	eg, ctx := errgroup.WithContext(t.context)

	eg.Go(func() error {
		sem := semaphore.NewWeighted(MigrationWorkerCount)
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

func (t *MigrationTaskRunner) UpdateCGMUserSummary(userID string) error {
	t.logger.WithField("UserID", userID).Debug("Updating User CGM Summary")

	// update summary
	_, err := t.dataClient.UpdateCGMSummary(t.context, userID)
	if err != nil {
		return err
	}

	t.logger.WithField("UserID", userID).Debug("Finished Updating User CGM Summary")

	return nil
}

func (t *MigrationTaskRunner) UpdateBGMUserSummary(userID string) error {
	t.logger.WithField("UserID", userID).Debug("Updating User BGM Summary")

	// update summary
	_, err := t.dataClient.UpdateBGMSummary(t.context, userID)
	if err != nil {
		return err
	}

	t.logger.WithField("UserID", userID).Debug("Finished Updating User BGM Summary")

	return nil
}