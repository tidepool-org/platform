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
	"encoding/json"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/app"
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

func (s *Store) NewSession(logger log.Logger) (store.Session, error) {
	baseSession, err := s.Store.NewSession(logger)
	if err != nil {
		return nil, err
	}

	return &Session{
		Session: baseSession,
	}, nil
}

type Session struct {
	*mongo.Session
}

func (s *Session) GetProfileByID(profileID string) (*profile.Profile, error) {
	if profileID == "" {
		return nil, app.Error("mongo", "profile id is missing")
	}

	if s.IsClosed() {
		return nil, app.Error("mongo", "session closed")
	}

	startTime := time.Now()

	profiles := []*profile.Profile{}
	selector := bson.M{
		"_id": profileID,
	}
	err := s.C().Find(selector).Limit(2).All(&profiles)

	loggerFields := log.Fields{"profileID": profileID, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("GetProfileByID")

	if err != nil {
		return nil, app.ExtError(err, "mongo", "unable to get profile by id")
	}

	if profilesCount := len(profiles); profilesCount == 0 {
		return nil, nil
	} else if profilesCount > 1 {
		s.Logger().WithField("profileID", profileID).Warn("Multiple profiles found for profile id")
	}

	profile := profiles[0]

	// NOTE: Partial implementation; only what is needed at present
	if profile.Value != "" {
		var value map[string]interface{}
		if err = json.Unmarshal([]byte(profile.Value), &value); err != nil {
			s.Logger().WithField("profileID", profileID).WithError(err).Warn("Unable to unmarshal profile value")
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
		return app.Error("mongo", "profile id is missing")
	}

	if s.IsClosed() {
		return app.Error("mongo", "session closed")
	}

	startTime := time.Now()

	selector := bson.M{
		"_id": profileID,
	}
	removeInfo, err := s.C().RemoveAll(selector)

	loggerFields := log.Fields{"profileID": profileID, "remove-info": removeInfo, "duration": time.Since(startTime) / time.Microsecond}
	s.Logger().WithFields(loggerFields).WithError(err).Debug("DestroyProfileByID")

	if err != nil {
		return app.ExtError(err, "mongo", "unable to destroy profile by id")
	}
	return nil
}
