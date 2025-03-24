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
	MigrationWorkerCount                          = 10
	MigrationType                                 = "org.tidepool.summary.migrate"
)

type MigrationRunner struct {
	authClient  auth.Client
	dataClient  dataClient.Client
	summaryType string
}

func NewDefaultMigrationTaskCreate(summaryType string) *task.TaskCreate {
	typ := MigrationType + "." + summaryType
	return &task.TaskCreate{
		Name:          pointer.FromAny(typ),
		Type:          typ,
		Priority:      5,
		AvailableTime: pointer.FromAny(time.Now().UTC()),
		Data: map[string]interface{}{
			"minInterval": DefaultMigrationAvailableAfterDurationMinimum,
			"maxInterval": DefaultMigrationAvailableAfterDurationMaximum,
			"batch":       DefaultMigrationWorkerBatchSize,
		},
	}
}

func NewMigrationRunner(authClient auth.Client, dataClient dataClient.Client, summaryType string) (*MigrationRunner, error) {
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
	if taskRunner, err := NewMigrationTaskRunner(r.authClient, r.dataClient, r.summaryType, tsk); err != nil {
		log.LoggerFromContext(ctx).WithError(err).Warn("Unable to create task runner")
	} else {
		taskRunner.Run(ctx, time.Now().Add(r.GetRunnerDurationMaximum()))
	}
}

type MigrationTaskRunner struct {
	authClient  auth.Client
	dataClient  dataClient.Client
	summaryType string
	task        *task.Task
	context     context.Context
	logger      log.Logger
	deadline    time.Time
}

func NewMigrationTaskRunner(authClient auth.Client, dataClient dataClient.Client, summaryType string, tsk *task.Task) (*MigrationTaskRunner, error) {
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

	return &MigrationTaskRunner{
		authClient:  authClient,
		dataClient:  dataClient,
		summaryType: summaryType,
		task:        tsk,
	}, nil
}

func (t *MigrationTaskRunner) getBatch() int {
	batch, ok := t.task.Data["batch"].(int)
	if !ok || batch < 1 {
		batch = DefaultMigrationWorkerBatchSize
		t.task.Data["batch"] = batch
	}

	return batch
}

func (t *MigrationTaskRunner) Run(ctx context.Context, deadline time.Time) {
	t.context = ctx
	t.logger = log.LoggerFromContext(t.context)
	t.deadline = deadline

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
	err = updateSummaries(t.context, t.dataClient, typ, outdatedUserIds, MigrationWorkerCount, t.deadline)
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
	minSeconds, ok := t.task.Data["MinInterval"].(int)
	if !ok || minSeconds < 1 {
		minSeconds = int(DefaultMigrationAvailableAfterDurationMinimum.Seconds())
		t.task.Data["minInterval"] = minSeconds
	}
	maxSeconds, ok := t.task.Data["MaxInterval"].(int)
	if !ok || maxSeconds < minSeconds {
		maxSeconds = int(DefaultMigrationAvailableAfterDurationMaximum.Seconds())
		t.task.Data["maxInterval"] = maxSeconds
	}
	t.task.RepeatAvailableAfter(GenerateNextTime(minSeconds, maxSeconds))
}
