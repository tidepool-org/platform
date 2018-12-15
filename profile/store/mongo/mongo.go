package mongo

import (
	"context"
	"encoding/json"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/profile"
	"github.com/tidepool-org/platform/profile/store"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

func NewStore(cfg *storeStructuredMongo.Config, lgr log.Logger) (*Store, error) {
	baseStore, err := storeStructuredMongo.NewStore(cfg, lgr)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: baseStore,
	}, nil
}

type Store struct {
	*storeStructuredMongo.Store
}

func (s *Store) NewProfilesSession() store.ProfilesSession {
	return &ProfilesSession{
		Session: s.Store.NewSession("seagull"),
	}
}

type ProfilesSession struct {
	*storeStructuredMongo.Session
}

func (p *ProfilesSession) GetProfileByID(ctx context.Context, profileID string) (*profile.Profile, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if profileID == "" {
		return nil, errors.New("profile id is missing")
	}

	if p.IsClosed() {
		return nil, errors.New("session closed")
	}

	now := time.Now()

	profiles := []*profile.Profile{}
	selector := bson.M{
		"_id": profileID,
	}
	err := p.C().Find(selector).Limit(2).All(&profiles)

	loggerFields := log.Fields{"profileId": profileID, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("GetProfileByID")

	if err != nil {
		return nil, errors.Wrap(err, "unable to get profile by id")
	}

	if profilesCount := len(profiles); profilesCount == 0 {
		return nil, nil
	} else if profilesCount > 1 {
		log.LoggerFromContext(ctx).WithField("profileId", profileID).Warn("Multiple profiles found for profile id")
	}

	profile := profiles[0]

	// NOTE: Partial implementation; only what is needed at present
	if profile.Value != "" {
		var value map[string]interface{}
		if err = json.Unmarshal([]byte(profile.Value), &value); err != nil {
			log.LoggerFromContext(ctx).WithField("profileId", profileID).WithError(err).Warn("Unable to unmarshal profile value")
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

func (p *ProfilesSession) DestroyProfileByID(ctx context.Context, profileID string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if profileID == "" {
		return errors.New("profile id is missing")
	}

	if p.IsClosed() {
		return errors.New("session closed")
	}

	now := time.Now()

	selector := bson.M{
		"_id": profileID,
	}
	removeInfo, err := p.C().RemoveAll(selector)

	loggerFields := log.Fields{"profileId": profileID, "removeInfo": removeInfo, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DestroyProfileByID")

	if err != nil {
		return errors.Wrap(err, "unable to destroy profile by id")
	}
	return nil
}
