package summary

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"math/rand"
	"time"

	"github.com/tidepool-org/platform/auth"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/version"
)

const (
	DefaultBackfillAvailableAfterDurationMaximum = 24 * time.Hour
	DefaultBackfillAvailableAfterDurationMinimum = 23 * time.Hour
	BackfillTaskDurationMaximum                  = 5 * time.Minute
	BackfillType                                 = "org.tidepool.summary.backfill"
)

type BackfillRunner struct {
	logger          log.Logger
	versionReporter version.Reporter
	authClient      auth.Client
	dataClient      dataClient.Client
}

func NewBackfillRunner(logger log.Logger, versionReporter version.Reporter, authClient auth.Client, dataClient dataClient.Client) (*BackfillRunner, error) {
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

	return &BackfillRunner{
		logger:          logger,
		versionReporter: versionReporter,
		authClient:      authClient,
		dataClient:      dataClient,
	}, nil
}

func BackfillTaskName() string {
	return fmt.Sprintf("%s", BackfillType)
}

func (r *BackfillRunner) CanRunTask(tsk *task.Task) bool {
	return tsk != nil && tsk.Type == BackfillType
}

func (r *BackfillRunner) GenerateNextTime(interval MinuteRange) time.Duration {

	Min := time.Duration(interval.Min) * time.Minute
	Max := time.Duration(interval.Max) * time.Minute

	randTime := time.Duration(rand.Int63n(int64(Max - Min + 1)))
	return Min + randTime
}

func (r *BackfillRunner) GetConfig(tsk *task.Task) TaskConfiguration {
	var config TaskConfiguration
	var valid bool
	if raw, ok := tsk.Data["config"]; ok {
		unmarshalError := bson.Unmarshal(raw.([]byte), &config)
		if unmarshalError != nil {
			r.logger.WithField("unmarshalError", unmarshalError).Warn("Task configuration invalid, falling back to defaults.")
		} else {
			if configErr := ValidateConfig(config); configErr != nil {
				r.logger.WithField("validationError", configErr).Warn("Task configuration invalid, falling back to defaults.")
			} else {
				valid = true
			}
		}
	}

	if !valid {
		config = NewDefaultBackfillConfig()

		if tsk.Data == nil {
			tsk.Data = make(map[string]interface{})
		}
		tsk.Data["config"] = config
	}

	return config
}
func (r *BackfillRunner) Run(ctx context.Context, tsk *task.Task) {
	now := time.Now()

	ctx = log.NewContextWithLogger(ctx, r.logger)

	tsk.ClearError()

	config := r.GetConfig(tsk)

	if serverSessionToken, sErr := r.authClient.ServerSessionToken(); sErr != nil {
		tsk.AppendError(errors.Wrap(sErr, "unable to get server session token"))
	} else {
		ctx = auth.NewContextWithServerSessionToken(ctx, serverSessionToken)

		if taskRunner, tErr := NewBackfillTaskRunner(r, tsk); tErr != nil {
			tsk.AppendError(errors.Wrap(tErr, "unable to create task runner"))
		} else if tErr = taskRunner.Run(ctx); tErr != nil {
			tsk.AppendError(tErr)
		}
	}

	if !tsk.IsFailed() {
		tsk.RepeatAvailableAfter(r.GenerateNextTime(config.Interval))
	}

	if taskDuration := time.Since(now); taskDuration > BackfillTaskDurationMaximum {
		r.logger.WithField("taskDuration", taskDuration.Truncate(time.Millisecond).Seconds()).Warn("Task duration exceeds maximum")
	}
}

type BackfillTaskRunner struct {
	*BackfillRunner
	task      *task.Task
	context   context.Context
	validator structure.Validator
}

func NewBackfillTaskRunner(rnnr *BackfillRunner, tsk *task.Task) (*BackfillTaskRunner, error) {
	if rnnr == nil {
		return nil, errors.New("runner is missing")
	}
	if tsk == nil {
		return nil, errors.New("task is missing")
	}

	return &BackfillTaskRunner{
		BackfillRunner: rnnr,
		task:           tsk,
	}, nil
}

func (t *BackfillTaskRunner) Run(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is missing")
	}

	t.context = ctx
	t.validator = structureValidator.New()

	t.logger.Debug("Starting User CGM Summary Creation")
	count, err := t.dataClient.BackfillSummaries(t.context, "cgm")
	if err != nil {
		return err
	}
	t.logger.Info(fmt.Sprintf("Backfilled %d CGM summaries", count))

	t.logger.Debug("Starting User BGM Summary Creation")
	count, err = t.dataClient.BackfillSummaries(t.context, "bgm")
	if err != nil {
		return err
	}
	t.logger.Info(fmt.Sprintf("Backfilled %d BGM summaries", count))

	return nil
}
