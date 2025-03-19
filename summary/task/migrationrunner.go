package task

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/task"
)

const (
	DefaultMigrationAvailableAfterDurationMinimum = 5 * time.Minute
	DefaultMigrationAvailableAfterDurationMaximum = 5 * time.Minute
	MigrationTaskDurationMaximum                  = 4 * time.Minute
	DefaultMigrationWorkerBatchSize               = 500
	MigrationType                                 = "org.tidepool.summary.migrate.%s"
)

type MigrationRunner struct {
	authClient  AuthClient
	dataClient  DataClient
	summaryType string
}

func NewMigrationRunner(authClient AuthClient, dataClient DataClient, summaryType string) (*MigrationRunner, error) {
	if authClient == nil {
		return nil, errors.New("auth client is missing")
	}
	if dataClient == nil {
		return nil, errors.New("data client is missing")
	}
	if summaryType != "cgm" && summaryType != "bgm" && summaryType != "con" {
		return nil, errors.Newf("summary type \"%s\" not supported by migration runner", summaryType)
	}

	return &MigrationRunner{
		authClient:  authClient,
		dataClient:  dataClient,
		summaryType: summaryType,
	}, nil
}

func (r *MigrationRunner) AuthClient() AuthClient {
	return r.authClient
}

func (r *MigrationRunner) DataClient() DataClient {
	return r.dataClient
}

func (r *MigrationRunner) GetRunnerType() string {
	return fmt.Sprintf(MigrationType, r.summaryType)
}

func (r *MigrationRunner) SummaryType() string {
	return r.summaryType
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
	ctx = auth.NewContextWithServerSessionTokenProvider(ctx, r.AuthClient())
	if taskRunner, err := NewMigrationTaskRunner(r, tsk); err != nil {
		log.LoggerFromContext(ctx).WithError(err).Warn("Unable to create task runner")
	} else {
		taskRunner.Run(ctx)
	}
}

type MigrationTaskRunner struct {
	Provider
	task     *task.Task
	context  context.Context
	logger   log.Logger
	deadline time.Time
	config   Configuration
}

func NewMigrationTaskRunner(provider Provider, tsk *task.Task) (*MigrationTaskRunner, error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}
	if tsk == nil {
		return nil, errors.New("task is missing")
	}

	return &MigrationTaskRunner{
		Provider: provider,
		task:     tsk,
	}, nil
}

func (t *MigrationTaskRunner) Run(ctx context.Context) {
	t.context = ctx
	t.logger = log.LoggerFromContext(t.context)
	t.deadline = time.Now().Add(t.GetRunnerDurationMaximum())
	t.config = t.GetConfig()

	t.task.ClearError()
	if err := t.run(); err == nil {
		t.rescheduleTask()
	} else if !t.task.HasError() {
		t.rescheduleTaskWithResourceError(err)
	}
}

func (t *MigrationTaskRunner) GetConfig() Configuration {
	var config Configuration
	var valid bool
	if raw, ok := t.task.Data["config"]; ok {
		// this is abuse of marshal/unmarshal, this was done with interface{} target when loading the task,
		// but we require something more specific at this point
		bs, _ := bson.Marshal(raw)
		unmarshalError := bson.Unmarshal(bs, &config)
		if unmarshalError != nil {
			t.logger.WithField("unmarshalError", unmarshalError).Warn("Task configuration invalid, falling back to defaults.")
		} else {
			if configErr := ValidateConfig(config, true); configErr != nil {
				t.logger.WithField("validationError", configErr).Warn("Task configuration invalid, falling back to defaults.")
			} else {
				valid = true
			}
		}
	}

	if !valid {
		config = NewDefaultMigrationConfig()

		if t.task.Data == nil {
			t.task.Data = make(map[string]interface{})
		}
		t.task.Data["config"] = config
	}
	return config
}

func (t *MigrationTaskRunner) run() error {
	pagination := page.NewPagination()
	pagination.Size = *t.config.Batch
	typ := t.SummaryType()

	t.logger.Infof("Searching for User %s Summaries requiring Migration", typ)
	outdatedUserIds, err := t.DataClient().GetMigratableUserIDs(t.context, typ, pagination)
	if err != nil {
		return err
	}
	if len(outdatedUserIds) == 0 {
		t.logger.Infof("No %s Summaries requiring migrations found", typ)
		return nil
	}

	t.logger.Infof("Found batch of %d %s Summaries to Migrate", len(outdatedUserIds), typ)

	t.logger.Debugf("Starting User %s Summary Migration", typ)
	err = updateSummaries(t.context, t.DataClient(), typ, outdatedUserIds)
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
	t.task.RepeatAvailableAfter(GenerateNextTime(t.config.Interval))
}
