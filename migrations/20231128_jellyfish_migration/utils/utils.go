package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/r3labs/diff/v3"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/blood/glucose"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	"github.com/tidepool-org/platform/data/types/blood/ketone"
	"github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/data/types/calculator"
	"github.com/tidepool-org/platform/data/types/device"
	dataTypesFactory "github.com/tidepool-org/platform/data/types/factory"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	errorsP "github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/metadata"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

// func updateIfExistsPumpSettingsSleepSchedules(bsonData bson.M) (*pump.SleepScheduleMap, error) {
// 	//TODO: currently an array but should be a map for consistency. On pump is "Sleep Schedule 1", "Sleep Schedule 2"
// 	scheduleNames := map[int]string{0: "1", 1: "2"}

// 	if schedules := bsonData["sleepSchedules"]; schedules != nil {
// 		sleepScheduleMap := pump.SleepScheduleMap{}
// 		dataBytes, err := json.Marshal(schedules)
// 		if err != nil {
// 			return nil, err
// 		}
// 		schedulesArray := []*pump.SleepSchedule{}
// 		err = json.Unmarshal(dataBytes, &schedulesArray)
// 		if err != nil {
// 			return nil, err
// 		}
// 		for i, schedule := range schedulesArray {
// 			days := schedule.Days
// 			updatedDays := []string{}
// 			for _, day := range *days {
// 				if !slices.Contains(common.DaysOfWeek(), strings.ToLower(day)) {
// 					return nil, errorsP.Newf("pumpSettings.sleepSchedules has an invalid day of week %s", day)
// 				}
// 				updatedDays = append(updatedDays, strings.ToLower(day))
// 			}
// 			schedule.Days = &updatedDays
// 			sleepScheduleMap[scheduleNames[i]] = schedule
// 		}
// 		//sorts schedules based on day
// 		sleepScheduleMap.Normalize(dataNormalizer.New())
// 		return &sleepScheduleMap, nil
// 	}
// 	return nil, nil
// }

// func pumpSettingsHasBolus(bsonData bson.M) bool {
// 	if bolus := bsonData["bolus"]; bolus != nil {
// 		if _, ok := bolus.(*pump.BolusMap); ok {
// 			return true
// 		}
// 	}
// 	return false
// }

