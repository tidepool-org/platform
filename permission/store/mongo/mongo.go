package mongo

import (
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/permission/store"
	"github.com/tidepool-org/platform/store/mongo"
)

func New(logger log.Logger, config *Config) (*Store, error) {
	if config == nil {
		return nil, errors.New("mongo", "config is missing")
	}

	baseStore, err := mongo.New(logger, config.Config)
	if err != nil {
		return nil, err
	}

	if err = config.Validate(); err != nil {
		return nil, errors.Wrap(err, "mongo", "config is invalid")
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

func (s *Store) NewPermissionsSession(logger log.Logger) store.PermissionsSession {
	return &PermissionsSession{
		Session: s.Store.NewSession(logger, "perms"),
		config:  s.config,
	}
}

type PermissionsSession struct {
	*mongo.Session
	config *Config
}

func (p *PermissionsSession) DestroyPermissionsForUserByID(userID string) error {
	if userID == "" {
		return errors.New("mongo", "user id is missing")
	}

	if p.IsClosed() {
		return errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	groupID, err := permission.GroupIDFromUserID(userID, p.config.Secret)
	if err != nil {
		return errors.Wrap(err, "mongo", "unable to determine group id from user id")
	}

	selector := bson.M{
		"$or": []bson.M{
			{"groupId": groupID},
			{"userId": userID},
		},
	}
	removeInfo, err := p.C().RemoveAll(selector)

	loggerFields := log.Fields{"userId": userID, "removeInfo": removeInfo, "duration": time.Since(startTime) / time.Microsecond}
	p.Logger().WithFields(loggerFields).WithError(err).Debug("DestroyPermissionsForUserByID")

	if err != nil {
		return errors.Wrap(err, "mongo", "unable to destroy permissions for user by id")
	}
	return nil
}
