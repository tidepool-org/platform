package task

import (
	"context"
	"github.com/tidepool-org/platform/pointer"
	"slices"
	"time"

	"github.com/tidepool-org/platform/auth"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
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
}

func NewDefaultUpdateTaskCreate(summaryType string) *task.TaskCreate {
	typ := UpdateType + "." + summaryType
	return &task.TaskCreate{
		Name:          pointer.FromAny(typ),
		Type:          typ,
		Priority:      5,
		AvailableTime: pointer.FromAny(time.Now().UTC()),
		Data: map[string]interface{}{
			"minInterval": DefaultUpdateAvailableAfterDurationMinimum,
			"maxInterval": DefaultUpdateAvailableAfterDurationMaximum,
			"batch":       DefaultUpdateWorkerBatchSize,
		},
	}
}

func NewUpdateRunner(authClient auth.Client, dataClient dataClient.Client, summaryType string) (*UpdateRunner, error) {
	if authClient == nil {
		return nil, errors.New("auth client is missing")
	}
	if dataClient == nil {
		return nil, errors.New("data client is missing")
	}
	if slices.Contains(SummaryTypes, summaryType) {
		return nil, errors.Newf("summary type \"%s\" not supported by update runner", summaryType)
	}

	return &UpdateRunner{
		authClient:  authClient,
		dataClient:  dataClient,
		summaryType: summaryType,
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
	if taskRunner, err := NewUpdateTaskRunner(r, tsk); err != nil {
		log.LoggerFromContext(ctx).WithError(err).Warn("Unable to create task runner")
	} else {
		taskRunner.Run(ctx)
	}
}

type UpdateTaskRunner struct {
	*UpdateRunner
	task     *task.Task
	context  context.Context
	logger   log.Logger
	deadline time.Time
}

func NewUpdateTaskRunner(runner *UpdateRunner, tsk *task.Task) (*UpdateTaskRunner, error) {
	if runner == nil {
		return nil, errors.New("provider is missing")
	}
	if tsk == nil {
		return nil, errors.New("task is missing")
	}

	return &UpdateTaskRunner{
		UpdateRunner: runner,
		task:         tsk,
	}, nil
}

func (t *UpdateTaskRunner) Run(ctx context.Context) {
	t.context = ctx
	t.logger = log.LoggerFromContext(t.context)
	t.deadline = time.Now().Add(t.GetRunnerDurationMaximum())

	t.task.ClearError()
	if err := t.run(); err == nil {
		t.rescheduleTask()
	} else if !t.task.HasError() {
		t.rescheduleTaskWithResourceError(err)
	}
}

func (t *UpdateTaskRunner) getBatch() int {
	batch, ok := t.task.Data["batch"].(int)
	if !ok || batch < 1 {
		batch = DefaultUpdateWorkerBatchSize
		t.task.Data["batch"] = batch
	}

	return batch
}

func (t *UpdateTaskRunner) run() error {
	pagination := page.NewPagination()
	pagination.Size = t.getBatch()
	targetTime := time.Now().UTC().Add(-1 * time.Minute)
	typ := t.summaryType

	t.logger.Debugf("Starting User %s Summary Update", typ)

	for i := 1; i <= IterLimit; i++ {
		t.logger.Infof("Searching for User %s Summaries requiring Update", typ)
		outdated, err := t.dataClient.GetOutdatedUserIDs(t.context, "cgm", pagination)
		if err != nil {
			return err
		}
		if len(outdated.UserIds) == 0 {
			t.logger.Infof("No %s Summaries requiring updates found", typ)
			return nil
		}

		t.logger.Infof("Found batch of %d %s Summaries to Migrate", len(outdated.UserIds), typ)

		err = updateSummaries(t.context, t.dataClient, typ, outdated.UserIds)
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
	t.task.AppendError(err)
	t.rescheduleTask()
}

func (t *UpdateTaskRunner) rescheduleTask() {
	minSeconds, ok := t.task.Data["MinInterval"].(int)
	if !ok || minSeconds < 1 {
		minSeconds = int(DefaultUpdateAvailableAfterDurationMinimum.Seconds())
		t.task.Data["minInterval"] = minSeconds
	}
	maxSeconds, ok := t.task.Data["MaxInterval"].(int)
	if !ok || maxSeconds < minSeconds {
		maxSeconds = int(DefaultUpdateAvailableAfterDurationMaximum.Seconds())
		t.task.Data["maxInterval"] = maxSeconds
	}
	t.task.RepeatAvailableAfter(GenerateNextTime(minSeconds, maxSeconds))
}
