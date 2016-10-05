package mongo

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"crypto/sha1"
	"encoding/hex"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store/mongo"
	"github.com/tidepool-org/platform/user"
	"github.com/tidepool-org/platform/user/store"
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

func (s *Session) GetUserByID(userID string) (*user.User, error) {
	if userID == "" {
		return nil, app.Error("mongo", "user id is missing")
	}

	if s.IsClosed() {
		return nil, app.Error("mongo", "session closed")
	}

	startTime := time.Now()

	users := []*user.User{}
	selector := bson.M{
		"userid": userID,
	}
	err := s.C().Find(selector).Limit(2).All(&users)

	loggerFields := log.Fields{"userID": userID, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("GetUserByID")

	if err != nil {
		return nil, app.ExtError(err, "mongo", "unable to get user by id")
	}

	if usersCount := len(users); usersCount == 0 {
		return nil, nil
	} else if usersCount > 1 {
		s.Logger().WithField("userID", userID).Warn("Multiple users found for user id")
	}

	user := users[0]

	if meta, ok := user.Private["meta"]; ok && meta.ID != "" {
		user.ProfileID = &meta.ID
	}

	return user, nil
}

func (s *Session) DeleteUser(user *user.User) error {
	if user == nil {
		return app.Error("mongo", "user is missing")
	}
	if user.ID == "" {
		return app.Error("mongo", "user id is missing")
	}

	if s.IsClosed() {
		return app.Error("mongo", "session closed")
	}

	startTime := time.Now()

	user.DeletedTime = s.Timestamp()
	user.DeletedUserID = s.AgentUserID()

	selector := bson.M{
		"userid": user.ID,
	}
	err := s.C().Update(selector, user)

	loggerFields := log.Fields{"userID": user.ID, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("DeleteUser")

	if err != nil {
		return app.ExtError(err, "mongo", "unable to delete user")
	}
	return nil
}

func (s *Session) DestroyUserByID(userID string) error {
	if userID == "" {
		return app.Error("mongo", "user id is missing")
	}

	if s.IsClosed() {
		return app.Error("mongo", "session closed")
	}

	startTime := time.Now()

	selector := bson.M{
		"userid": userID,
	}
	err := s.C().Remove(selector)

	loggerFields := log.Fields{"userID": userID, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("DestroyUserByID")

	if err != nil {
		return app.ExtError(err, "mongo", "unable to destroy user by id")
	}
	return nil
}

// TODO: This really isn't the right place for this, but we shouldn't be using a
// password hash algorithm with an external salt, but instead something like bcrypt

// TODO: We should use a constant-time password matching algorithm

func (s *Session) PasswordMatches(user *user.User, password string) bool {
	return user.PasswordHash == s.HashPassword(user.ID, password)
}

// TODO: Do away with external salt and use hash algorithm with internal salt (eg. bcrypt/scrypt)

func (s *Session) HashPassword(userID string, password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))
	hash.Write([]byte(s.config.PasswordSalt))
	hash.Write([]byte(userID))
	return hex.EncodeToString(hash.Sum(nil))
}
