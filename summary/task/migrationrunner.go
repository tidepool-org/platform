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
	DefaultMigrationAvailableAfterDurationMinimum = 5 * time.Minute
	DefaultMigrationAvailableAfterDurationMaximum = 5 * time.Minute
	MigrationTaskDurationMaximum                  = 4 * time.Minute
	DefaultMigrationWorkerBatchSize               = 500
	MigrationWorkerCount                          = 1
	MigrationType                                 = "org.tidepool.summary.migrate"
)

type MigrationRunner struct {
	authClient  auth.Client
	dataClient  dataClient.Client
	summaryType string
	logger      log.Logger
}

func NewDefaultMigrationTaskCreate(summaryType string) *task.TaskCreate {
	typ := MigrationType + "." + summaryType
	return &task.TaskCreate{
		Name:          pointer.FromAny(typ),
		Type:          typ,
		Priority:      5,
		AvailableTime: pointer.FromAny(time.Now().UTC()),
		Data: map[string]any{
			ConfigMinInterval: DefaultMigrationAvailableAfterDurationMinimum,
			ConfigMaxInterval: DefaultMigrationAvailableAfterDurationMaximum,
			ConfigBatch:       DefaultMigrationWorkerBatchSize,
		},
	}
}

func NewMigrationRunner(logger log.Logger, authClient auth.Client, dataClient dataClient.Client, summaryType string) (*MigrationRunner, error) {
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
		return nil, errors.Newf("summary type \"%s\" not supported by migration runner", summaryType)
	}

	return &MigrationRunner{
		authClient:  authClient,
		dataClient:  dataClient,
		summaryType: summaryType,
		logger:      logger,
	}, nil
}

func (r *MigrationRunner) GetRunnerType() string {
	return MigrationType + "." + r.summaryType
}

func (r *MigrationRunner) GetRunnerDeadline() time.Time {
	return time.Now().Add(MigrationTaskDurationMaximum * 3)
}

func (r *MigrationRunner) GetRunnerTimeout() time.Duration {
	return MigrationTaskDurationMaximum * 2
}

func (r *MigrationRunner) GetRunnerDurationMaximum() time.Duration {
	return MigrationTaskDurationMaximum
}

func (r *MigrationRunner) Run(ctx context.Context, tsk *task.Task) {
	ctx = auth.NewContextWithServerSessionTokenProvider(ctx, r.authClient)
	deadline := time.Now().Add(r.GetRunnerDurationMaximum())
	if taskRunner, err := NewMigrationTaskRunner(ctx, r.logger, r.authClient, r.dataClient, r.summaryType, tsk, deadline); err != nil {
		r.logger.WithError(err).Warn("Unable to create task runner")
	} else {
		taskRunner.Run()
	}
}

type MigrationTaskRunner struct {
	context     context.Context
	authClient  auth.Client
	dataClient  dataClient.Client
	summaryType string
	task        *task.Task
	logger      log.Logger
	deadline    time.Time
}

func NewMigrationTaskRunner(ctx context.Context, logger log.Logger, authClient auth.Client, dataClient dataClient.Client, summaryType string, tsk *task.Task, deadline time.Time) (*MigrationTaskRunner, error) {
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
		return nil, errors.Newf("summary type \"%s\" not supported by migration runner", summaryType)
	}
	if tsk == nil {
		return nil, errors.New("task is missing")
	}
	if deadline.Before(time.Now()) {
		return nil, errors.New("deadline is invalid")
	}

	return &MigrationTaskRunner{
		context:     ctx,
		authClient:  authClient,
		dataClient:  dataClient,
		summaryType: summaryType,
		task:        tsk,
		logger:      logger,
	}, nil
}

func (t *MigrationTaskRunner) getBatch() int {
	batch, ok := t.task.Data[ConfigBatch].(int)
	if !ok || batch < 1 {
		batch = DefaultMigrationWorkerBatchSize
		t.task.Data[ConfigBatch] = batch
	}

	return batch
}

func (t *MigrationTaskRunner) Run() {
	t.task.ClearError()
	if err := t.run(); err == nil {
		t.rescheduleTask()
	} else if !t.task.HasError() {
		t.rescheduleTaskWithResourceError(err)
	}
}

func (t *MigrationTaskRunner) run() error {
	pagination := page.NewPagination()
	pagination.Size = t.getBatch()
	typ := t.summaryType

	t.logger.Infof("Searching for User %s Summaries requiring Migration", typ)
	outdatedUserIds, err := t.dataClient.GetMigratableUserIDs(t.context, typ, pagination)
	if err != nil {
		return err
	}
	if len(outdatedUserIds) == 0 {
		t.logger.Infof("No %s Summaries requiring migrations found", typ)
		return nil
	}

	t.logger.Infof("Found batch of %d %s Summaries to Migrate", len(outdatedUserIds), typ)

	t.logger.Debugf("Starting User %s Summary Migration", typ)
	err = updateSummaries(t.context, t.logger, t.dataClient, typ, outdatedUserIds, MigrationWorkerCount, t.deadline, "Migrating")
	if err != nil {
		return err
	}
	t.logger.Debugf("Finished User %s Summary Migration", typ)

	return nil
}

func (t *MigrationTaskRunner) rescheduleTaskWithResourceError(err error) {
	t.rescheduleTaskWithError(ErrorResourceFailureError(err))
}

// Reschedule task for next run. Append error to task.
func (t *MigrationTaskRunner) rescheduleTaskWithError(err error) {
	t.task.AppendError(err)
	t.rescheduleTask()
}

func (t *MigrationTaskRunner) rescheduleTask() {
	minSeconds, ok := t.task.Data[ConfigMinInterval].(int)
	if !ok || minSeconds < 1 {
		minSeconds = int(DefaultMigrationAvailableAfterDurationMinimum.Seconds())
		t.task.Data[ConfigMinInterval] = minSeconds
	}
	maxSeconds, ok := t.task.Data[ConfigMaxInterval].(int)
	if !ok || maxSeconds < minSeconds {
		maxSeconds = int(DefaultMigrationAvailableAfterDurationMaximum.Seconds())
		t.task.Data[ConfigMaxInterval] = maxSeconds
	}
	t.task.RepeatAvailableAfter(GenerateNextTime(minSeconds, maxSeconds))
}
