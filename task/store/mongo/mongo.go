package mongo

import (
	"context"
	"slices"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	ehrReconcile "github.com/tidepool-org/platform/ehr/reconcile"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	storeStructured "github.com/tidepool-org/platform/store/structured"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	summaryTask "github.com/tidepool-org/platform/summary/task"
	"github.com/tidepool-org/platform/task"
	taskStore "github.com/tidepool-org/platform/task/store"
)

const (
	MaxTaskCreationDuration = 30 * time.Second

	// TransitionTimeout bounds a task state-transition write (start or stop) that must
	// complete regardless of the caller's context being canceled (e.g. during shutdown).
	TransitionTimeout = 10 * time.Second
)

type Store struct {
	*storeStructuredMongo.Store
	typeFilter *string
}

func NewStore(config *storeStructuredMongo.Config) (*Store, error) {
	str, err := storeStructuredMongo.NewStore(config)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: str,
	}, nil
}

func (s *Store) WithTypeFilter(typeFilter string) taskStore.Store {
	return &Store{
		Store:      s.Store,
		typeFilter: &typeFilter,
	}
}

func (s *Store) NewTaskRepository() taskStore.TaskRepository {
	repo := s.TaskRepository()
	repo.typeFilter = s.typeFilter
	return repo
}

func (s *Store) TaskRepository() *TaskRepository {
	return &TaskRepository{
		Repository: s.Store.GetRepository("tasks"),
	}
}

func (s *Store) EnsureIndexes() error {
	repository := s.TaskRepository()
	return repository.EnsureIndexes()
}

func (s *Store) EnsureDefaultTasks() error {
	ctx, cancel := context.WithTimeout(context.Background(), MaxTaskCreationDuration)
	defer cancel()

	repository := s.TaskRepository()
	fs := []func(context.Context) error{
		repository.EnsureSummaryUpdateTask,
		repository.EnsureSummaryMigrationTask,
		repository.EnsureEHRReconcileTask,
	}

	for _, f := range fs {
		if err := f(ctx); err != nil {
			return err
		}
	}

	return nil
}

type TaskRepository struct {
	*storeStructuredMongo.Repository
	typeFilter *string
}

func (t *TaskRepository) EnsureIndexes() error {
	// Repositories operating on a subset of the tasks shouldn't invoke this method
	if t.typeFilter != nil {
		return errors.New("calling EnsureIndexes() on a partitioned repository is not allowed")
	}

	return t.CreateAllIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "id", Value: 1}},
			Options: options.Index().
				SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "name", Value: 1}},
			Options: options.Index().
				SetUnique(true).
				SetSparse(true),
		},
		{
			Keys: bson.D{{Key: "availableTime", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "state", Value: 1}},
		},
		{
			// Used by IteratePending; type equality, then availableTime for range and sort; partial
			// on pending since that is the only state it queries.
			Keys: bson.D{{Key: "type", Value: 1}, {Key: "availableTime", Value: 1}},
			Options: options.Index().
				SetPartialFilterExpression(bson.D{{Key: "state", Value: task.TaskStatePending}}),
		},
		{
			// Used by UnstickTasks; type equality, then deadlineTime for range and sort; partial
			// on running since that is the only state it queries.
			Keys: bson.D{{Key: "type", Value: 1}, {Key: "deadlineTime", Value: 1}},
			Options: options.Index().
				SetPartialFilterExpression(bson.D{{Key: "state", Value: task.TaskStateRunning}}),
		},
	})
}

