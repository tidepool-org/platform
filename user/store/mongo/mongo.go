package mongo

import (
	"crypto/sha1"
	"encoding/hex"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store/mongo"
	"github.com/tidepool-org/platform/user"
	"github.com/tidepool-org/platform/user/store"
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

func (s *Store) NewUsersSession(logger log.Logger) store.UsersSession {
	return &UsersSession{
		Session: s.Store.NewSession(logger, "users"),
		config:  s.config,
	}
}

type UsersSession struct {
	*mongo.Session
	config *Config
}

func (u *UsersSession) GetUserByID(userID string) (*user.User, error) {
	if userID == "" {
		return nil, errors.New("mongo", "user id is missing")
	}

	if u.IsClosed() {
		return nil, errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	users := []*user.User{}
	selector := bson.M{
		"userid": userID,
	}
	err := u.C().Find(selector).Limit(2).All(&users)

	loggerFields := log.Fields{"userId": userID, "duration": time.Since(startTime) / time.Microsecond}
	u.Logger().WithFields(loggerFields).WithError(err).Debug("GetUserByID")

	if err != nil {
		return nil, errors.Wrap(err, "mongo", "unable to get user by id")
	}

	if usersCount := len(users); usersCount == 0 {
		return nil, nil
	} else if usersCount > 1 {
		u.Logger().WithField("userId", userID).Warn("Multiple users found for user id")
	}

	user := users[0]

	if meta, ok := user.Private["meta"]; ok && meta.ID != "" {
		user.ProfileID = &meta.ID
	}

	return user, nil
}

func (u *UsersSession) DeleteUser(user *user.User) error {
	if user == nil {
		return errors.New("mongo", "user is missing")
	}
	if user.ID == "" {
		return errors.New("mongo", "user id is missing")
	}

	if u.IsClosed() {
		return errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	user.DeletedTime = u.Timestamp()
	user.DeletedUserID = u.AgentUserID()

	selector := bson.M{
		"userid": user.ID,
	}
	err := u.C().Update(selector, user)

	loggerFields := log.Fields{"userId": user.ID, "duration": time.Since(startTime) / time.Microsecond}
	u.Logger().WithFields(loggerFields).WithError(err).Debug("DeleteUser")

	if err != nil {
		return errors.Wrap(err, "mongo", "unable to delete user")
	}
	return nil
}

func (u *UsersSession) DestroyUserByID(userID string) error {
	if userID == "" {
		return errors.New("mongo", "user id is missing")
	}

	if u.IsClosed() {
		return errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	selector := bson.M{
		"userid": userID,
	}
	err := u.C().Remove(selector)

	loggerFields := log.Fields{"userId": userID, "duration": time.Since(startTime) / time.Microsecond}
	u.Logger().WithFields(loggerFields).WithError(err).Debug("DestroyUserByID")

	if err != nil {
		return errors.Wrap(err, "mongo", "unable to destroy user by id")
	}
	return nil
}

// TODO: This really isn't the right place for this, but we shouldn't be using a
// password hash algorithm with an external salt, but instead something like bcrypt

// TODO: We should use a constant-time password matching algorithm

func (u *UsersSession) PasswordMatches(user *user.User, password string) bool {
	return user.PasswordHash == u.HashPassword(user.ID, password)
}

// TODO: Do away with external salt and use hash algorithm with internal salt (eg. bcrypt/scrypt)

func (u *UsersSession) HashPassword(userID string, password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))
	hash.Write([]byte(u.config.PasswordSalt))
	hash.Write([]byte(userID))
	return hex.EncodeToString(hash.Sum(nil))
}