func logDiff(id string, updates interface{}) {
	updatesJSON, _ := json.Marshal(updates)
	if string(updatesJSON) != "[]" {
		f, err := os.OpenFile("diff.log",
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		defer f.Close()
		f.WriteString(fmt.Sprintf(`{"_id":"%s","diff":%s},`, id, string(updatesJSON)))
	}
}

func ApplyBaseChanges(bsonData bson.M) error {
	dType := fmt.Sprintf("%v", bsonData["type"])
	switch dType {
	case pump.Type:
		if boluses := bsonData["bolus"]; boluses != nil {
			bsonData["boluses"] = boluses
			//TODO delete from mongo
			delete(bsonData, "bolus")
		}
	case selfmonitored.Type, ketone.Type, continuous.Type:
		units := fmt.Sprintf("%v", bsonData["units"])
		if units == glucose.MmolL || units == glucose.Mmoll {
			floatStr := fmt.Sprintf("%v", bsonData["value"])
			floatParts := strings.Split(floatStr, ".")
			if len(floatParts) == 2 {
				if len(floatParts[1]) > 5 {
					if floatVal, ok := bsonData["value"].(float64); ok {
						mgdlVal := floatVal * glucose.MmolLToMgdLConversionFactor
						intValue := int(mgdlVal/glucose.MmolLToMgdLConversionFactor*glucose.MmolLToMgdLPrecisionFactor + 0.5)
						floatValue := float64(intValue) / glucose.MmolLToMgdLPrecisionFactor
						bsonData["value"] = floatValue
					}
				}
			}
		}
	case calculator.Type:
		if bolus := bsonData["bolus"]; bolus != nil {
			//TODO ignore these, the property is just a pointer to the actual bolus
			delete(bsonData, "bolus")
		}
	}

	if payload := bsonData["payload"]; payload != nil {
		if _, ok := payload.(string); ok {
			dataBytes, err := bson.Marshal(payload)
			if err != nil {
				return err
			}
			var payloadMetadata metadata.Metadata
			err = bson.Unmarshal(dataBytes, &payloadMetadata)
			if err != nil {
				return errorsP.Newf("payload could not be set from %v ", string(dataBytes))
			}
			bsonData["payload"] = &payloadMetadata
		}
	}
	if annotations := bsonData["annotations"]; annotations != nil {
		if _, ok := annotations.(string); ok {
			dataBytes, err := bson.Marshal(annotations)
			if err != nil {
				return err
			}
			var metadataArray metadata.MetadataArray
			if err := bson.Unmarshal(dataBytes, &metadataArray); err != nil {
				return errorsP.Newf("annotations could not be set from %v ", string(dataBytes))
			}
			bsonData["annotations"] = &metadataArray
		}
	}
	return nil
}

func BuildPlatformDatum(objID string, objType string, objectData map[string]interface{}) (*data.Datum, error) {
	parser := structureParser.NewObject(&objectData)
	validator := structureValidator.New()
	normalizer := dataNormalizer.New()

	datum := dataTypesFactory.ParseDatum(parser)
	if datum != nil && *datum != nil {
		(*datum).Validate(validator)
		(*datum).Normalize(normalizer)
	} else {
		return nil, errorsP.Newf("no datum returned for id=[%s]", objID)
	}

	validator.Bool("_active", parser.Bool("_active")).Exists()
	validator.String("_archivedTime", parser.String("_archivedTime"))
	validator.String("_groupId", parser.String("_groupId")).Exists()
	validator.String("_id", parser.String("_id")).Exists()
	validator.String("_userId", parser.String("_userId")).Exists()
	validator.Int("_version", parser.Int("_version")).Exists()
	validator.Int("_schemaVersion", parser.Int("_schemaVersion"))
	validator.Object("_deduplicator", parser.Object("_deduplicator")).Exists()

	validator.String("uploadId", parser.String("uploadId")).Exists()
	validator.String("guid", parser.String("guid"))
	validator.Time("createdTime", parser.Time("createdTime", time.RFC3339Nano)).Exists()
	validator.Time("modifiedTime", parser.Time("modifiedTime", time.RFC3339Nano))

	//parsed but not used in the platform
	//deletes will be created from the diff

	switch objType {
	case continuous.Type:
		validator.String("subType", parser.String("subType"))
	case bolus.Type:
		validator.String("deliveryContext", parser.String("deliveryContext"))
	case basal.Type:
		validator.Object("suppressed", parser.Object("suppressed"))
		validator.Float64("percent", parser.Float64("percent"))
	case device.Type:
		validator.Object("previous", parser.Object("previous"))
		validator.Int("index", parser.Int("index"))
	}

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

	// TODO set the hash
	// fields, err := (*datum).IdentityFields()
	// if err != nil {
	// 	return nil, errorsP.Wrap(err, "unable to gather identity fields for datum")
	// }

	// hash, err := deduplicator.GenerateIdentityHash(fields)
	// if err != nil {
	// 	return nil, errorsP.Wrap(err, "unable to generate identity hash for datum")
	// }

	// deduplicator := (*datum).DeduplicatorDescriptor()
	// if deduplicator == nil {
	// 	deduplicator = data.NewDeduplicatorDescriptor()
	// }
	// deduplicator.Hash = pointer.FromString(hash)

	// (*datum).SetDeduplicatorDescriptor(deduplicator)

	return datum, nil
}

func GetDatumChanges(id string, datum interface{}, original map[string]interface{}, logging bool) ([]bson.M, error) {

	outgoingJSONData, err := json.Marshal(datum)
	if err != nil {
		return nil, err
	}

	processedObject := map[string]interface{}{}
	if err := json.Unmarshal(outgoingJSONData, &processedObject); err != nil {
		return nil, err
	}

	// these are extras that we want to leave on the
	// original object so don't compare
	notRequired := []string{
		"_active",
		"_archivedTime",
		"_deduplicator",
		"_groupId",
		"_id",
		"_schemaVersion",
		"_userId",
		"_version",
		"createdTime",
		"guid",
		"modifiedTime",
		"uploadId",
	}

	for _, key := range notRequired {
		delete(original, key)
		delete(processedObject, key)
	}

	changelog, err := diff.Diff(original, processedObject, diff.StructMapKeySupport())
	if err != nil {
		return nil, err
	}

	set := bson.M{}
	unset := bson.M{}

	// ["path","to","change"]
	// {path: {to: {change: true}}}
	var getValue = func(path []string, val interface{}) interface{} {
		if len(path) == 1 {
			return val
		} else if len(path) == 2 {
			return bson.M{path[1]: val}
		}
		return bson.M{path[1]: bson.M{path[2]: val}}
	}

	for _, change := range changelog.FilterOut([]string{"payload"}) {
		switch change.Type {
		case diff.CREATE, diff.UPDATE:
			set[change.Path[0]] = getValue(change.Path, change.To)
		case diff.DELETE:
			unset[change.Path[0]] = getValue(change.Path, "")
		}
	}

	difference := []bson.M{}
	if len(set) > 0 {
		difference = append(difference, bson.M{"$set": set})
	}
	if len(unset) > 0 {
		difference = append(difference, bson.M{"$unset": unset})
	}
	if logging {
		logDiff(id, difference)
	}
	return difference, nil
}

func ProcessDatum(dataID string, dataType string, bsonData bson.M) ([]bson.M, error) {

	if err := ApplyBaseChanges(bsonData); err != nil {
		return nil, err
	}

	incomingJSONData, err := json.Marshal(bsonData)
	if err != nil {
		return nil, err
	}
	ojbData := map[string]interface{}{}
	if err := json.Unmarshal(incomingJSONData, &ojbData); err != nil {
		return nil, err
	}

	datum, err := BuildPlatformDatum(dataID, dataType, ojbData)
	if err != nil {
		return nil, err
	}

	updates, err := GetDatumChanges(dataID, datum, ojbData, true)
	if err != nil {
		return nil, err
	}
	return updates, nil
}

// func GetDatumUpdates(bsonData bson.M) (string, []bson.M, error) {
// 	updates := []bson.M{}
// 	set := bson.M{}
// 	var rename bson.M
// 	var identityFields []string

// 	datumID, ok := bsonData["_id"].(string)
// 	if !ok {
// 		return "", nil, errorsP.New("cannot get the datum id")
// 	}

// 	datumType, ok := bsonData["type"].(string)
// 	if !ok {
// 		return datumID, nil, errorsP.New("cannot get the datum type")
// 	}

// 	// TODO: based on discussions we want to ensure that these are the correct type
// 	// even though we are not using them for the hash generation
// 	delete(bsonData, "payload")
// 	delete(bsonData, "annotations")

// 	switch datumType {
// 	case basal.Type:
// 		var datum *basal.Basal
// 		dataBytes, err := bson.Marshal(bsonData)
// 		if err != nil {
// 			return datumID, nil, err
// 		}
// 		err = bson.Unmarshal(dataBytes, &datum)
// 		if err != nil {
// 			return datumID, nil, err
// 		}
// 		identityFields, err = datum.IdentityFields()
// 		if err != nil {
// 			return datumID, nil, err
// 		}
// 	case bolus.Type:
// 		var datum *bolus.Bolus
// 		dataBytes, err := bson.Marshal(bsonData)
// 		if err != nil {
// 			return datumID, nil, err
// 		}
// 		err = bson.Unmarshal(dataBytes, &datum)
// 		if err != nil {
// 			return datumID, nil, err
// 		}
// 		identityFields, err = datum.IdentityFields()
// 		if err != nil {
// 			return datumID, nil, err
// 		}
// 	case device.Type:
// 		var datum bolus.Bolus
// 		dataBytes, err := bson.Marshal(bsonData)
// 		if err != nil {
// 			return datumID, nil, err
// 		}
// 		err = bson.Unmarshal(dataBytes, &datum)
// 		if err != nil {
// 			return datumID, nil, err
// 		}
// 		identityFields, err = datum.IdentityFields()
// 		if err != nil {
// 			return datumID, nil, err
// 		}
// 	case pump.Type:
// 		var datum types.Base
// 		dataBytes, err := bson.Marshal(bsonData)
// 		if err != nil {
// 			return datumID, nil, err
// 		}
// 		err = bson.Unmarshal(dataBytes, &datum)
// 		if err != nil {
// 			return datumID, nil, err
// 		}
// 		identityFields, err = datum.IdentityFields()
// 		if err != nil {
// 			return datumID, nil, err
// 		}

// 		if pumpSettingsHasBolus(bsonData) {
// 			rename = bson.M{"bolus": "boluses"}
// 		}

// 		sleepSchedules, err := updateIfExistsPumpSettingsSleepSchedules(bsonData)
// 		if err != nil {
// 			return datumID, nil, err
// 		} else if sleepSchedules != nil {
// 			set["sleepSchedules"] = sleepSchedules
// 		}
// 	case selfmonitored.Type:
// 		var datum selfmonitored.SelfMonitored
// 		dataBytes, err := bson.Marshal(bsonData)
// 		if err != nil {
// 			return datumID, nil, err
// 		}
// 		err = bson.Unmarshal(dataBytes, &datum)
// 		if err != nil {
// 			return datumID, nil, err
// 		}
// 		beforeVal := datum.Value
// 		beforeUnits := datum.Units
// 		datum.Normalize(dataNormalizer.New())
// 		afterVal := datum.Value
// 		afterUnits := datum.Units
// 		if *beforeVal != *afterVal {
// 			set["value"] = afterVal
// 		}
// 		if *beforeUnits != *afterUnits {
// 			set["units"] = afterUnits
// 		}
// 		identityFields, err = datum.IdentityFields()
// 		if err != nil {
// 			return datumID, nil, err
// 		}
// 	case ketone.Type:
// 		var datum ketone.Ketone
// 		dataBytes, err := bson.Marshal(bsonData)
// 		if err != nil {
// 			return datumID, nil, err
// 		}
// 		err = bson.Unmarshal(dataBytes, &datum)
// 		if err != nil {
// 			return datumID, nil, err
// 		}
// 		beforeVal := datum.Value
// 		beforeUnits := datum.Units
// 		datum.Normalize(dataNormalizer.New())
// 		afterVal := datum.Value
// 		afterUnits := datum.Units
// 		if *beforeVal != *afterVal {
// 			set["value"] = afterVal
// 		}
// 		if *beforeUnits != *afterUnits {
// 			set["units"] = afterUnits
// 		}
// 		identityFields, err = datum.IdentityFields()
// 		if err != nil {
// 			return datumID, nil, err
// 		}
// 	case continuous.Type:
// 		var datum continuous.Continuous
// 		dataBytes, err := bson.Marshal(bsonData)
// 		if err != nil {
// 			return datumID, nil, err
// 		}
// 		err = bson.Unmarshal(dataBytes, &datum)
// 		if err != nil {
// 			return datumID, nil, err
// 		}
// 		// NOTE: applies to any type that has a `Glucose` property
// 		// we need to normalise so that we can get the correct `Units`` and `Value`` precsion that we would if ingested via the platform.
// 		// as these are both being used in the hash calc via the IdentityFields we want to persist these changes if they are infact updated.
// 		beforeVal := datum.Value
// 		beforeUnits := datum.Units
// 		datum.Normalize(dataNormalizer.New())
// 		afterVal := datum.Value
// 		afterUnits := datum.Units
// 		if *beforeVal != *afterVal {
// 			set["value"] = afterVal
// 		}
// 		if *beforeUnits != *afterUnits {
// 			set["units"] = afterUnits
// 		}
// 		identityFields, err = datum.IdentityFields()
// 		if err != nil {
// 			return datumID, nil, err
// 		}
// 	default:
// 		var datum types.Base
// 		dataBytes, err := bson.Marshal(bsonData)
// 		if err != nil {
// 			return datumID, nil, err
// 		}
// 		err = bson.Unmarshal(dataBytes, &datum)
// 		if err != nil {
// 			return datumID, nil, err
// 		}
// 		identityFields, err = datum.IdentityFields()
// 		if err != nil {
// 			return datumID, nil, err
// 		}
// 	}

// 	hash, err := deduplicator.GenerateIdentityHash(identityFields)
// 	if err != nil {
// 		return datumID, nil, err
// 	}

// 	set["_deduplicator"] = bson.M{"hash": hash}

// 	updates = append(updates, bson.M{"$set": set})
// 	if rename != nil {
// 		log.Printf("rename %v", rename)
// 		updates = append(updates, bson.M{"$rename": rename})
// 	}
// 	if len(updates) != 1 {
// 		log.Printf("datum updates %d", len(updates))
// 	}
// 	return datumID, updates, nil
// }
