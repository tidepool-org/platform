package mongo

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/ehr/reconcile"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/task/store"
	"github.com/tidepool-org/platform/task/summary"
)

var (
	TasksStateTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tidepool_task_tasks_state_total",
		Help: "The total number of tasks sorted by state and type",
	}, []string{"state", "type"})
)

const MaxTaskCreationDuration = 30 * time.Second

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

func (s *Store) WithTypeFilter(typeFilter string) store.Store {
	return &Store{
		Store:      s.Store,
		typeFilter: &typeFilter,
	}
}

func (s *Store) NewTaskRepository() store.TaskRepository {
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
		repository.EnsureSummaryBackfillTask,
		repository.EnsureSummaryMigrationTask,
		repository.EnsureEHRReconcileTask,
		repository.EnsureCarePartnerTask,
	}

	for _, f := range fs {
		if err := f(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) EnsureSummaryMigrationTask() error {
	ctx, cancel := context.WithTimeout(context.Background(), MaxTaskCreationDuration)
	defer cancel()

	repository := s.TaskRepository()
	return repository.EnsureSummaryMigrationTask(ctx)
}

func (s *Store) EnsureEHRReconcileTask() error {
	ctx, cancel := context.WithTimeout(context.Background(), MaxTaskCreationDuration)
	defer cancel()

	repository := s.TaskRepository()
	return repository.EnsureEHRReconcileTask(ctx)
}

type TaskRepository struct {
	*storeStructuredMongo.Repository
	typeFilter *string
}

func (t *TaskRepository) EnsureIndexes() error {
	// Repositories operation only a subset of the tasks shouldn't invoke this method
	if t.typeFilter != nil {
		return errors.New("calling EnsureIndexes() on a partitioned repository is not allowed")
	}

	return t.CreateAllIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "id", Value: 1}},
			Options: options.Index().
				SetUnique(true).
				SetBackground(true),
		},
		{
			Keys: bson.D{{Key: "name", Value: 1}},
			Options: options.Index().
				SetUnique(true).
				SetSparse(true).
				SetBackground(true),
		},
		{
			Keys: bson.D{{Key: "priority", Value: 1}},
			Options: options.Index().
				SetBackground(true),
		},
		{
			Keys: bson.D{{Key: "availableTime", Value: 1}},
			Options: options.Index().
				SetBackground(true),
		},
		{
			Keys: bson.D{{Key: "expirationTime", Value: 1}},
			Options: options.Index().
				SetBackground(true),
		},
		{
			Keys: bson.D{{Key: "state", Value: 1}},
			Options: options.Index().
				SetBackground(true),
		},
	})
}

func (t *TaskRepository) EnsureSummaryUpdateTask(ctx context.Context) error {
	create := summary.NewDefaultUpdateTaskCreate()
	return t.ensureTask(ctx, create)
}

func (t *TaskRepository) EnsureSummaryBackfillTask(ctx context.Context) error {
	create := summary.NewDefaultBackfillTaskCreate()
	return t.ensureTask(ctx, create)
}

func (t *TaskRepository) EnsureSummaryMigrationTask(ctx context.Context) error {
	create := summary.NewDefaultMigrationTaskCreate()
	return t.ensureTask(ctx, create)
}

func (t *TaskRepository) EnsureEHRReconcileTask(ctx context.Context) error {
	create := reconcile.NewTaskCreate()
	return t.ensureTask(ctx, create)
}

func (t *TaskRepository) EnsureCarePartnerTask(ctx context.Context) error {
	create := task.NewCarePartnerTaskCreate()
	return t.ensureTask(ctx, create)
}

func (t *TaskRepository) ensureTask(ctx context.Context, create *task.TaskCreate) error {
	tsk, err := task.NewTask(ctx, create)
	if err != nil {
		return err
	} else if err = structureValidator.New(log.LoggerFromContext(ctx)).Validate(tsk); err != nil {
		return errors.Wrap(err, "task is invalid")
	}
	if err := t.assertType(t.typeFilter, &tsk.Type); err != nil {
		return err
	}

	upsert := true
	after := options.After
	opts := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}

	res := t.FindOneAndUpdate(ctx,
		bson.M{"name": tsk.Name},
		bson.M{"$setOnInsert": tsk},
		&opts,
	)

	if res.Err() != nil && !errors.Is(res.Err(), mongo.ErrNoDocuments) {
		return errors.Wrap(res.Err(), "unable to create task")
	}

	TasksStateTotal.WithLabelValues(task.TaskStatePending, create.Type).Inc()

	return res.Err()
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

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"filter": filter, "pagination": pagination})

	tasks := task.Tasks{}
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
	opts := storeStructuredMongo.FindWithPagination(pagination).
		SetSort(bson.M{"createdTime": -1})
	cursor, err := t.Find(ctx, selector, opts)
	logger.WithFields(log.Fields{"count": len(tasks), "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("ListTasks")
	if err != nil {
		return nil, errors.Wrap(err, "unable to list tasks")
	}

	if err = cursor.All(ctx, &tasks); err != nil {
		return nil, errors.Wrap(err, "unable to decode tasks")
	}

	if tasks == nil {
		tasks = task.Tasks{}
	}

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
	if err := t.assertType(t.typeFilter, &tsk.Type); err != nil {
		return nil, err
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"create": create})

	_, err = t.InsertOne(ctx, tsk)
	logger.WithFields(log.Fields{"id": tsk.ID, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("CreateTask")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create task")
	}

	TasksStateTotal.WithLabelValues(task.TaskStatePending, create.Type).Inc()
	return tsk, nil
}

func (t *TaskRepository) GetTask(ctx context.Context, id string) (*task.Task, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("id", id)

	var task *task.Task

	selector := bson.M{"id": id}
	if t.typeFilter != nil {
		selector["type"] = t.typeFilter
	}

	err := t.FindOne(ctx, selector).Decode(&task)
	logger.WithField("duration", time.Since(now)/time.Microsecond).WithError(err).Debug("GetTask")

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to get task")
	}

	return task, nil
}

