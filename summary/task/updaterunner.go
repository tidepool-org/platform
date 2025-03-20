package task

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"

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
	UpdateType                                 = "org.tidepool.summary.update.%s"
	IterLimit                                  = 3
)

type UpdateRunner struct {
	authClient  auth.Client
	dataClient  dataClient.Client
	summaryType string
}

func NewUpdateRunner(authClient auth.Client, dataClient dataClient.Client, summaryType string) (*UpdateRunner, error) {
	if authClient == nil {
		return nil, errors.New("auth client is missing")
	}
	if dataClient == nil {
		return nil, errors.New("data client is missing")
	}
	if summaryType != "cgm" && summaryType != "bgm" && summaryType != "con" {
		return nil, errors.Newf("summary type \"%s\" not supported by update runner", summaryType)
	}

	return &UpdateRunner{
		authClient:  authClient,
		dataClient:  dataClient,
		summaryType: summaryType,
	}, nil
}

func (r *UpdateRunner) GetRunnerType() string {
	return fmt.Sprintf(UpdateType, r.summaryType)
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
	config   Configuration
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
	t.config = t.GetConfig()

	t.task.ClearError()
	if err := t.run(); err == nil {
		t.rescheduleTask()
	} else if !t.task.HasError() {
		t.rescheduleTaskWithResourceError(err)
	}
}

func (t *UpdateTaskRunner) GetConfig() Configuration {
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
		config = NewDefaultUpdateConfig()

		if t.task.Data == nil {
			t.task.Data = make(map[string]interface{})
		}
		t.task.Data["config"] = config
	}
	return config
}

func (t *UpdateTaskRunner) run() error {
	pagination := page.NewPagination()
	pagination.Size = *t.config.Batch
	typ := t.summaryType
	targetTime := time.Now().UTC().Add(-1 * time.Minute)

	t.logger.Debug("Starting User CGM Summary Update")
	iCount := 0
	// this loop is a bit odd looking, we are iterating until the end of the previous loop is past the target
	// this avoids a round trip, and allows the default time zero value to work as a starter
	for {
		if iCount >= IterLimit {
			t.logger.Warn("Exiting CGM batch loop early, too many iterations")
			break
		}

		t.logger.Infof("Searching for User %s Summaries requiring Update", typ)
		outdatedCGM, err := t.dataClient.GetOutdatedUserIDs(t.context, "cgm", pagination)
		if err != nil {
			return err
		}
		if len(outdatedCGM.UserIds) == 0 {
			t.logger.Infof("No %s Summaries requiring updates found", typ)
			return nil
		}

		t.logger.Infof("Found batch of %d %s Summaries to Migrate", len(outdatedCGM.UserIds), typ)

		err = updateSummaries(t.context, t.dataClient, typ, outdatedCGM.UserIds)
		if err != nil {
			return err
		}

		if outdatedCGM.End.After(targetTime) || outdatedCGM.End.IsZero() {
			// we are sufficiently caught up
			break
		}

		iCount++
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
	t.task.RepeatAvailableAfter(GenerateNextTime(t.config.Interval))
}
