package main

import (
	"encoding/json"
	"os"

	"github.com/urfave/cli"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	mongoMigration "github.com/tidepool-org/platform/migration/mongo"
	"github.com/tidepool-org/platform/store/mongo"
)

func main() {
	os.Exit(application.Run(NewMigration()))
}

type Migration struct {
	*mongoMigration.Migration
}

func NewMigration() (*Migration, error) {
	migration, err := mongoMigration.NewMigration("TIDEPOOL")
	if err != nil {
		return nil, err
	}

	return &Migration{
		Migration: migration,
	}, nil
}

func (m *Migration) Initialize() error {
	if err := m.Migration.Initialize(); err != nil {
		return err
	}

	m.CLI().Usage = "Migrate all device data to add user id derived from group id"
	m.CLI().Description = "Migrate all device data to add the '_userId' field derived from the '_groupId' field." +
		"\n\n   One or more warnings will be reported if partially created accounts or invalid data are found." +
		"\n\n   This migration is idempotent." +
		"\n\n   NOTE: This migration MUST be executed immediately AFTER upgrading Jellyfish to v0.12.1, Tide Whisperer to v0.9.1, or Platform to v0.1.0."
	m.CLI().Authors = []cli.Author{
		{
			Name:  "Darin Krauss",
			Email: "darin@tidepool.org",
		},
	}

	m.CLI().Action = func(context *cli.Context) error {
		if !m.ParseContext(context) {
			return nil
		}
		return m.execute()
	}

	return nil
}

func (m *Migration) execute() error {
	metaIDToUserIDMap, err := m.buildMetaIDToUserIDMap()
	if err != nil {
		return errors.Wrap(err, "unable to build meta id to user id map")
	}

	groupIDToUserIDMap, err := m.buildGroupIDToUserIDMap(metaIDToUserIDMap)
	if err != nil {
		return errors.Wrap(err, "unable to build group id to user id map")
	}

	err = m.migrateGroupIDToUserIDForDeviceData(groupIDToUserIDMap)
	if err != nil {
		return errors.Wrap(err, "unable to migrate group id to user id for device data")
	}

	return nil
}

func (m *Migration) buildMetaIDToUserIDMap() (map[string]string, error) {
	m.Logger().Debug("Building meta id to user id map")

	userIDMap := map[string]bool{}
	metaIDToUserIDMap := map[string]string{}

	m.Logger().Debug("Creating users store")

	mongoConfig := m.NewMongoConfig()
	mongoConfig.Database = "user"
	usersStore, err := mongo.NewStore(mongoConfig, m.Logger())
	if err != nil {
		return nil, errors.Wrap(err, "unable to create users store")
	}
	defer usersStore.Close()

	m.Logger().Debug("Creating users session")

	usersSession := usersStore.NewSession("users")
	defer usersSession.Close()

	m.Logger().Debug("Iterating users")

	iter := usersSession.C().Find(bson.M{}).Select(bson.M{"_id": 0, "userid": 1, "private.meta.id": 1}).Iter()

	var result struct {
		UserID  string `bson:"userid"`
		Private struct {
			Meta struct {
				ID string `bson:"id"`
			} `bson:"meta"`
		} `bson:"private"`
	}
	for iter.Next(&result) {
		userLogger := m.Logger()

		userID := result.UserID
		if userID == "" {
			userLogger.Warn("Missing user id in result from users query")
			continue
		}

		userLogger = userLogger.WithField("userId", userID)

		if _, ok := userIDMap[userID]; ok {
			userLogger.Error("Found multiple users with same user id")
			continue
		}
		userIDMap[userID] = true

		metaID := result.Private.Meta.ID
		if metaID == "" {
			userLogger.Warn("Missing private meta id in result from users query for user id")
			continue
		}

		userLogger = userLogger.WithField("metaId", metaID)

		if _, ok := metaIDToUserIDMap[metaID]; ok {
			userLogger.Error("Found multiple users with same meta id")
			continue
		}
		metaIDToUserIDMap[metaID] = userID
	}
	if err = iter.Close(); err != nil {
		return nil, errors.Wrap(err, "unable to iterate users")
	}

	m.Logger().Debugf("Found %d users with meta", len(metaIDToUserIDMap))

	return metaIDToUserIDMap, nil
}