func (t *TaskRepository) EnsureSummaryUpdateTask(ctx context.Context) error {
	for _, summaryType := range summaryTask.SummaryTypes {
		create := summaryTask.NewDefaultUpdateTaskCreate(summaryType)
		err := t.ensureTask(ctx, create)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *TaskRepository) EnsureSummaryMigrationTask(ctx context.Context) error {
	for _, summaryType := range summaryTask.SummaryTypes {
		create := summaryTask.NewDefaultMigrationTaskCreate(summaryType)
		err := t.ensureTask(ctx, create)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *TaskRepository) EnsureEHRReconcileTask(ctx context.Context) error {
	create := ehrReconcile.NewTaskCreate()
	return t.ensureTask(ctx, create)
}

func (t *TaskRepository) ensureTask(ctx context.Context, create *task.TaskCreate) error {
	tsk, err := task.NewTask(ctx, create)
	if err != nil {
		return err
	} else if err = structureValidator.New(log.LoggerFromContext(ctx)).Validate(tsk); err != nil {
		return errors.Wrap(err, "task is invalid")
	}

	if result, err := t.UpdateOne(ctx, bson.M{"name": tsk.Name}, bson.M{"$setOnInsert": tsk}, options.Update().SetUpsert(true)); err != nil {
		return errors.Wrap(err, "unable to create task")
	} else if result.UpsertedCount > 0 {
		TypeStateTotal.WithLabelValues(create.Type, task.TaskStatePending).Inc()
	}

	return nil
}

func (t *TaskRepository) ListTasks(ctx context.Context, filter *task.TaskFilter, pagination *page.Pagination) (task.Tasks, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if filter == nil {
		filter = task.NewTaskFilter()
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}
	if err := t.assertType(t.typeFilter, filter.Type); err != nil {
		return nil, err
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"filter": filter, "pagination": pagination})

	now := time.Now().UTC()
	defer func() { logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("ListTasks") }()

	selector := bson.M{}

	if filter.Name != nil {
		selector["name"] = *filter.Name
	}
	if filter.Type != nil {
		selector["type"] = *filter.Type
	}
	if filter.State != nil {
		selector["state"] = *filter.State
	}

	if t.typeFilter != nil {
		selector["type"] = *t.typeFilter
	}

	opts := storeStructuredMongo.FindWithPagination(pagination).SetSort(bson.M{"createdTime": -1})
	cursor, err := t.Find(ctx, selector, opts)
	if err != nil {
		logger = logger.WithError(err)
		return nil, errors.Wrap(err, "unable to list tasks")
	}

	tasks := task.Tasks{}
	if err = cursor.All(ctx, &tasks); err != nil {
		logger = logger.WithError(err)
		return nil, errors.Wrap(err, "unable to decode tasks")
	}

	if tasks == nil {
		tasks = task.Tasks{}
	}

	taskIds := make([]string, len(tasks))
	for index, tsk := range tasks {
		taskIds[index] = tsk.ID
	}
	logger = logger.WithField("taskIds", taskIds)

	return tasks, nil
}

func (t *TaskRepository) CreateTask(ctx context.Context, create *task.TaskCreate) (*task.Task, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	tsk, err := task.NewTask(ctx, create)
	if err != nil {
		return nil, err
	} else if err = structureValidator.New(log.LoggerFromContext(ctx)).Validate(tsk); err != nil {
		return nil, errors.Wrap(err, "task is invalid")
	}
	if err = t.assertType(t.typeFilter, &tsk.Type); err != nil {
		return nil, err
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"create": create})

	now := time.Now().UTC()
	defer func() { logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("CreateTask") }()

	if _, err = t.InsertOne(ctx, tsk); err != nil {
		logger = logger.WithError(err)
		return nil, errors.Wrap(err, "unable to create task")
	}

	logger = logger.WithField("task", tsk.LogFields())

	TypeStateTotal.WithLabelValues(create.Type, task.TaskStatePending).Inc()
	return tsk, nil
}

func (t *TaskRepository) GetTask(ctx context.Context, id string, condition *storeStructured.Condition) (*task.Task, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	}
	if condition == nil {
		condition = &storeStructured.Condition{}
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(condition); err != nil {
		return nil, errors.Wrap(err, "condition is invalid")
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"id": id, "condition": condition})

	now := time.Now().UTC()
	defer func() { logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("GetTask") }()

	tsk := &task.Task{}
	err := t.FindOne(ctx, t.selector(id, condition)).Decode(tsk)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	} else if err != nil {
		logger = logger.WithError(err)
		return nil, errors.Wrap(err, "unable to get task")
	}

	return tsk, nil
}

