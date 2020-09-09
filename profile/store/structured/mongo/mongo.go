package mongo

import (
	"context"
	"encoding/json"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/profile"
	profileStoreStructured "github.com/tidepool-org/platform/profile/store/structured"
	"github.com/tidepool-org/platform/request"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
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

func (s *Store) NewSession() profileStoreStructured.Session {
	return s.newSession()
}

func (s *Store) newSession() *Session {
	return &Session{
		Session: s.Store.NewSession("seagull"),
	}
}

type Session struct {
	*storeStructuredMongo.Session
}

func (s *Session) Get(ctx context.Context, userID string, condition *request.Condition) (*profile.Profile, error) {
	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"userId": userID, "condition": condition})

	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return nil, errors.New("user id is invalid")
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

	result, err := s.get(logger, userID, condition, storeStructuredMongo.NotDeleted)
	if err != nil {
		return nil, err
	}

	logger.WithField("duration", time.Since(now)/time.Microsecond).Debug("Get")
	return result, nil
}

func (s *Session) Delete(ctx context.Context, userID string, condition *request.Condition) (bool, error) {
	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"userId": userID, "condition": condition})

	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if userID == "" {
		return false, errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return false, errors.New("user id is invalid")
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
		"userId": userID,
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
		logger.WithError(err).Error("Unable to delete profile")
		return false, errors.Wrap(err, "unable to delete profile")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("Delete")
	return changeInfo.Updated > 0, nil
}

func (s *Session) Destroy(ctx context.Context, userID string, condition *request.Condition) (bool, error) {
	ctx, logger := log.ContextAndLoggerWithFields(ctx, log.Fields{"userId": userID, "condition": condition})

	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if userID == "" {
		return false, errors.New("user id is missing")
	} else if !user.IsValidID(userID) {
		return false, errors.New("user id is invalid")
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
		"userId": userID,
	}
	if condition.Revision != nil {
		query["revision"] = *condition.Revision
	}
	changeInfo, err := s.C().RemoveAll(query)
	if err != nil {
		logger.WithError(err).Error("Unable to destroy profile")
		return false, errors.Wrap(err, "unable to destroy profile")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("Destroy")
	return changeInfo.Removed > 0, nil
}

func (s *Session) get(logger log.Logger, userID string, condition *request.Condition, queryModifiers ...storeStructuredMongo.QueryModifier) (*profile.Profile, error) {
	logger = logger.WithFields(log.Fields{"userId": userID, "condition": condition})

	results := profile.ProfileArray{}
	query := bson.M{
		"userId": userID,
	}
	if condition != nil && condition.Revision != nil {
		query["revision"] = *condition.Revision
	}
	query = storeStructuredMongo.ModifyQuery(query, queryModifiers...)
	err := s.C().Find(query).Limit(2).All(&results)
	if err != nil {
		logger.WithError(err).Error("Unable to get profile")
		return nil, errors.Wrap(err, "unable to get profile")
	}

	var result *profile.Profile
	switch len(results) {
	case 0:
		return nil, nil
	case 1:
		result = results[0]
	default:
		logger.Error("Multiple profiles found")
		result = results[0]
	}

	if result.Revision == nil {
		result.Revision = pointer.FromInt(0)
	}

	// NOTE: Partial implementation; only what is needed at present
	if result.Value != nil && *result.Value != "" {
		var value map[string]interface{}
		if err = json.Unmarshal([]byte(*result.Value), &value); err != nil {
			logger.WithError(err).Error("Unable to unmarshal profile value")
		} else {
			if profileMap, profileMapOk := value["profile"].(map[string]interface{}); profileMapOk {
				if fullName, fullNameOk := profileMap["fullName"].(string); fullNameOk {
					result.FullName = &fullName
				}
			}
		}
	}
	logger.Debug("Get")
	return result, nil
}
