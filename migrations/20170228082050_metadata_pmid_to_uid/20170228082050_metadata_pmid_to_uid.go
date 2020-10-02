package main

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/urfave/cli"

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
	application.RunAndExit(NewMigration())
}

type Migration struct {
	*mongoMigration.Migration
	index bool
}

func NewMigration() *Migration {
	return &Migration{
		Migration: mongoMigration.NewMigration(),
	}
}

func (m *Migration) Initialize(provider application.Provider) error {
	if err := m.Migration.Initialize(provider); err != nil {
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

	m.CLI().Action = func(ctx *cli.Context) error {
		if !m.ParseContext(ctx) {
			return nil
		}
		return m.execute()
	}

	return nil
}

func (m *Migration) ParseContext(ctx *cli.Context) bool {
	if parsed := m.Migration.ParseContext(ctx); !parsed {
		return parsed
	}

	m.index = ctx.Bool(IndexFlag)

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
	params := storeStructuredMongo.Params{DatabaseConfig: mongoConfig}
	usersStore, err := storeStructuredMongo.NewStore(params)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create users store")
	}
	defer usersStore.Terminate(context.Background())

	m.Logger().Debug("Creating users repository")

	usersRepository := usersStore.GetRepository("users")

	m.Logger().Debug("Iterating users")

	opts := options.Find().SetProjection(bson.M{"_id": 0, "userid": 1, "private.meta.id": 1})
	cursor, err := usersRepository.Find(context.Background(), bson.M{}, opts)

	var result struct {
		UserID  string `bson:"userid"`
		Private struct {
			Meta struct {
				ID string `bson:"id"`
			} `bson:"meta"`
		} `bson:"private"`
	}
	for cursor.Next(context.Background()) {
		if err = cursor.Decode(result); err != nil {
			return nil, errors.Wrap(err, "unable to decode users")
		}
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

	m.Logger().Debugf("Found %d users with meta", len(metaIDToUserIDMap))

	return metaIDToUserIDMap, nil
}

func (m *Migration) migrateMetaIDToUserIDForMetadata(metaIDToUserIDMap map[string]string) error {
	m.Logger().Debug("Migrating meta id to user id for metadata")

	var migrateMetaCount int
	var migrateMetadataCount int64

	m.Logger().Debug("Creating metadata data store")

	mongoConfig := m.NewMongoConfig()
	mongoConfig.Database = "seagull"
	params := storeStructuredMongo.Params{DatabaseConfig: mongoConfig}
	metadataStore, err := storeStructuredMongo.NewStore(params)
	if err != nil {
		return errors.Wrap(err, "unable to create metadata store")
	}
	defer metadataStore.Terminate(context.Background())

	m.Logger().Debug("Creating metadata repository")

	metadataRepository := metadataStore.GetRepository("seagull")

	m.Logger().Debug("Walking meta id to user id map")

	var count int64
	for metaID, userID := range metaIDToUserIDMap {
		metadataLogger := m.Logger().WithFields(log.Fields{"metaId": metaID, "userId": userID})

		metadataLogger.Debug("Finding metadata for meta id")

		var results []struct {
			ID     string  `bson:"_id"`
			UserID *string `bson:"userId"`
			Value  *string `bson:"value"`
		}
		cursor, err := metadataRepository.Find(context.Background(), bson.M{"_id": metaID})
		if err != nil {
			metadataLogger.WithError(err).Error("Unable to query for metadata")
			continue
		}

		if err = cursor.All(context.Background(), &results); err != nil {
			return errors.Wrap(err, "unable to decode metadata")
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
			count, err = metadataRepository.CountDocuments(context.Background(), selector)
		} else {
			update := bson.M{
				"$set": bson.M{"userId": userID},
			}

			var changeInfo *mongo.UpdateResult
			changeInfo, err = metadataRepository.UpdateMany(context.Background(), selector, update)
			if changeInfo != nil {
				count = changeInfo.ModifiedCount
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
		cursor, err := metadataRepository.Find(context.Background(), bson.M{"userId": bson.M{"$exists": false}})
		var result map[string]interface{}
		for cursor.Next(context.Background()) {
			if err = cursor.Decode(&result); err != nil {
				return errors.Wrap(err, "unable to decode metadata")
			}
			m.Logger().WithField("metaId", result["_id"]).Error("Metadata found without user id")
		}
	}

	m.Logger().Infof("Migrated %d metadata for %d meta", migrateMetadataCount, migrateMetaCount)

	if m.Index() {
		m.Logger().Info("Creating unique index on user id")

		index := []mongo.IndexModel{{
			Keys: bson.D{{Key: "userId", Value: 1}},
			Options: options.Index().
				SetUnique(true).
				SetBackground(true),
		}}
		err = metadataRepository.CreateAllIndexes(context.Background(), index)
		if err != nil {
			return errors.Wrap(err, "unable to create metadata index on user id")
		}
	}

	return nil
}