func (t *TaskRepository) UpdateTask(ctx context.Context, id string, condition *storeStructured.Condition, update *task.TaskUpdate) (*task.Task, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	}
	if condition == nil {
		condition = &storeStructured.Condition{}
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(condition); err != nil {
		return nil, errors.Wrap(err, "condition is invalid")
	}
	if update == nil {
		return nil, errors.New("update is missing")
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(update); err != nil {
		return nil, errors.Wrap(err, "update is invalid")
	}

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"id": id, "condition": condition, "update": update})

	now := time.Now().UTC()
	defer func() { logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("UpdateTask") }()

	set, unset := t.parseUpdate(update)
	set["modifiedTime"] = now

	updatedTask := &task.Task{}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err := t.FindOneAndUpdate(ctx, t.selector(id, condition), t.ConstructUpdate(set, unset), opts).Decode(updatedTask)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	} else if err != nil {
		logger = logger.WithError(err)
		return nil, errors.Wrap(err, "unable to update task")
	}

	return updatedTask, nil
}

func (t *TaskRepository) DeleteTask(ctx context.Context, id string, condition *storeStructured.Condition) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if id == "" {
		return errors.New("id is missing")
	}
	if condition == nil {
		condition = &storeStructured.Condition{}
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(condition); err != nil {
		return errors.Wrap(err, "condition is invalid")
	}

	logger := log.LoggerFromContext(ctx).WithField("id", id)

	now := time.Now().UTC()
	defer func() { logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("DeleteTask") }()

	if changeInfo, err := t.DeleteOne(ctx, t.selector(id, condition)); err != nil {
		logger = logger.WithError(err)
		return errors.Wrap(err, "unable to delete task")
	} else {
		logger = logger.WithField("changeInfo", changeInfo)
	}

	return nil
}

func (t *TaskRepository) StartTask(ctx context.Context, id string, revision int, deadline time.Duration) (*task.Task, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	} else if !task.IsValidID(id) {
		return nil, errors.New("id is invalid")
	}
	if deadline <= 0 {
		return nil, errors.New("deadline is invalid")
	}

	// Add a timeout, but ignore cancel from the parent context so the claim write completes
	// and its outcome is known. A write abandoned mid-flight (e.g. on shutdown) can still
	// commit in the database, leaving the task running with no reliable way to revert it.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), TransitionTimeout)
	defer cancel()

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"id": id, "revision": revision, "deadline": deadline.String()})

	now := time.Now().UTC()
	defer func() { logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("StartTask") }()

	set := bson.M{
		"state":        task.TaskStateRunning,
		"runTime":      now,
		"deadlineTime": now.Add(deadline),
		"modifiedTime": now,
		"stateLock":    newStateLock(),
	}
	unset := bson.M{
		"availableTime": 1,
		"duration":      1,
	}

	selector := t.selector(id, storeStructured.NewConditionWithRevision(&revision))
	selector["state"] = task.TaskStatePending

	tsk := &task.Task{}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err := t.FindOneAndUpdate(ctx, selector, t.ConstructUpdate(set, unset), opts).Decode(tsk)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	} else if err != nil {
		logger = logger.WithError(err)
		return nil, errors.Wrap(err, "unable to start task")
	}

	TypeStateTotal.WithLabelValues(tsk.Type, task.TaskStateRunning).Inc()
	return tsk, nil
}

