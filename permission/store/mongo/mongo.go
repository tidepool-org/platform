package mongo

import (
	"context"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/permission/store"
	"github.com/tidepool-org/platform/store/mongo"
)

func NewStore(cfg *Config, lgr log.Logger) (*Store, error) {
	if cfg == nil {
		return nil, errors.New("config is missing")
	}

	baseStore, err := mongo.NewStore(cfg.Config, lgr)
	if err != nil {
		return nil, err
	}

	if err = cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}

	return &Store{
		Store:  baseStore,
		config: cfg,
	}, nil
}

type Store struct {
	*mongo.Store
	config *Config
}

func (s *Store) NewPermissionsSession() store.PermissionsSession {
	return &PermissionsSession{
		Session: s.Store.NewSession("perms"),
		config:  s.config,
	}
}

type PermissionsSession struct {
	*mongo.Session
	config *Config
}

func (p *PermissionsSession) DestroyPermissionsForUserByID(ctx context.Context, userID string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userID == "" {
		return errors.New("user id is missing")
	}

	if p.IsClosed() {
		return errors.New("session closed")
	}

	startTime := time.Now()

	groupID, err := permission.GroupIDFromUserID(userID, p.config.Secret)
	if err != nil {
		return errors.Wrap(err, "unable to determine group id from user id")
	}

	selector := bson.M{
		"$or": []bson.M{
			{"groupId": groupID},
			{"userId": userID},
		},
	}
	removeInfo, err := p.C().RemoveAll(selector)

	loggerFields := log.Fields{"userId": userID, "removeInfo": removeInfo, "duration": time.Since(startTime) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DestroyPermissionsForUserByID")

	if err != nil {
		return errors.Wrap(err, "unable to destroy permissions for user by id")
	}
	return nil
}
