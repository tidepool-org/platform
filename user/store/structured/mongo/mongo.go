package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

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

func NewStore(params storeStructuredMongo.Params) (*Store, error) {
	store, err := storeStructuredMongo.NewStore(params)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: store,
	}, nil
}

func (s *Store) EnsureIndexes() error {
	repository := s.newUserRepository()
	return repository.EnsureIndexes()
}

func (s *Store) NewUserRepository() userStoreStructured.UserRepository {
	return s.newUserRepository()
}

func (s *Store) newUserRepository() *UserRepository {
	return &UserRepository{
		s.Store.GetRepository("users"),
	}
}

type UserRepository struct {
	*storeStructuredMongo.Repository
}

func (u *UserRepository) EnsureIndexes() error {
	// Indexes are created in `shoreline`
	return nil
}

func (u *UserRepository) Get(ctx context.Context, id string, condition *request.Condition) (*user.User, error) {
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

	now := time.Now()

	result, err := u.get(ctx, logger, id, condition, storeStructuredMongo.NotDeleted)
	if err != nil {
		return nil, err
	}

	logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("Get")
	return result, nil
}

func (u *UserRepository) Delete(ctx context.Context, id string, condition *request.Condition) (bool, error) {
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
	changeInfo, err := u.UpdateMany(ctx, query, u.ConstructUpdate(set, unset))
	if err != nil {
		logger.WithError(err).Error("Unable to delete user")
		return false, errors.Wrap(err, "unable to delete user")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("Delete")
	return changeInfo.ModifiedCount > 0, nil
}

func (u *UserRepository) Destroy(ctx context.Context, id string, condition *request.Condition) (bool, error) {
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

	now := time.Now()

	query := bson.M{
		"userid": id,
	}
	if condition.Revision != nil {
		query["revision"] = *condition.Revision
	}
	changeInfo, err := u.DeleteMany(ctx, query)
	if err != nil {
		logger.WithError(err).Error("Unable to destroy user")
		return false, errors.Wrap(err, "unable to destroy user")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("Destroy")
	return changeInfo.DeletedCount > 0, nil
}

func (u *UserRepository) get(ctx context.Context, logger log.Logger, id string, condition *request.Condition, queryModifiers ...storeStructuredMongo.QueryModifier) (*user.User, error) {
	logger = logger.WithFields(log.Fields{"id": id, "condition": condition})

	var result *user.User
	query := bson.M{
		"userid": id,
	}
	if condition != nil && condition.Revision != nil {
		query["revision"] = *condition.Revision
	}
	query = storeStructuredMongo.ModifyQuery(query, queryModifiers...)
	err := u.FindOne(ctx, query).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		logger.WithError(err).Error("Unable to get user")
		return nil, errors.Wrap(err, "unable to get user")
	}

	if result.Revision == nil {
		result.Revision = pointer.FromInt(0)
	}

	return result, nil
}