// Will only timeout after 10 seconds even if parent context is canceled.
func (t *TaskRepository) StopTask(ctx context.Context, id string, revision int, stateLock *string, state string, duration *time.Duration, update *task.TaskUpdate) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if id == "" {
		return errors.New("id is missing")
	} else if !task.IsValidID(id) {
		return errors.New("id is invalid")
	}
	if stateLock == nil {
		return errors.New("state lock is missing")
	} else if *stateLock == "" {
		return errors.New("state lock is invalid")
	}
	if state == "" {
		return errors.New("state is missing")
	} else if !slices.Contains(task.TaskStates(), state) {
		return errors.New("state is invalid")
	}
	if duration != nil && *duration < 0 {
		return errors.New("duration is invalid")
	}
	if update != nil {
		if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(update); err != nil {
			return errors.Wrap(err, "update is invalid")
		}
	}

	// Add a timeout, but ignore cancel from parent context to ensure we stop task even exiting
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), TransitionTimeout)
	defer cancel()

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"id": id, "revision": revision, "stateLock": stateLock, "state": state, "duration": duration, "update": update})

	now := time.Now().UTC()
	defer func() { logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("StopTask") }()

	set, unset := t.parseUpdate(update)
	set["modifiedTime"] = now
	set["state"] = state
	unset["deadlineTime"] = 1
	unset["stateLock"] = 1
	if duration != nil {
		set["duration"] = duration.Truncate(time.Millisecond).Seconds()
	} else {
		// A nil duration means no run actually happened (e.g. a claimed task whose dispatch
		// was reverted during shutdown), so also clear the run time recorded by StartTask,
		// keeping the runTime/duration pair describing only the last actual run.
		unset["duration"] = 1
		unset["runTime"] = 1
	}

	selector := t.selector(id, nil)
	selector["state"] = task.TaskStateRunning
	selector["stateLock"] = stateLock

	tsk := &task.Task{}
	err := t.FindOneAndUpdate(ctx, selector, t.ConstructUpdate(set, unset)).Decode(tsk)
	if errors.Is(err, mongo.ErrNoDocuments) {
		// The compare-and-swap missed: no running task matched the expected state lock
		// (it was concurrently modified, unstuck, or deleted since it started).
		// The state transition is dropped; the deadline and unstick mechanism will
		// recover the task if it was left running. This is logged and counted so lost
		// completions are observable rather than silently swallowed.
		logger.Error("Unable to stop task; no running task matched the expected condition")
		TypeLostCompletionTotal.WithLabelValues(pointer.Default(t.typeFilter, "")).Inc()
		return nil
	} else if err != nil {
		logger = logger.WithError(err)
		return errors.Wrap(err, "unable to stop task")
	}

	// If the on-disk task revision does not match the runner task revision, then either:
	//   - the runner did not follow the Runner contract; i.e. the runner updated the task during
	//     run, but it did not use the updated task, or,
	//   - the task was concurrently modified outside of the runner while the task was running.
	if tsk.Revision != revision {
		logger.WithField("revision", log.Fields{"expected": revision, "actual": tsk.Revision}).Warn("Database task revision does not match running task revision; Runner contract broken or concurrent update")
		TypeRevisionMismatchTotal.WithLabelValues(tsk.Type).Inc()
	}

	TypeStateTotal.WithLabelValues(tsk.Type, state).Inc()
	return nil
}

func (t *TaskRepository) UnstickTasks(ctx context.Context) ([]string, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	logger := log.LoggerFromContext(ctx)

	now := time.Now().UTC()
	defer func() { logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("UnstickTasks") }()

	findSelector := bson.M{
		"state":        task.TaskStateRunning,
		"deadlineTime": bson.M{"$lt": now},
	}
	if t.typeFilter != nil {
		findSelector["type"] = *t.typeFilter
	}

	opts := options.Find().SetSort(bson.M{"deadlineTime": 1})
	cursor, err := t.Find(ctx, findSelector, opts)
	if err != nil {
		logger = logger.WithError(err)
		return nil, errors.Wrap(err, "unable to list tasks")
	}
	defer storeStructuredMongo.CloseCursor(ctx, cursor)

	var ids []string
	for cursor.Next(ctx) {
		tsk := &task.Task{}
		if err = cursor.Decode(tsk); err != nil {
			logger = logger.WithError(err)
			log.LoggerFromContext(ctx).WithError(err).Error("Unable to decode task")
			continue
		}

		// The state clause is defensive: a matching deadline time already implies the same
		// running claim (every stop clears the deadline time and any re-claim records a
		// strictly later one), but including it keeps the invariant local to this update.
		updateSelector := bson.M{
			"id":           tsk.ID,
			"state":        task.TaskStateRunning,
			"deadlineTime": tsk.DeadlineTime,
		}
		set := bson.M{
			"state":         task.TaskStatePending,
			"availableTime": now,
			"modifiedTime":  now,
		}
		unset := bson.M{
			"deadlineTime": 1,
			"stateLock":    1,
		}
		if result, updateErr := t.UpdateOne(ctx, updateSelector, t.ConstructUpdate(set, unset)); updateErr != nil {
			logger = logger.WithError(updateErr)
			return ids, updateErr
		} else if result.ModifiedCount > 0 {
			ids = append(ids, tsk.ID)
		}
	}

	logger = logger.WithField("taskIds", ids)

	if err = cursor.Err(); err != nil {
		logger = logger.WithError(err)
		return ids, err // Still want to return the ids of tasks that were successfully unstuck
	}

	return ids, nil
}

