package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/r3labs/diff/v3"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/data"

	"github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/basal"

	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	"github.com/tidepool-org/platform/data/types/blood/ketone"
	"github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/data/types/common"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	errorsP "github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/metadata"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypesFactory "github.com/tidepool-org/platform/data/types/factory"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func updateIfExistsPumpSettingsSleepSchedules(bsonData bson.M) (*pump.SleepScheduleMap, error) {
	//TODO: currently an array but should be a map for consistency. On pump is "Sleep Schedule 1", "Sleep Schedule 2"
	scheduleNames := map[int]string{0: "1", 1: "2"}

	if schedules := bsonData["sleepSchedules"]; schedules != nil {
		sleepScheduleMap := pump.SleepScheduleMap{}
		dataBytes, err := json.Marshal(schedules)
		if err != nil {
			return nil, err
		}
		schedulesArray := []*pump.SleepSchedule{}
		err = json.Unmarshal(dataBytes, &schedulesArray)
		if err != nil {
			return nil, err
		}
		for i, schedule := range schedulesArray {
			days := schedule.Days
			updatedDays := []string{}
			for _, day := range *days {
				if !slices.Contains(common.DaysOfWeek(), strings.ToLower(day)) {
					return nil, errorsP.Newf("pumpSettings.sleepSchedules has an invalid day of week %s", day)
				}
				updatedDays = append(updatedDays, strings.ToLower(day))
			}
			schedule.Days = &updatedDays
			sleepScheduleMap[scheduleNames[i]] = schedule
		}
		//sorts schedules based on day
		sleepScheduleMap.Normalize(dataNormalizer.New())
		return &sleepScheduleMap, nil
	}
	return nil, nil
}

func pumpSettingsHasBolus(bsonData bson.M) bool {
	if bolus := bsonData["bolus"]; bolus != nil {
		if _, ok := bolus.(*pump.BolusMap); ok {
			return true
		}
	}
	return false
}

