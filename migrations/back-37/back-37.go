package main

import (
	"context"
	"fmt"
	"time"

	"github.com/urfave/cli"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	"github.com/tidepool-org/platform/errors"
	migrationMongo "github.com/tidepool-org/platform/migration/mongo"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

func main() {
	application.RunAndExit(NewMigration())
}

type Migration struct {
	*migrationMongo.Migration
	dataRepository *storeStructuredMongo.Repository
}

func NewMigration() *Migration {
	return &Migration{
		Migration: migrationMongo.NewMigration(),
	}
}

func (m *Migration) Initialize(provider application.Provider) error {
	if err := m.Migration.Initialize(provider); err != nil {
		return err
	}

	m.CLI().Usage = "BACK-37: Migrate all existing data to add required Platform deduplication hash fields"
	m.CLI().Description = "BACK-37: To fully migrate devices from the `jellyfish`\n" +
		" 	upload API to the `platform` upload API"
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
	m.Logger().Debug("Migrate jellyfish upload API data")
	m.Logger().Debug("Creating data store")

	mongoConfig := m.NewMongoConfig()
	mongoConfig.Database = "data"
	mongoConfig.Timeout = 60 * time.Minute
	dataStore, err := storeStructuredMongo.NewStore(mongoConfig)
	if err != nil {
		return errors.Wrap(err, "unable to create data store")
	}
	defer dataStore.Terminate(context.Background())

	m.Logger().Debug("Creating data repository")

	m.dataRepository = dataStore.GetRepository("deviceData")
	hashUpdatedCount, archivedCount, errorCount := m.migrateJellyfishDocuments()
	m.Logger().Infof("Migrated %d jellyfish documents", hashUpdatedCount)
	m.Logger().Infof("Archived %d duplicate jellyfish documents", archivedCount)
	m.Logger().Infof("%d errors occurred", errorCount)

	return nil
}

func (m *Migration) migrateJellyfishDocuments() (int, int, int) {
	logger := m.Logger()

	logger.Debug("Finding distinct users")

	var hashUpdatedCount, archivedCount, errorCount int

	userIDs, err := m.dataRepository.Distinct(context.Background(), "_userId", bson.M{})
	if err != nil {
		logger.WithError(err).Error("Unable to execute distinct query")
	} else {
		logger.Debugf("Finding jellyfish records for %d users", len(userIDs))

		for _, userID := range userIDs {
			logger.Debugf("Finding jellyfish records for user ID %s", userID)
			selector := bson.M{
				"_userId": userID,
				"_active": true,
				// uploads aren't de-duped.
				"type":          bson.M{"$ne": "upload"},
				"_deduplicator": bson.M{"$exists": false},
			}

			var jellyfishResult bson.M
			jellyfishDocCursor, err := m.dataRepository.Find(context.Background(), selector)
			if err != nil {
				logger.WithError(err).Error("Unable to find jellyfish results")
			}
			for jellyfishDocCursor.Next(context.Background()) {
				jellyfishDocCursor.Decode(&jellyfishResult)

				dupQuery := bson.M{
					"_userId":       userID,
					"_active":       true,
					"_id":           jellyfishResult["_id"],
					"_deduplicator": bson.M{"$exists": true},
				}
				dupCursor, err := m.dataRepository.Find(context.Background(), dupQuery)
				if err != nil {
					logger.WithError(err).Errorf("Could not query for duplicate datum %s.", jellyfishResult["_id"])
					errorCount++
					continue
				}
				if !dupCursor.Next(context.Background()) {
					err = m.archiveDocument(jellyfishResult["_id"])
					if err != nil {
						logger.WithError(err).Error("Unable to archive jellyfish document")
						errorCount++
					}
					archivedCount++
				} else {
					if err := m.migrateDocument(jellyfishResult); err != nil {
						logger.WithError(err).Error("Unable to migrate jellyfish document")
						errorCount++
					}
					hashUpdatedCount++
				}
			}
			if err := jellyfishDocCursor.Err(); err != nil {
				logger.WithError(err).Error("error while fetching data. Please re-run to complete the migration.")
				errorCount++
			}
			jellyfishDocCursor.Close(context.Background())
		}
	}

	if err != nil {
		logger.WithError(err).Error("Unable to migrate jellyfish documents")
		errorCount++
	}

	return hashUpdatedCount, archivedCount, errorCount
}

func (m *Migration) migrateDocument(jfDatum bson.M) error {
	var deduplicatorUpdate bson.M

	switch jfDatum["type"] {
	case "smbg", "bloodKetone", "cbg":
		if len(fmt.Sprintf("%v", jfDatum["value"])) > 7 {
			// NOTE: we need to ensure the same precision for the
			// converted value as it is used to calculate the hash
			val := jfDatum["value"].(float64)
			mgdlVal := val*18.01559 + 0.5
			mgdL := glucose.MgdL
			jfDatum["value"] = glucose.NormalizeValueForUnits(&mgdlVal, &mgdL)

			deduplicatorUpdate = bson.M{
				"$set": bson.M{
					"_deduplicator": bson.M{
						"name":    deduplicator.DeviceDeactivateHashName,
						"version": "1.1.0",
						"hash":    CreateHash(jfDatum),
					},
					"value": jfDatum["value"],
				},
			}
		}
	default:
		deduplicatorUpdate = bson.M{
			"$set": bson.M{
				"_deduplicator": bson.M{
					"name":    deduplicator.DeviceDeactivateHashName,
					"version": "1.1.0",
					"hash":    CreateHash(jfDatum),
				},
			},
		}
	}
	_, err := m.dataRepository.UpdateOne(context.Background(), bson.M{"_id": jfDatum["_id"]}, deduplicatorUpdate)
	return err
}

func (m *Migration) archiveDocument(objectId interface{}) error {
	archiveUpdate := bson.M{
		"$set": bson.M{
			"_active":       false,
			"_archivedTime": time.Now().UnixNano() / int64(time.Millisecond),
		},
	}
	_, err := m.dataRepository.UpdateOne(context.Background(), bson.M{"_id": objectId}, archiveUpdate)
	return err
}