func (t *TaskRepository) IteratePending(ctx context.Context) (*mongo.Cursor, error) {
	now := time.Now().UTC()

	selector := bson.M{
		"state": task.TaskStatePending,
		"$or": []bson.M{
			{
				"availableTime": bson.M{
					"$exists": false,
				},
			},
			{
				"availableTime": bson.M{
					"$lte": now,
				},
			},
		},
	}
	if t.typeFilter != nil {
		selector["type"] = *t.typeFilter
	}
	opts := options.Find().SetSort(bson.D{{Key: "availableTime", Value: 1}})
	return t.Find(ctx, selector, opts)
}

func (t *TaskRepository) selector(id string, condition *storeStructured.Condition) bson.M {
	selector := bson.M{"id": id}
	if condition != nil {
		if condition.Revision != nil {
			if *condition.Revision == 0 {
				selector["revision"] = bson.M{"$in": bson.A{0, nil}}
			} else {
				selector["revision"] = *condition.Revision
			}
		}
	}
	if t.typeFilter != nil {
		selector["type"] = *t.typeFilter
	}
	return selector
}

func (t *TaskRepository) parseUpdate(update *task.TaskUpdate) (bson.M, bson.M) {
	set := bson.M{}
	unset := bson.M{}

	if update != nil {
		if update.Data != nil {
			if *update.Data != nil {
				set["data"] = *update.Data
			} else {
				unset["data"] = true
			}
		}
		if update.AvailableTime != nil {
			set["availableTime"] = *update.AvailableTime
		}
		if update.Error != nil {
			if update.Error.Error != nil {
				set["error"] = *update.Error
			} else {
				unset["error"] = true
			}
		}
	}

	return set, unset
}

// assertType return an error if the expected type doesn't match the actual type
func (t *TaskRepository) assertType(expected *string, actual *string) error {
	if expected != nil && actual != nil && *expected != *actual {
		return errors.Newf("expected task type %s but got %s", *expected, *actual)
	}
	return nil
}

func newStateLock() string {
	return id.Must(id.New(16))
}

var (
	// TypeStateTotal counts the total number of tasks run, sorted by type and state.
	TypeStateTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tidepool_task_type_state_total",
		Help: "The total number of tasks run, sorted by type and state",
	}, []string{"type", "state"})

	// TypeLostCompletionTotal counts task completions dropped because the compare-and-swap in
	// StopTask missed (task state lock was concurrently modified, task unstuck, or task deleted). The task
	// is recovered by the deadline/unstick mechanism, but the intended terminal state is lost. The
	// type label is populated only when the repository is type-filtered (as it is for each queue in
	// a MultiQueue); otherwise it is empty.
	TypeLostCompletionTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tidepool_task_type_lost_completion_total",
		Help: "The total number of task completions dropped because the state-lock compare-and-swap missed, sorted by type",
	}, []string{"type"})

	// TypeRevisionMismatchTotal counts task completions where the task revision does not match
	// the task revision in the database. This only occurs if the task runner does not follow the Runner
	// contract when updating a task during run, or if there is a concurrent modification outside of
	// the expected Runner behavior.
	TypeRevisionMismatchTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tidepool_task_type_revision_mismatch_total",
		Help: "The total number of task revisions that do not match the task revision in the database, sorted by type",
	}, []string{"type"})
)
