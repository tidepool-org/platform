package summary

import (
	"context"
	"fmt"
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
	BackfillAvailableAfterDurationMaximum = 4 * time.Hour
	BackfillAvailableAfterDurationMinimum = 4 * time.Hour
	BackfillTaskDurationMaximum           = 30 * time.Minute
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

func (r *BackfillRunner) CanRunTask(tsk *task.Task) bool {
	return tsk != nil && tsk.Type == BackfillType
}

func (r *BackfillRunner) GenerateNextTime() time.Duration {
	randTime := time.Duration(rand.Int63n(int64(BackfillAvailableAfterDurationMaximum - BackfillAvailableAfterDurationMinimum + 1)))
	return BackfillAvailableAfterDurationMinimum + randTime
}

func (r *BackfillRunner) Run(ctx context.Context, tsk *task.Task) {
	now := time.Now()

	ctx = log.NewContextWithLogger(ctx, r.logger)

	tsk.ClearError()

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
		tsk.RepeatAvailableAfter(r.GenerateNextTime())
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

	t.logger.Debug("Starting User Summary Creation")

	count, err := t.dataClient.BackfillSummaries(t.context)
	if err != nil {
		return err
	}

	t.logger.Info(fmt.Sprintf("Backfilled %d summaries", count))

	return nil
}