func (m *Migration) buildGroupIDToUserIDMap(metaIDToUserIDMap map[string]string) (map[string]string, error) {
	m.Logger().Debug("Building group id to user id map")

	metaIDMap := map[string]bool{}
	groupIDToUserIDMap := map[string]string{}

	m.Logger().Debug("Creating meta store")

	mongoConfig := m.NewMongoConfig()
	mongoConfig.Database = "seagull"
	metaStore, err := mongo.NewStore(mongoConfig, m.Logger())
	if err != nil {
		return nil, errors.Wrap(err, "unable to create meta store")
	}
	defer metaStore.Close()

	m.Logger().Debug("Creating meta session")

	metaSession := metaStore.NewSession("seagull")
	defer metaSession.Close()

	m.Logger().Debug("Iterating meta")

	iter := metaSession.C().Find(bson.M{}).Iter()

	var result struct {
		MetaID string `bson:"_id"`
		Value  string `bson:"value"`
	}
	for iter.Next(&result) {
		metaLogger := m.Logger()

		metaID := result.MetaID
		if metaID == "" {
			metaLogger.Warn("Missing meta id in result from meta query")
			continue
		}

		metaLogger = metaLogger.WithField("metaId", metaID)

		if _, ok := metaIDMap[metaID]; ok {
			metaLogger.Error("Found multiple metas with same meta id")
			continue
		}
		metaIDMap[metaID] = true

		userID, ok := metaIDToUserIDMap[metaID]
		if !ok {
			metaLogger.Error("Missing user id for meta id")
			continue
		}

		metaLogger = metaLogger.WithField("userId", userID)

		if result.Value == "" {
			metaLogger.Warn("Missing value in result from meta query for meta id")
			continue
		}

		var value struct {
			Private struct {
				Uploads struct {
					ID string `json:"id"`
				} `json:"uploads"`
			} `json:"private"`
		}
		if err = json.Unmarshal([]byte(result.Value), &value); err != nil {
			metaLogger.WithError(err).Warn("Unable to unmarshal value from meta query for meta id")
			continue
		}

		groupID := value.Private.Uploads.ID
		if groupID == "" {
			metaLogger.Debug("Missing group id in value in result from meta query for meta id")
			continue
		}

		metaLogger = metaLogger.WithField("groupId", groupID)

		if _, ok = groupIDToUserIDMap[groupID]; ok {
			metaLogger.Error("Found multiple metas with same group id")
			continue
		}
		groupIDToUserIDMap[groupID] = userID
	}
	if err = iter.Close(); err != nil {
		return nil, errors.Wrap(err, "unable to iterate meta")
	}

	m.Logger().Debugf("Found %d groups with user", len(groupIDToUserIDMap))

	return groupIDToUserIDMap, nil
}

func (m *Migration) migrateGroupIDToUserIDForDeviceData(groupIDToUserIDMap map[string]string) error {
	m.Logger().Debug("Migrating group id to user id for device data")

	var migrateGroupCount int
	var migrateDeviceDataCount int

	m.Logger().Debug("Creating device data store")

	mongoConfig := m.NewMongoConfig()
	mongoConfig.Database = "data"
	deviceDataStore, err := mongo.NewStore(mongoConfig, m.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create device data store")
	}
	defer deviceDataStore.Close()

	m.Logger().Debug("Creating device data session")

	deviceDataSession := deviceDataStore.NewSession("deviceData")
	defer deviceDataSession.Close()

	m.Logger().Debug("Walking group id to user id map")

	var count int
	for groupID, userID := range groupIDToUserIDMap {
		dataLogger := m.Logger().WithFields(log.Fields{"groupId": groupID, "userId": userID})

		dataLogger.Debug("Finding device data for group id with incorrect existing user id")

		selector := bson.M{
			"$and": []bson.M{
				{"_groupId": groupID},
				{"_userId": bson.M{"$exists": true}},
				{"_userId": bson.M{"$ne": userID}},
			},
		}
		count, err = deviceDataSession.C().Find(selector).Count()
		if err != nil {
			dataLogger.WithError(err).Error("Unable to find incorrect device data")
			continue
		}

		if count != 0 {
			dataLogger.WithField("count", count).Error("Found device data for group id with incorrect existing user id")
			continue
		}

		dataLogger.Debug("Updating device data for group id with user id")

		selector = bson.M{
			"_groupId": groupID,
			"_userId":  bson.M{"$exists": false},
		}

		if m.DryRun() {
			count, err = deviceDataSession.C().Find(selector).Count()
		} else {
			update := bson.M{
				"$set": bson.M{"_userId": userID},
			}

			var changeInfo *mgo.ChangeInfo
			changeInfo, err = deviceDataSession.C().UpdateAll(selector, update)
			if changeInfo != nil {
				count = changeInfo.Updated
			}
		}

		if err != nil {
			dataLogger.WithError(err).Error("Unable to update device data for group id with user id")
			continue
		}

		if count > 0 {
			dataLogger.Infof("Migrated %d device data", count)
			migrateGroupCount++
			migrateDeviceDataCount += count
		}
	}

	if !m.DryRun() {
		if count, err = deviceDataSession.C().Find(bson.M{"_userId": bson.M{"$exists": false}}).Count(); err != nil {
			m.Logger().WithError(err).Error("Unable to query for device data without user id")
		} else if count != 0 {
			m.Logger().WithField("count", count).Error("Found device data without user id")
		}
	}

	m.Logger().Infof("Migrated %d device data for %d groups", migrateDeviceDataCount, migrateGroupCount)

	return nil
}
