package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/urfave/cli"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/errors"
	migrationMongo "github.com/tidepool-org/platform/migration/mongo"
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

func createDatumHash(bsonData bson.M) (string, error) {
	identityFields := []string{}
	if bsonData["_userId"] == nil {
		return "", errors.New("user id is missing")
	}
	userID := bsonData["_userId"].(string)
	if userID == "" {
		return "", errors.New("user id is empty")
	}
	identityFields = append(identityFields, userID)
	if bsonData["deviceId"] == nil {
		return "", errors.New("device id is missing")
	}
	deviceID := bsonData["deviceId"].(string)
	if deviceID == "" {
		return "", errors.New("device id is empty")
	}
	identityFields = append(identityFields, deviceID)
	if bsonData["time"] == nil {
		return "", errors.New("time is missing")
	}
	dataTime := bsonData["time"].(time.Time)
	if dataTime.IsZero() {
		return "", errors.New("time is empty")
	}
	identityFields = append(identityFields, dataTime.Format(types.TimeFormat))
	if bsonData["type"] == nil {
		return "", errors.New("type is missing")
	}
	dataType := bsonData["type"].(string)
	if dataType == "" {
		return "", errors.New("type is empty")
	}
	identityFields = append(identityFields, dataType)

	switch dataType {
	case "basal":
		if bsonData["deliveryType"] == nil {
			return "", errors.New("deliveryType is missing")
		}
		deliveryType := bsonData["deliveryType"].(string)
		if deliveryType == "" {
			return "", errors.New("deliveryType is empty")
		}
		identityFields = append(identityFields, deliveryType)
	case "bolus", "deviceEvent":
		if bsonData["subType"] == nil {
			return "", errors.New("subType is missing")
		}
		subType := bsonData["subType"].(string)
		if subType == "" {
			return "", errors.New("subType is empty")
		}
		identityFields = append(identityFields, subType)
	case "smbg", "bloodKetone", "cbg":
		if bsonData["units"] == nil {
			return "", errors.New("units is missing")
		}
		units := bsonData["units"].(string)
		if units == "" {
			return "", errors.New("units is empty")
		}
		identityFields = append(identityFields, units)
		if bsonData["value"] == nil {
			return "", errors.New("value is missing")
		}
		value := strconv.FormatFloat(bsonData["value"].(float64), 'f', -1, 64)
		identityFields = append(identityFields, value)
	}
	return deduplicator.GenerateIdentityHash(identityFields)
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
			if err := m.migrateDocument(jellyfishResult); err != nil {
				logger.WithError(err).Errorf("Unable to migrate jellyfish document %s.", jellyfishResult["_id"])
				errorCount++
				continue
			}
		}
		hashUpdatedCount++
	}
	if err := jellyfishDocCursor.Err(); err != nil {
		logger.WithError(err).Error("error while fetching data. Please re-run to complete the migration.")
		errorCount++
	}
	return hashUpdatedCount, errorCount
}

func (m *Migration) migrateDocument(jfDatum bson.M) error {
	var update bson.M

	switch jfDatum["type"] {
	case "smbg", "bloodKetone", "cbg":
		if len(fmt.Sprintf("%v", jfDatum["value"])) > 7 {
			// NOTE: we need to ensure the same precision for the
			// converted value as it is used to calculate the hash
			val := jfDatum["value"].(float64)
			mgdlVal := val*18.01559 + 0.5
			mgdL := glucose.MgdL
			jfDatum["value"] = glucose.NormalizeValueForUnits(&mgdlVal, &mgdL)
			hash, err := createDatumHash(jfDatum)
			if err != nil {
				return err
			}

			update = bson.M{
				"$set": bson.M{
					"_deduplicator": bson.M{"hash": hash},
					"value":         jfDatum["value"],
				},
			}
		}
	default:
		hash, err := createDatumHash(jfDatum)
		if err != nil {
			return err
		}
		update = bson.M{
			"$set": bson.M{"_deduplicator": bson.M{"hash": hash}},
		}
	}
	_, err := m.dataRepository.UpdateOne(m.ctx, bson.M{"_id": jfDatum["_id"], "modifiedTime": jfDatum["modifiedTime"]}, update)
	return err
}
