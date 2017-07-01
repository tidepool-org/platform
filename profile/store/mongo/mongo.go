package mongo

import (
	"encoding/json"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/profile"
	"github.com/tidepool-org/platform/profile/store"
	"github.com/tidepool-org/platform/store/mongo"
)

func New(logger log.Logger, config *mongo.Config) (*Store, error) {
	baseStore, err := mongo.New(logger, config)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: baseStore,
	}, nil
}

type Store struct {
	*mongo.Store
}

func (s *Store) NewSession(logger log.Logger) store.Session {
	return &Session{
		Session: s.Store.NewSession(logger),
	}
}

type Session struct {
	*mongo.Session
}

func (s *Session) GetProfileByID(profileID string) (*profile.Profile, error) {
	if profileID == "" {
		return nil, errors.New("mongo", "profile id is missing")
	}

	if s.IsClosed() {
		return nil, errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	profiles := []*profile.Profile{}
	selector := bson.M{
		"_id": profileID,
	}
	err := s.C().Find(selector).Limit(2).All(&profiles)

	loggerFields := log.Fields{"profileId": profileID, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("GetProfileByID")

	if err != nil {
		return nil, errors.Wrap(err, "mongo", "unable to get profile by id")
	}

	if profilesCount := len(profiles); profilesCount == 0 {
		return nil, nil
	} else if profilesCount > 1 {
		s.Logger().WithField("profileId", profileID).Warn("Multiple profiles found for profile id")
	}

	profile := profiles[0]

	// NOTE: Partial implementation; only what is needed at present
	if profile.Value != "" {
		var value map[string]interface{}
		if err = json.Unmarshal([]byte(profile.Value), &value); err != nil {
			s.Logger().WithField("profileId", profileID).WithError(err).Warn("Unable to unmarshal profile value")
		} else {
			if profileMap, profileMapOk := value["profile"].(map[string]interface{}); profileMapOk {
				if fullName, fullNameOk := profileMap["fullName"].(string); fullNameOk {
					profile.FullName = &fullName
				}
			}
		}
	}

	return profile, nil
}

func (s *Session) DestroyProfileByID(profileID string) error {
	if profileID == "" {
		return errors.New("mongo", "profile id is missing")
	}

	if s.IsClosed() {
		return errors.New("mongo", "session closed")
	}

	startTime := time.Now()

	selector := bson.M{
		"_id": profileID,
	}
	removeInfo, err := s.C().RemoveAll(selector)

	loggerFields := log.Fields{"profileId": profileID, "removeInfo": removeInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("DestroyProfileByID")

	if err != nil {
		return errors.Wrap(err, "mongo", "unable to destroy profile by id")
	}
	return nil
}
