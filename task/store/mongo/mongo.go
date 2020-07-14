package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/task/store"
)

type Store struct {
	*storeStructuredMongo.Store
}

func NewStore(params storeStructuredMongo.Params) (*Store, error) {
	str, err := storeStructuredMongo.NewStore(params)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: str,
	}, nil
}

func (s *Store) NewTaskRepository() store.TaskRepository {
	return s.TaskRepository()
}

func (s *Store) TaskRepository() *TaskRepository {
	return &TaskRepository{
		s.Store.GetRepository("tasks"),
	}
}

func (s *Store) EnsureIndexes() error {
	ssn := s.TaskRepository()
	return ssn.EnsureIndexes()
}

type TaskRepository struct {
	*storeStructuredMongo.Repository
}

func (t *TaskRepository) EnsureIndexes() error {
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

func (t *TaskRepository) ListTasks(ctx context.Context, filter *task.TaskFilter, pagination *page.Pagination) (task.Tasks, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if filter == nil {
		filter = task.NewTaskFilter()
	} else if err := structureValidator.New().Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New().Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
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

	tsk, err := task.NewTask(create)
	if err != nil {
		return nil, err
	} else if err = structureValidator.New().Validate(tsk); err != nil {
		return nil, errors.Wrap(err, "task is invalid")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"create": create})

	_, err = t.InsertOne(ctx, tsk)
	logger.WithFields(log.Fields{"id": tsk.ID, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("CreateTask")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create task")
	}

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

	tasks := task.Tasks{}
	opts := options.Find().SetLimit(2)
	cursor, err := t.Find(ctx, bson.M{"id": id}, opts)
	logger.WithField("duration", time.Since(now)/time.Microsecond).WithError(err).Debug("GetTask")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get task")
	}

	if err = cursor.All(ctx, &tasks); err != nil {
		return nil, errors.Wrap(err, "unable to decode task")
	}

	switch count := len(tasks); count {
	case 0:
		return nil, nil
	case 1:
		return tasks[0], nil
	default:
		logger.WithField("count", count).Warnf("Multiple tasks found for id %q", id)
		return tasks[0], nil
	}
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
	} else if err := structureValidator.New().Validate(update); err != nil {
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
	changeInfo, err := t.UpdateMany(ctx, bson.M{"id": id}, t.ConstructUpdate(set, bson.M{}))
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

	changeInfo, err := t.DeleteMany(ctx, bson.M{"id": id})
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
	_, err := t.UpdateOne(ctx, selector, tsk)
	logger.WithField("duration", time.Since(now)/time.Microsecond).WithError(err).Debug("UpdateFromState")
	if err != nil {
		return nil, errors.Wrap(err, "unable to update from state")
	}

	return tsk, nil
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

	opts := options.Find().SetSort(bson.M{"priority": -1})
	return t.Find(ctx, selector, opts)
}
