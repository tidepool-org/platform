package summary

import (
	"context"
	"math/rand"
	//"sort"
	"time"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/data"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/version"
)

const (
	AvailableAfterDurationMaximum = 50 * time.Minute
	AvailableAfterDurationMinimum = 30 * time.Minute
	TaskDurationMaximum           = 30 * time.Minute
)

type Runner struct {
	logger           log.Logger
	versionReporter  version.Reporter
	authClient       auth.Client
	dataClient       dataClient.Client
}

func NewRunner(logger log.Logger, versionReporter version.Reporter, authClient auth.Client, dataClient dataClient.Client) (*Runner, error) {
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

	return &Runner{
		logger:           logger,
		versionReporter:  versionReporter,
		authClient:       authClient,
		dataClient:       dataClient,
	}, nil
}

func (r *Runner) CanRunTask(tsk *task.Task) bool {
	return tsk != nil && tsk.Type == Type
}

func (r *Runner) Run(ctx context.Context, tsk *task.Task) {
	now := time.Now()

	ctx = log.NewContextWithLogger(ctx, r.logger)

    tsk.ClearError()

    if serverSessionToken, sErr := r.authClient.ServerSessionToken(); sErr != nil {
        tsk.AppendError(errors.Wrap(sErr, "unable to get server session token"))
    } else {
        ctx = auth.NewContextWithServerSessionToken(ctx, serverSessionToken)

        if taskRunner, tErr := NewTaskRunner(r, tsk); tErr != nil {
            tsk.AppendError(errors.Wrap(tErr, "unable to create task runner"))
        } else if tErr = taskRunner.Run(ctx); tErr != nil {
            tsk.AppendError(errors.Wrap(tErr, "unable to run task runner"))
        }
    }

	if !tsk.IsFailed() {
		tsk.RepeatAvailableAfter(AvailableAfterDurationMinimum + time.Duration(rand.Int63n(int64(AvailableAfterDurationMaximum-AvailableAfterDurationMinimum+1))))
	}

	if taskDuration := time.Since(now); taskDuration > TaskDurationMaximum {
		r.logger.WithField("taskDuration", taskDuration.Truncate(time.Millisecond).Seconds()).Warn("Task duration exceeds maximum")
	}
}

type TaskRunner struct {
	*Runner
	task             *task.Task
	context          context.Context
	validator        structure.Validator
	dataSet          *data.DataSet
	dataSetPreloaded bool
}

func NewTaskRunner(rnnr *Runner, tsk *task.Task) (*TaskRunner, error) {
	if rnnr == nil {
		return nil, errors.New("runner is missing")
	}
	if tsk == nil {
		return nil, errors.New("task is missing")
	}

	return &TaskRunner{
		Runner: rnnr,
		task:   tsk,
	}, nil
}


func (t *TaskRunner) Run(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is missing")
	}

	if len(t.task.Data) == 0 {
		t.task.SetFailed()
		return errors.New("data is missing")
	}

	t.context = ctx
	t.validator = structureValidator.New()

	if err := t.update(); err != nil {
		if request.IsErrorUnauthenticated(errors.Cause(err)) {
			t.task.SetFailed()
		}
		return err
	}
	return nil
}

func (t *TaskRunner) update() error {
	// find needed users

	//for user in users
    // somehow batch of 20 threads
// 		if err := t.updateUser(id); err != nil {
// 			return err
// 		}
	return nil
}

func (t *TaskRunner) updateUser(id string) error {

	return nil
}

func (t *TaskRunner) getPendingUsers() error {

    return nil
}
