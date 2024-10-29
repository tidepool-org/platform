package summary

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/auth"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/version"
)

const (
	DefaultBackfillAvailableAfterDurationMinimum = 23 * time.Hour
	DefaultBackfillAvailableAfterDurationMaximum = 24 * time.Hour
	BackfillTaskDurationMaximum                  = 15 * time.Minute
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

func (r *BackfillRunner) GetRunnerType() string {
	return BackfillType
}

func (r *BackfillRunner) GetRunnerDeadline() time.Time {
	return time.Now().Add(BackfillTaskDurationMaximum * 3)
}

func (r *BackfillRunner) GetRunnerTimeout() time.Duration {
	return BackfillTaskDurationMaximum * 2
}

func (r *BackfillRunner) GetRunnerDurationMaximum() time.Duration {
	return BackfillTaskDurationMaximum
}

func (r *BackfillRunner) GenerateNextTime(interval MinuteRange) time.Duration {

	Min := time.Duration(interval.Min) * time.Second
	Max := time.Duration(interval.Max) * time.Second

	randTime := time.Duration(rand.Int63n(int64(Max - Min + 1)))
	return Min + randTime
}

func (r *BackfillRunner) GetConfig(tsk *task.Task) TaskConfiguration {
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
			if configErr := ValidateConfig(config, false); configErr != nil {
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
	ctx = log.NewContextWithLogger(ctx, r.logger)
	ctx = auth.NewContextWithServerSessionTokenProvider(ctx, r.authClient)

	tsk.ClearError()

	config := r.GetConfig(tsk)

	if taskRunner, tErr := NewBackfillTaskRunner(r, tsk); tErr != nil {
		tsk.AppendError(fmt.Errorf("unable to create task runner: %w", tErr))
	} else if tErr = taskRunner.Run(ctx); tErr != nil {
		tsk.AppendError(tErr)
	}

	if !tsk.IsFailed() {
		tsk.RepeatAvailableAfter(r.GenerateNextTime(config.Interval))
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
	t.validator = structureValidator.New(log.LoggerFromContext(ctx))

	for _, typ := range []string{"continuous"} {
		t.logger.Debugf("Starting User %s Summary Backfill", typ)
		count, err := t.dataClient.BackfillSummaries(t.context, typ)
		if err != nil {
			return err
		}
		t.logger.Infof("Backfilled %d %s summaries", count, typ)
	}
	return nil
}
