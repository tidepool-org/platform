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
	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	"github.com/tidepool-org/platform/data/types/blood/ketone"
	"github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/data/types/device"
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
	if datumUserID, err := getValidatedString(bsonData, "_userId"); err != nil {
		return "", err
	} else {
		identityFields = append(identityFields, datumUserID)
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
	case basal.Type:
		if deliveryType, err := getValidatedString(bsonData, "deliveryType"); err != nil {
			return "", err
		} else {
			identityFields = append(identityFields, deliveryType)
		}
	case bolus.Type, device.Type:
		if subType, err := getValidatedString(bsonData, "subType"); err != nil {
			return "", err
		} else {
			identityFields = append(identityFields, subType)
		}
	case selfmonitored.Type, ketone.Type, continuous.Type:
		units, err := getValidatedString(bsonData, "units")
		if err != nil {
			return "", err
		} else {
			identityFields = append(identityFields, units)
		}

		if valueRaw, ok := bsonData["value"]; !ok {
			return "", errors.New("value is missing")
		} else if val, ok := valueRaw.(float64); !ok {
			return "", errors.New("value is not of expected type")
		} else {
			if units != glucose.MgdL && units != glucose.Mgdl {
				// NOTE: we need to ensure the same precision for the
				// converted value as it is used to calculate the hash
				val = getBGValuePlatformPrecision(val)
			}
			identityFields = append(identityFields, strconv.FormatFloat(val, 'f', -1, 64))
		}
	}
	return deduplicator.GenerateIdentityHash(identityFields)
}

func updateIfExistsPumpSettingsBolus(bsonData bson.M) (interface{}, error) {
	dataType, err := getValidatedString(bsonData, "type")
	if err != nil {
		return nil, err
	}
	if dataType == "pumpSettings" {
		if bolus := bsonData["bolus"]; bolus != nil {
			boluses, ok := bolus.(map[string]interface{})
			if !ok {
				return nil, errors.Newf("pumpSettings.bolus is not the expected type %v", bolus)
			}
			return boluses, nil
		}
	}
	return nil, nil
}

func getBGValuePlatformPrecision(mmolVal float64) float64 {
	if len(fmt.Sprintf("%v", mmolVal)) > 7 {
		mgdlVal := mmolVal * glucose.MmolLToMgdLConversionFactor
		mgdL := glucose.MgdL
		mmolVal = *glucose.NormalizeValueForUnits(&mgdlVal, &mgdL)
	}
	return mmolVal
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

	datumID, err := getValidatedString(jfDatum, "_id")
	if err != nil {
		return false, err
	}

	updates := bson.M{}
	hash, err := createDatumHash(jfDatum)
	if err != nil {
		return false, err
	}

	updates["_deduplicator"] = bson.M{"hash": hash}

	if boluses, err := updateIfExistsPumpSettingsBolus(jfDatum); err != nil {
		return false, err
	} else if boluses != nil {
		updates["pumpSettings"] = bson.M{"boluses": boluses}
	}

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