func (t *TaskRepository) UpdateTask(ctx context.Context, id string, update *task.TaskUpdate) (*task.Task, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	}
	if update == nil {
		return nil, errors.New("update is missing")
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(update); err != nil {
		return nil, errors.Wrap(err, "update is invalid")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"id": id, "update": update})

	set := bson.M{
		"modifiedTime": now,
	}
	if update.Priority != nil {
		set["priority"] = *update.Priority
	}
	if update.Data != nil {
		set["data"] = *update.Data
	}
	if update.AvailableTime != nil {
		set["availableTime"] = *update.AvailableTime
	}
	if update.ExpirationTime != nil {
		set["expirationTime"] = *update.ExpirationTime
	}

	selector := bson.M{"id": id}
	if t.typeFilter != nil {
		selector["type"] = t.typeFilter
	}

	changeInfo, err := t.UpdateMany(ctx, selector, t.ConstructUpdate(set, bson.M{}))
	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("UpdateTask")
	if err != nil {
		return nil, errors.Wrap(err, "unable to update task")
	}

	return t.GetTask(ctx, id)
}

func (t *TaskRepository) DeleteTask(ctx context.Context, id string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if id == "" {
		return errors.New("id is missing")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("id", id)

	selector := bson.M{"id": id}
	if t.typeFilter != nil {
		selector["type"] = t.typeFilter
	}

	changeInfo, err := t.DeleteOne(ctx, selector)
	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("DeleteTask")
	if err != nil {
		return errors.Wrap(err, "unable to delete task")
	}

	return nil
}

// TODO: Consider using an "update only specific fields" approach, as above

func (t *TaskRepository) UpdateFromState(ctx context.Context, tsk *task.Task, state string) (*task.Task, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if tsk == nil {
		return nil, errors.New("task is missing")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"id": tsk.ID, "state": state})

	tsk.ModifiedTime = pointer.FromTime(now.Truncate(time.Millisecond))

	selector := bson.M{
		"id":    tsk.ID,
		"state": state,
	}
	if t.typeFilter != nil {
		selector["type"] = t.typeFilter
	}
	result, err := t.ReplaceOne(ctx, selector, tsk)
	logger.WithField("duration", time.Since(now)/time.Microsecond).WithError(err).Debug("UpdateFromState")
	if err != nil {
		return nil, errors.Wrap(err, "unable to update from state")
	}
	if result.ModifiedCount != 1 {
		return nil, task.AlreadyClaimedTask
	}

	TasksStateTotal.WithLabelValues(tsk.State, tsk.Type).Inc()
	return tsk, nil
}

func (t *TaskRepository) UnstickTasks(ctx context.Context) (int64, error) {
	selector := bson.M{
		"state":        task.TaskStateRunning,
		"deadlineTime": bson.M{"$lt": time.Now()},
	}
	if t.typeFilter != nil {
		selector["type"] = t.typeFilter
	}

	update := bson.M{
		"$set":   bson.M{"state": task.TaskStatePending},
		"$unset": bson.M{"deadlineTime": ""},
	}

	result, err := t.UpdateMany(ctx, selector, update)
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, err
}

func (t *TaskRepository) IteratePending(ctx context.Context) (*mongo.Cursor, error) {
	now := time.Now()

	selector := bson.M{
		"state": task.TaskStatePending,
		"$and": []bson.M{
			{
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
			},
			{
				"$or": []bson.M{
					{
						"expirationTime": bson.M{
							"$exists": false,
						},
					},
					{
						"expirationTime": bson.M{
							"$gt": now,
						},
					},
				},
			},
		},
	}
	if t.typeFilter != nil {
		selector["type"] = t.typeFilter
	}
	opts := options.Find().SetSort(bson.M{"priority": -1})
	return t.Find(ctx, selector, opts)
}

// assertType return an error if the expected type doesn't match the actual type
func (t *TaskRepository) assertType(expected *string, actual *string) error {
	if expected != nil && actual != nil && *expected != *actual {
		return errors.Newf("expected task type %s but got %s", *expected, *actual)
	}
	return nil
}
