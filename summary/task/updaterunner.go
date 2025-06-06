package task

import (
	"context"
	"slices"
	"time"

	"github.com/tidepool-org/platform/auth"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/task"
)

const (
	DefaultUpdateAvailableAfterDurationMinimum = 20 * time.Second
	DefaultUpdateAvailableAfterDurationMaximum = 30 * time.Second
	UpdateTaskDurationMaximum                  = 2 * time.Minute
	DefaultUpdateWorkerBatchSize               = 250
	UpdateWorkerCount                          = 10
	UpdateType                                 = "org.tidepool.summary.update"
	IterLimit                                  = 3
)

type UpdateRunner struct {
	authClient  auth.Client
	dataClient  dataClient.Client
	summaryType string
	logger      log.Logger
}

func NewDefaultUpdateTaskCreate(summaryType string) *task.TaskCreate {
	typ := UpdateType + "." + summaryType
	return &task.TaskCreate{
		Name:          pointer.FromAny(typ),
		Type:          typ,
		Priority:      5,
		AvailableTime: pointer.FromAny(time.Now().UTC()),
		Data: map[string]any{
			ConfigMinInterval: int32(DefaultUpdateAvailableAfterDurationMinimum.Seconds()),
			ConfigMaxInterval: int32(DefaultUpdateAvailableAfterDurationMaximum.Seconds()),
			ConfigBatch:       int32(DefaultUpdateWorkerBatchSize),
		},
	}
}

func NewUpdateRunner(logger log.Logger, authClient auth.Client, dataClient dataClient.Client, summaryType string) (*UpdateRunner, error) {
	if authClient == nil {
		return nil, errors.New("auth client is missing")
	}
	if dataClient == nil {
		return nil, errors.New("data client is missing")
	}
	if !slices.Contains(SummaryTypes, summaryType) {
		return nil, errors.Newf("summary type \"%s\" not supported by update runner", summaryType)
	}

	return &UpdateRunner{
		authClient:  authClient,
		dataClient:  dataClient,
		summaryType: summaryType,
		logger:      logger,
	}, nil
}

func (r *UpdateRunner) GetRunnerType() string {
	return UpdateType + "." + r.summaryType
}

func (r *UpdateRunner) GetRunnerDeadline() time.Time {
	return time.Now().Add(UpdateTaskDurationMaximum * 3)
}

func (r *UpdateRunner) GetRunnerTimeout() time.Duration {
	return UpdateTaskDurationMaximum * 2
}

func (r *UpdateRunner) GetRunnerDurationMaximum() time.Duration {
	return UpdateTaskDurationMaximum
}

func (r *UpdateRunner) Run(ctx context.Context, tsk *task.Task) {
	ctx = auth.NewContextWithServerSessionTokenProvider(ctx, r.authClient)
	deadline := time.Now().Add(UpdateTaskDurationMaximum)
	if taskRunner, err := NewUpdateTaskRunner(ctx, r.logger, r.authClient, r.dataClient, r.summaryType, tsk, deadline); err != nil {
		r.logger.WithError(err).Warn("Unable to create task runner")
	} else {
		taskRunner.Run()
	}
}

type UpdateTaskRunner struct {
	authClient  auth.Client
	dataClient  dataClient.Client
	summaryType string
	Task        *task.Task
	context     context.Context
	logger      log.Logger
	deadline    time.Time
}

func NewUpdateTaskRunner(ctx context.Context, logger log.Logger, authClient auth.Client, dataClient dataClient.Client, summaryType string, tsk *task.Task, deadline time.Time) (*UpdateTaskRunner, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if logger == nil {
		return nil, errors.New("logger is missing")
	}
	if authClient == nil {
		return nil, errors.New("auth client is missing")
	}
	if dataClient == nil {
		return nil, errors.New("data client is missing")
	}
	if !slices.Contains(SummaryTypes, summaryType) {
		return nil, errors.Newf("summary type \"%s\" not supported by update runner", summaryType)
	}
	if tsk == nil {
		return nil, errors.New("task is missing")
	}
	if deadline.Before(time.Now()) {
		return nil, errors.New("deadline is invalid")
	}

	return &UpdateTaskRunner{
		context:     ctx,
		authClient:  authClient,
		dataClient:  dataClient,
		summaryType: summaryType,
		Task:        tsk,
		logger:      logger,
		deadline:    deadline,
	}, nil
}

func (t *UpdateTaskRunner) Run() {
	t.Task.ClearError()
	if err := t.run(); err == nil {
		t.rescheduleTask()
	} else if !t.Task.HasError() {
		t.rescheduleTaskWithResourceError(err)
	}
}

func (t *UpdateTaskRunner) GetBatch() int {
	return int(ValueFromTaskDataWithFallback[int32](t.Task.Data, ConfigBatch, isGtZero, DefaultUpdateWorkerBatchSize))
}

func (t *UpdateTaskRunner) GetIntervalRange() (int, int) {
	minSeconds, minOk := t.Task.Data[ConfigMinInterval].(int32)
	maxSeconds, maxOk := t.Task.Data[ConfigMaxInterval].(int32)

	// reset both min and max if either is considered invalid to prevent an accidental mismatch
	if !minOk || !maxOk || minSeconds < 1 || minSeconds > maxSeconds {
		minSeconds = int32(DefaultUpdateAvailableAfterDurationMinimum.Seconds())
		t.Task.Data[ConfigMinInterval] = minSeconds
		maxSeconds = int32(DefaultUpdateAvailableAfterDurationMaximum.Seconds())
		t.Task.Data[ConfigMaxInterval] = maxSeconds

	}

	return int(minSeconds), int(maxSeconds)
}

func (t *UpdateTaskRunner) run() error {
	pagination := page.NewPagination()
	pagination.Size = t.GetBatch()
	targetTime := time.Now().UTC().Add(-1 * time.Minute)
	typ := t.summaryType

	t.logger.Debugf("Starting User %s Summary Update", typ)

	for i := 1; i <= IterLimit; i++ {
		t.logger.Infof("Searching for User %s Summaries requiring Update", typ)
		outdated, err := t.dataClient.GetOutdatedUserIDs(t.context, typ, pagination)
		if err != nil {
			return err
		}
		if len(outdated.UserIds) == 0 {
			t.logger.Infof("No %s Summaries requiring updates found", typ)
			return nil
		}

		t.logger.Infof("Found batch of %d %s Summaries to Update", len(outdated.UserIds), typ)

		err = updateSummaries(t.context, t.logger, t.dataClient, typ, outdated.UserIds, UpdateWorkerCount, t.deadline, "Updating")
		if err != nil {
			return err
		}

		if outdated.End.After(targetTime) || outdated.End.IsZero() {
			// we are sufficiently caught up
			break
		}

		if i == IterLimit {
			t.logger.Warnf("Reached iteration limit in updating %s summaries, exiting", typ)
		}
	}
	t.logger.Debugf("Finished User %s Summary Update", typ)

	return nil
}

func (t *UpdateTaskRunner) rescheduleTaskWithResourceError(err error) {
	t.rescheduleTaskWithError(ErrorResourceFailureError(err))
}

// Reschedule task for next run. Append error to task.
func (t *UpdateTaskRunner) rescheduleTaskWithError(err error) {
	t.Task.AppendError(err)
	t.rescheduleTask()
}

func (t *UpdateTaskRunner) rescheduleTask() {
	t.Task.RepeatAvailableAfter(GenerateNextTime(t.GetIntervalRange()))
}
