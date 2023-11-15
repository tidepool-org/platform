package main

import (
	"context"
	"time"

	"github.com/urfave/cli"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/errors"
	migrationMongo "github.com/tidepool-org/platform/migration/mongo"
	"github.com/tidepool-org/platform/migrations/back_37/utils"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	application.RunAndExit(NewMigration(ctx))
}

type Migration struct {
	ctx context.Context
	*migrationMongo.Migration
	dataRepository *storeStructuredMongo.Repository
}

func NewMigration(ctx context.Context) *Migration {
	return &Migration{
		ctx:       ctx,
		Migration: migrationMongo.NewMigration(),
	}
}

func (m *Migration) Initialize(provider application.Provider) error {
	if err := m.Migration.Initialize(provider); err != nil {
		return err
	}

	m.CLI().Usage = "BACK-37: Migrate all existing data to add required Platform deduplication hash fields"
	m.CLI().Description = "BACK-37: To fully migrate devices from the `jellyfish` upload API to the `platform` upload API"
	m.CLI().Authors = []cli.Author{
		{
			Name:  "J H BATE",
			Email: "jamie@tidepool.org",
		},
	}

	m.CLI().Action = func(ctx *cli.Context) error {
		if !m.ParseContext(ctx) {
			return nil
		}
		return m.execute()
	}

	return nil
}

func (m *Migration) execute() error {
	m.Logger().Debug("Migrate jellyfish API data")
	m.Logger().Debug("Creating data store")

	mongoConfig := m.NewMongoConfig()
	mongoConfig.Database = "data"
	mongoConfig.Timeout = 60 * time.Minute
	dataStore, err := storeStructuredMongo.NewStore(mongoConfig)
	if err != nil {
		return errors.Wrap(err, "unable to create data store")
	}
	defer dataStore.Terminate(m.ctx)

	m.Logger().Debug("Creating data repository")
	m.dataRepository = dataStore.GetRepository("deviceData")
	m.Logger().Info("Migration of jellyfish documents has begun")
	hashUpdatedCount, errorCount := m.migrateJellyfishDocuments()
	m.Logger().Infof("Migrated %d jellyfish documents", hashUpdatedCount)
	m.Logger().Infof("%d errors occurred", errorCount)

	return nil
}

func (m *Migration) migrateJellyfishDocuments() (int, int) {
	logger := m.Logger()
	logger.Debug("Finding jellyfish data")
	var hashUpdatedCount, errorCount int
	selector := bson.M{
		// jellyfish uses a generated _id that is not an mongo objectId
		"_id":           bson.M{"$not": bson.M{"$type": "objectId"}},
		"_deduplicator": bson.M{"$exists": false},
	}

	var jellyfishResult bson.M
	jellyfishDocCursor, err := m.dataRepository.Find(m.ctx, selector)
	if err != nil {
		logger.WithError(err).Error("Unable to find jellyfish data")
		errorCount++
		return hashUpdatedCount, errorCount
	}
	defer jellyfishDocCursor.Close(m.ctx)
	for jellyfishDocCursor.Next(m.ctx) {
		err = jellyfishDocCursor.Decode(&jellyfishResult)
		if err != nil {
			logger.WithError(err).Error("Could not decode mongo doc")
			errorCount++
			continue
		}
		if !m.DryRun() {
			if updated, err := m.migrateDocument(jellyfishResult); err != nil {
				logger.WithError(err).Errorf("Unable to migrate jellyfish document %s.", jellyfishResult["_id"])
				errorCount++
				continue
			} else if updated {
				hashUpdatedCount++
			}
		}
	}
	if err := jellyfishDocCursor.Err(); err != nil {
		logger.WithError(err).Error("Error while fetching data. Please re-run to complete the migration.")
		errorCount++
	}
	return hashUpdatedCount, errorCount
}

func (m *Migration) migrateDocument(jfDatum bson.M) (bool, error) {

	datumID, err := utils.GetValidatedString(jfDatum, "_id")
	if err != nil {
		return false, err
	}

	// updates := bson.M{}
	// hash, err := utils.CreateDatumHash(jfDatum)
	// if err != nil {
	// 	return false, err
	// }

	// updates["_deduplicator"] = bson.M{"hash": hash}

	updates, err := utils.GetDatumUpdates(jfDatum)
	if err != nil {
		return false, err
	}

	// if boluses, err := utils.UpdateIfExistsPumpSettingsBolus(jfDatum); err != nil {
	// 	return false, err
	// } else if boluses != nil {
	// 	updates["boluses"] = boluses
	// }

	// if sleepSchedules, err := utils.UpdateIfExistsPumpSettingsSleepSchedules(jfDatum); err != nil {
	// 	return false, err
	// } else if sleepSchedules != nil {
	// 	updates["sleepSchedules"] = sleepSchedules
	// }

	result, err := m.dataRepository.UpdateOne(m.ctx, bson.M{
		"_id":          datumID,
		"modifiedTime": jfDatum["modifiedTime"],
	}, bson.M{
		"$set": updates,
	})

	if err != nil {
		return false, err
	}
	return result.ModifiedCount == 1, nil
}
