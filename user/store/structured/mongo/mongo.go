package mongo

import (
	"context"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
	userStoreStructured "github.com/tidepool-org/platform/user/store/structured"
)

type Store struct {
	*storeStructuredMongo.Store
}

func NewStore(config *storeStructuredMongo.Config, logger log.Logger) (*Store, error) {
	store, err := storeStructuredMongo.NewStore(config, logger)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: store,
	}, nil
}

func (s *Store) EnsureIndexes() error {
	session := s.newSession()
	defer session.Close()
	return session.EnsureIndexes()
}

func (s *Store) NewSession() userStoreStructured.Session {
	return s.newSession()
}

func (s *Store) newSession() *Session {
	return &Session{
		Session: s.Store.NewSession("users"),
	}
}

type Session struct {
	*storeStructuredMongo.Session
}

func (s *Session) EnsureIndexes() error {
	return s.EnsureAllIndexes([]mgo.Index{
		{Key: []string{"userid"}, Background: true, Unique: true},
	})
}

func (s *Session) Get(ctx context.Context, id string, condition *request.Condition) (*user.User, error) {
	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"id": id, "condition": condition})

	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	} else if !user.IsValidID(id) {
		return nil, errors.New("id is invalid")
	}
	if condition == nil {
		condition = request.NewCondition()
	} else if err := structureValidator.New().Validate(condition); err != nil {
		return nil, errors.Wrap(err, "condition is invalid")
	}

	if s.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()

	result, err := s.get(logger, id, condition, storeStructuredMongo.NotDeleted)
	if err != nil {
		return nil, err
	}

	logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("Get")
	return result, nil
}

func (s *Session) Delete(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"id": id, "condition": condition})

	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if id == "" {
		return false, errors.New("id is missing")
	} else if !user.IsValidID(id) {
		return false, errors.New("id is invalid")
	}
	if condition == nil {
		condition = request.NewCondition()
	} else if err := structureValidator.New().Validate(condition); err != nil {
		return false, errors.Wrap(err, "condition is invalid")
	}

	if s.IsClosed() {
		return false, errors.New("session closed")
	}

	now := time.Now()

	query := bson.M{
		"userid": id,
	}
	if condition.Revision != nil {
		query["revision"] = *condition.Revision
	}
	set := bson.M{
		"modifiedTime": now.Truncate(time.Second),
		"deletedTime":  now.Truncate(time.Second),
	}
	unset := bson.M{}
	changeInfo, err := s.C().UpdateAll(query, s.ConstructUpdate(set, unset))
	if err != nil {
		logger.WithError(err).Error("Unable to delete user")
		return false, errors.Wrap(err, "unable to delete user")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("Delete")
	return changeInfo.Updated > 0, nil
}

func (s *Session) Destroy(ctx context.Context, id string, condition *request.Condition) (bool, error) {
	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"id": id, "condition": condition})

	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if id == "" {
		return false, errors.New("id is missing")
	} else if !user.IsValidID(id) {
		return false, errors.New("id is invalid")
	}
	if condition == nil {
		condition = request.NewCondition()
	} else if err := structureValidator.New().Validate(condition); err != nil {
		return false, errors.Wrap(err, "condition is invalid")
	}

	if s.IsClosed() {
		return false, errors.New("session closed")
	}

	now := time.Now()

	query := bson.M{
		"userid": id,
	}
	if condition.Revision != nil {
		query["revision"] = *condition.Revision
	}
	changeInfo, err := s.C().RemoveAll(query)
	if err != nil {
		logger.WithError(err).Error("Unable to destroy user")
		return false, errors.Wrap(err, "unable to destroy user")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("Destroy")
	return changeInfo.Removed > 0, nil
}

func (s *Session) get(logger log.Logger, id string, condition *request.Condition, queryModifiers ...storeStructuredMongo.QueryModifier) (*user.User, error) {
	logger = logger.WithFields(log.Fields{"id": id, "condition": condition})

	results := user.UserArray{}
	query := bson.M{
		"userid": id,
	}
	if condition != nil && condition.Revision != nil {
		query["revision"] = *condition.Revision
	}
	query = storeStructuredMongo.ModifyQuery(query, queryModifiers...)
	err := s.C().Find(query).Limit(2).All(&results)
	if err != nil {
		logger.WithError(err).Error("Unable to get user")
		return nil, errors.Wrap(err, "unable to get user")
	}

	var result *user.User
	switch len(results) {
	case 0:
		return nil, nil
	case 1:
		result = results[0]
	default:
		logger.Error("Multiple users found")
		result = results[0]
	}

	if result.Revision == nil {
		result.Revision = pointer.FromInt(0)
	}

	return result, nil
}
