package mongo

import (
	"context"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/store/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/task/store"
)

type Store struct {
	*mongo.Store
}

func New(cfg *mongo.Config, lgr log.Logger) (*Store, error) {
	str, err := mongo.New(cfg, lgr)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: str,
	}, nil
}

func (s *Store) NewTaskSession() store.TaskSession {
	return &TaskSession{
		Session: s.Store.NewSession("tasks"),
	}
}

func (s *Store) EnsureIndexes() error {
	ssn := s.NewTaskSession()
	defer ssn.Close()
	return ssn.EnsureIndexes()
}

type TaskSession struct {
	*mongo.Session
}

func (t *TaskSession) EnsureIndexes() error {
	return t.EnsureAllIndexes([]mgo.Index{
		{Key: []string{"id"}, Unique: true, Background: true},
		{Key: []string{"name"}, Unique: true, Sparse: true, Background: true},
		{Key: []string{"priority"}, Background: true},
		{Key: []string{"availableTime"}, Background: true},
		{Key: []string{"expirationTime"}, Background: true},
		{Key: []string{"state"}, Background: true},
	})
}

func (t *TaskSession) ListTasks(ctx context.Context, filter *task.TaskFilter, pagination *page.Pagination) (task.Tasks, error) {
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

	if t.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"filter": filter, "pagination": pagination})

	tsks := task.Tasks{}
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
	err := t.C().Find(selector).Sort("-createdTime").Skip(pagination.Page * pagination.Size).Limit(pagination.Size).All(&tsks)
	logger.WithFields(log.Fields{"count": len(tsks), "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("ListTasks")
	if err != nil {
		return nil, errors.Wrap(err, "unable to list tasks")
	}

	if tsks == nil {
		tsks = task.Tasks{}
	}

	return tsks, nil
}

func (t *TaskSession) CreateTask(ctx context.Context, create *task.TaskCreate) (*task.Task, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	tsk, err := task.NewTask(create)
	if err != nil {
		return nil, err
	} else if err = structureValidator.New().Validate(tsk); err != nil {
		return nil, errors.Wrap(err, "task is invalid")
	}

	if t.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"create": create})

	err = t.C().Insert(tsk)
	logger.WithFields(log.Fields{"id": tsk.ID, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("CreateTask")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create task")
	}

	return tsk, nil
}

func (t *TaskSession) GetTask(ctx context.Context, id string) (*task.Task, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	}

	if t.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("id", id)

	tsks := task.Tasks{}
	err := t.C().Find(bson.M{"id": id}).Limit(2).All(&tsks)
	logger.WithField("duration", time.Since(now)/time.Microsecond).WithError(err).Debug("GetTask")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get task")
	}

	switch count := len(tsks); count {
	case 0:
		return nil, nil
	case 1:
		return tsks[0], nil
	default:
		logger.WithField("count", count).Warnf("Multiple tasks found for id %q", id)
		return tsks[0], nil
	}
}

func (t *TaskSession) UpdateTask(ctx context.Context, id string, update *task.TaskUpdate) (*task.Task, error) {
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

	if t.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"id": id, "update": update})

	set := bson.M{
		"modifiedTime": now.Truncate(time.Second),
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
	changeInfo, err := t.C().UpdateAll(bson.M{"id": id}, t.ConstructUpdate(set, bson.M{}))
	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("UpdateTask")
	if err != nil {
		return nil, errors.Wrap(err, "unable to update task")
	}

	return t.GetTask(ctx, id)
}

func (t *TaskSession) DeleteTask(ctx context.Context, id string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if id == "" {
		return errors.New("id is missing")
	}

	if t.IsClosed() {
		return errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("id", id)

	changeInfo, err := t.C().RemoveAll(bson.M{"id": id})
	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("DeleteTask")
	if err != nil {
		return errors.Wrap(err, "unable to delete task")
	}

	return nil
}

func (t *TaskSession) UpdateFromState(ctx context.Context, tsk *task.Task, state string) (*task.Task, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if tsk == nil {
		return nil, errors.New("task is missing")
	}

	if t.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"id": tsk.ID, "state": state})

	tsk.ModifiedTime = pointer.Time(now)

	selector := bson.M{
		"id":    tsk.ID,
		"state": state,
	}
	err := t.C().Update(selector, tsk)
	logger.WithField("duration", time.Since(now)/time.Microsecond).WithError(err).Debug("UpdateFromState")
	if err != nil {
		return nil, errors.Wrap(err, "unable to update from state")
	}

	return tsk, nil
}

func (t *TaskSession) IteratePending(ctx context.Context) store.TaskIterator {
	if ctx == nil {
		return &TaskIterator{err: errors.New("context is missing")}
	}

	if t.IsClosed() {
		return &TaskIterator{err: errors.New("session closed")}
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx)

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

	iterator := t.C().Find(selector).Sort("-priority").Iter()
	err := iterator.Err()

	logger.WithError(err).WithField("duration", time.Since(now)/time.Microsecond).Debug("IteratePending")

	return &TaskIterator{
		iterator: iterator,
		err:      err,
	}
}

type TaskIterator struct {
	iterator *mgo.Iter
	err      error
}

func (t *TaskIterator) Next(tsk *task.Task) bool {
	if tsk == nil {
		t.setError(errors.New("task is missing"))
	}

	if t.err != nil {
		return false
	}

	return t.iterator.Next(tsk)
}

func (t *TaskIterator) Close() error {
	if t.iterator != nil {
		if err := t.iterator.Close(); err != nil {
			t.setError(errors.Wrap(err, "unable to close iterator"))
		}
	}

	return t.Error()
}

func (t *TaskIterator) Error() error {
	if t.iterator != nil && t.err == nil {
		if err := t.iterator.Err(); err != nil {
			t.setError(errors.Wrap(err, "iterator failure"))
		}
	}

	return t.err
}

func (t *TaskIterator) setError(err error) {
	if t.err == nil {
		t.err = err
	}
}
