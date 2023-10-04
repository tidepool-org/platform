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

func getValidatedString(bsonData bson.M, fieldName string) (string, error) {
	if valRaw, ok := bsonData[fieldName]; !ok {
		return "", errors.Newf("%s is missing", fieldName)
	} else if val, ok := valRaw.(string); !ok {
		return "", errors.Newf("%s is not of expected type", fieldName)
	} else if val == "" {
		return "", errors.Newf("%s is empty", fieldName)
	} else {
		return val, nil
	}
}

func getValidatedTime(bsonData bson.M, fieldName string) (time.Time, error) {
	if valRaw, ok := bsonData[fieldName]; !ok {
		return time.Time{}, errors.Newf("%s is missing", fieldName)
	} else if val, ok := valRaw.(time.Time); !ok {
		return time.Time{}, errors.Newf("%s is not of expected type", fieldName)
	} else if val.IsZero() {
		return time.Time{}, errors.Newf("%s is empty", fieldName)
	} else {
		return val, nil
	}
}

func createDatumHash(bsonData bson.M) (string, error) {
	identityFields := []string{}
	if userID, err := getValidatedString(bsonData, "_userId"); err != nil {
		return "", err
	} else {
		identityFields = append(identityFields, userID)
	}
	if deviceID, err := getValidatedString(bsonData, "deviceId"); err != nil {
		return "", err
	} else {
		identityFields = append(identityFields, deviceID)
	}
	if datumTime, err := getValidatedTime(bsonData, "time"); err != nil {
		return "", err
	} else {
		identityFields = append(identityFields, datumTime.Format(types.TimeFormat))
	}
	datumType, err := getValidatedString(bsonData, "type")
	if err != nil {
		return "", err
	}
	identityFields = append(identityFields, datumType)

	switch datumType {
	case "basal":
		if deliveryType, err := getValidatedString(bsonData, "deliveryType"); err != nil {
			return "", err
		} else {
			identityFields = append(identityFields, deliveryType)
		}
	case "bolus", "deviceEvent":
		if subType, err := getValidatedString(bsonData, "subType"); err != nil {
			return "", err
		} else {
			identityFields = append(identityFields, subType)
		}
	case "smbg", "bloodKetone", "cbg":
		if units, err := getValidatedString(bsonData, "units"); err != nil {
			return "", err
		} else {
			identityFields = append(identityFields, units)
		}

		valueRaw, ok := bsonData["value"]
		if !ok {
			return "", errors.New("value is missing")
		}
		val, ok := valueRaw.(float64)
		if !ok {
			return "", errors.New("value is not of expected type")
		}

		if len(fmt.Sprintf("%v", valueRaw)) > 7 {
			// NOTE: we need to ensure the same precision for the
			// converted value as it is used to calculate the hash
			mgdlVal := val*18.01559 + 0.5
			mgdL := glucose.MgdL
			val = *glucose.NormalizeValueForUnits(&mgdlVal, &mgdL)
		}
		strVal := strconv.FormatFloat(val, 'f', -1, 64)
		identityFields = append(identityFields, strVal)
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

	datumID, err := getValidatedString(jfDatum, "_id")
	if err != nil {
		return err
	}

	var modifiedTime *time.Time
	if timeRaw, ok := jfDatum["modifiedTime"]; !ok {
		modifiedTime = nil
	} else if val, ok := timeRaw.(time.Time); !ok {
		modifiedTime = nil
	} else {
		modifiedTime = &val
	}

	hash, err := createDatumHash(jfDatum)
	if err != nil {
		return err
	}
	update := bson.M{
		"$set": bson.M{"_deduplicator": bson.M{"hash": hash}},
	}
	_, err = m.dataRepository.UpdateOne(m.ctx, bson.M{
		"_id":          datumID,
		"modifiedTime": modifiedTime,
	}, update)
	return err
}
