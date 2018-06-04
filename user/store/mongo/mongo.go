package mongo

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"github.com/tidepool-org/platform/user"
	"github.com/tidepool-org/platform/user/store"
)

func NewStore(cfg *Config, lgr log.Logger) (*Store, error) {
	if cfg == nil {
		return nil, errors.New("config is missing")
	}

	baseStore, err := storeStructuredMongo.NewStore(cfg.Config, lgr)
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
	*storeStructuredMongo.Store
	config *Config
}

func (s *Store) NewUsersSession() store.UsersSession {
	return &UsersSession{
		Session: s.Store.NewSession("users"),
		config:  s.config,
	}
}

type UsersSession struct {
	*storeStructuredMongo.Session
	config *Config
}

func (u *UsersSession) GetUserByID(ctx context.Context, userID string) (*user.User, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	}

	if u.IsClosed() {
		return nil, errors.New("session closed")
	}

	startTime := time.Now()

	users := []*user.User{}
	selector := bson.M{
		"userid": userID,
	}
	err := u.C().Find(selector).Limit(2).All(&users)

	loggerFields := log.Fields{"userId": userID, "duration": time.Since(startTime) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("GetUserByID")

	if err != nil {
		return nil, errors.Wrap(err, "unable to get user by id")
	}

	if usersCount := len(users); usersCount == 0 {
		return nil, nil
	} else if usersCount > 1 {
		log.LoggerFromContext(ctx).WithField("userId", userID).Warn("Multiple users found for user id")
	}

	user := users[0]

	if meta, ok := user.Private["meta"]; ok && meta.ID != "" {
		user.ProfileID = &meta.ID
	}

	return user, nil
}

func (u *UsersSession) DeleteUser(ctx context.Context, user *user.User) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if user == nil {
		return errors.New("user is missing")
	}
	if user.ID == "" {
		return errors.New("user id is missing")
	}

	if u.IsClosed() {
		return errors.New("session closed")
	}

	startTime := time.Now()

	user.DeletedTime = time.Now().Format(time.RFC3339)

	selector := bson.M{
		"userid": user.ID,
	}
	err := u.C().Update(selector, user)

	loggerFields := log.Fields{"userId": user.ID, "duration": time.Since(startTime) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DeleteUser")

	if err != nil {
		return errors.Wrap(err, "unable to delete user")
	}
	return nil
}

func (u *UsersSession) DestroyUserByID(ctx context.Context, userID string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userID == "" {
		return errors.New("user id is missing")
	}

	if u.IsClosed() {
		return errors.New("session closed")
	}

	startTime := time.Now()

	selector := bson.M{
		"userid": userID,
	}
	err := u.C().Remove(selector)

	loggerFields := log.Fields{"userId": userID, "duration": time.Since(startTime) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DestroyUserByID")

	if err != nil {
		return errors.Wrap(err, "unable to destroy user by id")
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
