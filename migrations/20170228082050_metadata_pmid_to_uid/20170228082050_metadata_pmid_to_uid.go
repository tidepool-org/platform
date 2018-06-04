package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	mongoMigration "github.com/tidepool-org/platform/migration/mongo"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

const (
	IndexFlag = "index"
)

func main() {
	os.Exit(application.Run(NewMigration()))
}

type Migration struct {
	*mongoMigration.Migration
	index bool
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

	m.CLI().Usage = "Migrate all metadata to add user id derived from _id"
	m.CLI().Description = "Migrate all metadata to add the 'userId' field derived from the '_id' field. The Seagull '_id' field matches the Shoreline 'private.meta.id' field." +
		"\n\n   One or more warnings will be reported if partially created accounts or invalid data are found." +
		"\n\n   This migration is idempotent." +
		"\n\n   NOTE: This migration MUST be executed immediately BEFORE and immediately AFTER Seagull is updated to v0.3.1."
	m.CLI().Authors = []cli.Author{
		{
			Name:  "Darin Krauss",
			Email: "darin@tidepool.org",
		},
	}
	m.CLI().Flags = append(m.CLI().Flags,
		cli.BoolFlag{
			Name:  fmt.Sprintf("%s,%s", IndexFlag, "i"),
			Usage: "add unique index after migration",
		},
	)

	m.CLI().Action = func(context *cli.Context) error {
		if !m.ParseContext(context) {
			return nil
		}
		return m.execute()
	}

	return nil
}

func (m *Migration) ParseContext(context *cli.Context) bool {
	if parsed := m.Migration.ParseContext(context); !parsed {
		return parsed
	}

	m.index = context.Bool(IndexFlag)

	return true
}

func (m *Migration) Index() bool {
	return m.index
}

func (m *Migration) execute() error {
	if m.Index() && m.DryRun() {
		return errors.New("cannot specify --index with --dry-run")
	}

	metaIDToUserIDMap, err := m.buildMetaIDToUserIDMap()
	if err != nil {
		return errors.Wrap(err, "unable to build meta id to user id map")
	}

	err = m.migrateMetaIDToUserIDForMetadata(metaIDToUserIDMap)
	if err != nil {
		return errors.Wrap(err, "unable to migrate meta id to user id for metadata")
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
	usersStore, err := storeStructuredMongo.NewStore(mongoConfig, m.Logger())
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
			userLogger.Error("Multiple users found with same user id")
			continue
		}
		userIDMap[userID] = true

		metaID := result.Private.Meta.ID
		if metaID == "" {
			userLogger.Warn("User found without meta id")
			continue
		}

		userLogger = userLogger.WithField("metaId", metaID)

		if _, ok := metaIDToUserIDMap[metaID]; ok {
			userLogger.Error("Multiple users found with same meta id")
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

func (m *Migration) migrateMetaIDToUserIDForMetadata(metaIDToUserIDMap map[string]string) error {
	m.Logger().Debug("Migrating meta id to user id for metadata")

	var migrateMetaCount int
	var migrateMetadataCount int

	m.Logger().Debug("Creating metadata data store")

	mongoConfig := m.NewMongoConfig()
	mongoConfig.Database = "seagull"
	metadataStore, err := storeStructuredMongo.NewStore(mongoConfig, m.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create metadata store")
	}
	defer metadataStore.Close()

	m.Logger().Debug("Creating metadata session")

	metadataSession := metadataStore.NewSession("seagull")
	defer metadataSession.Close()

	m.Logger().Debug("Walking meta id to user id map")

	var count int
	for metaID, userID := range metaIDToUserIDMap {
		metadataLogger := m.Logger().WithFields(log.Fields{"metaId": metaID, "userId": userID})

		metadataLogger.Debug("Finding metadata for meta id")

		var results []struct {
			ID     string  `bson:"_id"`
			UserID *string `bson:"userId"`
			Value  *string `bson:"value"`
		}
		err = metadataSession.C().Find(bson.M{"_id": metaID}).All(&results)
		if err != nil {
			metadataLogger.WithError(err).Error("Unable to query for metadata")
			continue
		}

		resultsCount := len(results)
		switch resultsCount {
		case 0:
			metadataLogger.Error("Metadata not found for meta id")
			continue
		case 1:
			break
		default:
			metadataLogger.WithField("count", resultsCount).Error("More than one metadata found for meta id")
			continue
		}

		if result := results[0]; result.UserID != nil {
			if existingUserID := *result.UserID; existingUserID != userID {
				metadataLogger.WithField("existingUserId", existingUserID).Error("Metadata found for meta id with incorrect existing user id")
			}
			continue
		}

		metadataLogger.Debug("Updating metadata for meta id with user id")

		selector := bson.M{
			"_id":    metaID,
			"userId": bson.M{"$exists": false},
		}

		if m.DryRun() {
			count, err = metadataSession.C().Find(selector).Count()
		} else {
			update := bson.M{
				"$set": bson.M{"userId": userID},
			}

			var changeInfo *mgo.ChangeInfo
			changeInfo, err = metadataSession.C().UpdateAll(selector, update)
			if changeInfo != nil {
				count = changeInfo.Updated
			}
		}

		if err != nil {
			metadataLogger.WithError(err).Error("Unable to update metadata for meta id with user id")
			continue
		}

		if count > 0 {
			metadataLogger.Infof("Migrated %d metadata", count)
			migrateMetaCount++
			migrateMetadataCount += count
		}
	}

	if !m.DryRun() {
		iter := metadataSession.C().Find(bson.M{"userId": bson.M{"$exists": false}}).Iter()
		var result map[string]interface{}
		for iter.Next(&result) {
			m.Logger().WithField("metaId", result["_id"]).Error("Metadata found without user id")
		}
		if err = iter.Close(); err != nil {
			return errors.Wrap(err, "unable to iterate metadata without user id")
		}
	}

	m.Logger().Infof("Migrated %d metadata for %d meta", migrateMetadataCount, migrateMetaCount)

	if m.Index() {
		m.Logger().Info("Creating unique index on user id")

		index := mgo.Index{
			Key:        []string{"userId"},
			Unique:     true,
			Background: false,
		}
		err = metadataSession.C().EnsureIndex(index)
		if err != nil {
			return errors.Wrap(err, "unable to create metadata index on user id")
		}
	}

	return nil
}
