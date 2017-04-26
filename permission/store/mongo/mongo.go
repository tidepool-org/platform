package mongo

import (
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/permission/store"
	"github.com/tidepool-org/platform/store/mongo"
)

func New(logger log.Logger, config *Config) (*Store, error) {
	if config == nil {
		return nil, app.Error("mongo", "config is missing")
	}

	baseStore, err := mongo.New(logger, config.Config)
	if err != nil {
		return nil, err
	}

	config = config.Clone()
	if err = config.Validate(); err != nil {
		return nil, app.ExtError(err, "mongo", "config is invalid")
	}

	return &Store{
		Store:  baseStore,
		config: config,
	}, nil
}

type Store struct {
	*mongo.Store
	config *Config
}

func (s *Store) NewSession(logger log.Logger) (store.Session, error) {
	baseSession, err := s.Store.NewSession(logger)
	if err != nil {
		return nil, err
	}

	return &Session{
		Session: baseSession,
		config:  s.config,
	}, nil
}

type Session struct {
	*mongo.Session
	config *Config
}

func (s *Session) DestroyPermissionsForUserByID(userID string) error {
	if userID == "" {
		return app.Error("mongo", "user id is missing")
	}

	if s.IsClosed() {
		return app.Error("mongo", "session closed")
	}

	startTime := time.Now()

	groupID, err := permission.GroupIDFromUserID(userID, s.config.Secret)
	if err != nil {
		return app.ExtError(err, "mongo", "unable to determine group id from user id")
	}

	selector := bson.M{
		"$or": []bson.M{
			{"groupId": groupID},
			{"userId": userID},
		},
	}
	removeInfo, err := s.C().RemoveAll(selector)

	loggerFields := log.Fields{"userId": userID, "removeInfo": removeInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("DestroyPermissionsForUserByID")

	if err != nil {
		return app.ExtError(err, "mongo", "unable to destroy permissions for user by id")
	}
	return nil
}