func logDiff(id string, updates interface{}) {
	f, err := os.OpenFile("diff.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer f.Close()
	updatesJSON, _ := json.Marshal(updates)
	f.WriteString(fmt.Sprintf(`{"_id":"%s","diff":%s},`, id, string(updatesJSON)))
}

func ProcessDatum(bsonData bson.M) (data.Datum, error) {

	dType := fmt.Sprintf("%v", bsonData["type"])
	dID := fmt.Sprintf("%v", bsonData["_id"])

	switch dType {
	case pump.Type:
		if boluses := bsonData["bolus"]; boluses != nil {
			bsonData["boluses"] = boluses
			delete(bsonData, "bolus")
		}
		// case selfmonitored.Type, ketone.Type, continuous.Type:
		// 	units := fmt.Sprintf("%v", bsonData["units"])
		// 	if units == glucose.MmolL || units == glucose.Mmoll {

		// 		if val, ok := bsonData["value"].(float64); ok {

		// 		}

		// 	}
	}

	if payload := bsonData["payload"]; payload != nil {
		if _, ok := payload.(string); ok {
			dataBytes, err := bson.Marshal(payload)
			if err != nil {
				return nil, err
			}
			var payloadMetadata metadata.Metadata
			err = bson.Unmarshal(dataBytes, &payloadMetadata)
			if err != nil {
				return nil, errorsP.Newf("payload could not be set from %v ", string(dataBytes))
			}
			bsonData["payload"] = &payloadMetadata
		}
	}
	if annotations := bsonData["annotations"]; annotations != nil {
		if _, ok := annotations.(string); ok {
			dataBytes, err := bson.Marshal(annotations)
			if err != nil {
				return nil, err
			}
			var metadataArray metadata.MetadataArray
			if err := bson.Unmarshal(dataBytes, &metadataArray); err != nil {
				return nil, errorsP.Newf("annotations could not be set from %v ", string(dataBytes))
			}
			bsonData["annotations"] = &metadataArray
		}
	}

	incomingJSONData, err := json.Marshal(bsonData)
	if err != nil {
		return nil, err
	}
	ojbData := map[string]interface{}{}
	if err := json.Unmarshal(incomingJSONData, &ojbData); err != nil {
		return nil, err
	}

	//parsing
	parser := structureParser.NewObject(&ojbData)
	validator := structureValidator.New()
	normalizer := dataNormalizer.New()

	datum := dataTypesFactory.ParseDatum(parser)
	if datum != nil && *datum != nil {
		(*datum).Validate(validator)
		(*datum).Normalize(normalizer)
	} else {
		return nil, errorsP.Newf("no datum returned for id=[%s]", dID)
	}

	validator.Bool("_active", parser.Bool("_active")).Exists()
	validator.String("_groupId", parser.String("_groupId")).Exists()
	validator.String("_id", parser.String("_id")).Exists()
	validator.String("_userId", parser.String("_userId")).Exists()
	validator.Int("_version", parser.Int("_version")).Exists()
	validator.Int("_schemaVersion", parser.Int("_schemaVersion"))
	validator.Object("_deduplicator", parser.Object("_deduplicator")).Exists()
	validator.String("uploadId", parser.String("uploadId")).Exists()
	validator.String("guid", parser.String("guid")).Exists()
	validator.Time("createdTime", parser.Time("createdTime", time.RFC3339Nano)).Exists()
	validator.Time("modifiedTime", parser.Time("modifiedTime", time.RFC3339Nano))

	parser.NotParsed()

	if err := parser.Error(); err != nil {
		return nil, err
	}

	if err := validator.Error(); err != nil {
		return nil, err
	}

	if err := normalizer.Error(); err != nil {
		return nil, err
	}

	outgoingJSONData, err := json.Marshal(datum)
	if err != nil {
		return nil, err
	}

	processedData := map[string]interface{}{}
	if err := json.Unmarshal(outgoingJSONData, &processedData); err != nil {
		return nil, err
	}

	// these are extras that we want to leave on the original object
	notRequired := []string{"_active", "_groupId", "_id", "_userId", "_version", "_schemaVersion", "_deduplicator", "uploadId"}
	for _, key := range notRequired {
		delete(ojbData, key)
	}

	changelog, _ := diff.Diff(ojbData, processedData, diff.StructMapKeySupport())
	logDiff(dID, changelog)

	return *datum, nil
}

func GetDatumUpdates(bsonData bson.M) (string, []bson.M, error) {
	updates := []bson.M{}
	set := bson.M{}
	var rename bson.M
	var identityFields []string

	datumID, ok := bsonData["_id"].(string)
	if !ok {
		return "", nil, errorsP.New("cannot get the datum id")
	}

	datumType, ok := bsonData["type"].(string)
	if !ok {
		return datumID, nil, errorsP.New("cannot get the datum type")
	}

	// TODO: based on discussions we want to ensure that these are the correct type
	// even though we are not using them for the hash generation
	delete(bsonData, "payload")
	delete(bsonData, "annotations")

	switch datumType {
	case basal.Type:
		var datum *basal.Basal
		dataBytes, err := bson.Marshal(bsonData)
		if err != nil {
			return datumID, nil, err
		}
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return datumID, nil, err
		}
		identityFields, err = datum.IdentityFields()
		if err != nil {
			return datumID, nil, err
		}
	case bolus.Type:
		var datum *bolus.Bolus
		dataBytes, err := bson.Marshal(bsonData)
		if err != nil {
			return datumID, nil, err
		}
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return datumID, nil, err
		}
		identityFields, err = datum.IdentityFields()
		if err != nil {
			return datumID, nil, err
		}
	case device.Type:
		var datum bolus.Bolus
		dataBytes, err := bson.Marshal(bsonData)
		if err != nil {
			return datumID, nil, err
		}
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return datumID, nil, err
		}
		identityFields, err = datum.IdentityFields()
		if err != nil {
			return datumID, nil, err
		}
	case pump.Type:
		var datum types.Base
		dataBytes, err := bson.Marshal(bsonData)
		if err != nil {
			return datumID, nil, err
		}
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return datumID, nil, err
		}
		identityFields, err = datum.IdentityFields()
		if err != nil {
			return datumID, nil, err
		}

		if pumpSettingsHasBolus(bsonData) {
			rename = bson.M{"bolus": "boluses"}
		}

		sleepSchedules, err := updateIfExistsPumpSettingsSleepSchedules(bsonData)
		if err != nil {
			return datumID, nil, err
		} else if sleepSchedules != nil {
			set["sleepSchedules"] = sleepSchedules
		}
	case selfmonitored.Type:
		var datum selfmonitored.SelfMonitored
		dataBytes, err := bson.Marshal(bsonData)
		if err != nil {
			return datumID, nil, err
		}
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return datumID, nil, err
		}
		beforeVal := datum.Value
		beforeUnits := datum.Units
		datum.Normalize(dataNormalizer.New())
		afterVal := datum.Value
		afterUnits := datum.Units
		if *beforeVal != *afterVal {
			set["value"] = afterVal
		}
		if *beforeUnits != *afterUnits {
			set["units"] = afterUnits
		}
		identityFields, err = datum.IdentityFields()
		if err != nil {
			return datumID, nil, err
		}
	case ketone.Type:
		var datum ketone.Ketone
		dataBytes, err := bson.Marshal(bsonData)
		if err != nil {
			return datumID, nil, err
		}
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return datumID, nil, err
		}
		beforeVal := datum.Value
		beforeUnits := datum.Units
		datum.Normalize(dataNormalizer.New())
		afterVal := datum.Value
		afterUnits := datum.Units
		if *beforeVal != *afterVal {
			set["value"] = afterVal
		}
		if *beforeUnits != *afterUnits {
			set["units"] = afterUnits
		}
		identityFields, err = datum.IdentityFields()
		if err != nil {
			return datumID, nil, err
		}
	case continuous.Type:
		var datum continuous.Continuous
		dataBytes, err := bson.Marshal(bsonData)
		if err != nil {
			return datumID, nil, err
		}
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return datumID, nil, err
		}
		// NOTE: applies to any type that has a `Glucose` property
		// we need to normalise so that we can get the correct `Units`` and `Value`` precsion that we would if ingested via the platform.
		// as these are both being used in the hash calc via the IdentityFields we want to persist these changes if they are infact updated.
		beforeVal := datum.Value
		beforeUnits := datum.Units
		datum.Normalize(dataNormalizer.New())
		afterVal := datum.Value
		afterUnits := datum.Units
		if *beforeVal != *afterVal {
			set["value"] = afterVal
		}
		if *beforeUnits != *afterUnits {
			set["units"] = afterUnits
		}
		identityFields, err = datum.IdentityFields()
		if err != nil {
			return datumID, nil, err
		}
	default:
		var datum types.Base
		dataBytes, err := bson.Marshal(bsonData)
		if err != nil {
			return datumID, nil, err
		}
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return datumID, nil, err
		}
		identityFields, err = datum.IdentityFields()
		if err != nil {
			return datumID, nil, err
		}
	}

	hash, err := deduplicator.GenerateIdentityHash(identityFields)
	if err != nil {
		return datumID, nil, err
	}

	set["_deduplicator"] = bson.M{"hash": hash}

	updates = append(updates, bson.M{"$set": set})
	if rename != nil {
		log.Printf("rename %v", rename)
		updates = append(updates, bson.M{"$rename": rename})
	}
	if len(updates) != 1 {
		log.Printf("datum updates %d", len(updates))
	}
	return datumID, updates, nil
}
