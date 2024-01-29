package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	"github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	"github.com/tidepool-org/platform/data/types/blood/ketone"
	"github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/data/types/common"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/errors"
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
					return nil, errors.Newf("pumpSettings.sleepSchedules has an invalid day of week %s", day)
				}
				updatedDays = append(updatedDays, strings.ToLower(day))
			}
			schedule.Days = &updatedDays
			sleepScheduleMap[scheduleNames[i]] = schedule
		}
		//sorts schedules based on day
		sleepScheduleMap.Normalize(normalizer.New())
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

func logUpdates(datumArray []data.Datum) {
	f, err := os.OpenFile("updated.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer f.Close()
	for _, v := range datumArray {
		fields, _ := v.IdentityFields()
		f.WriteString(fmt.Sprintf("%s - %s \n", v.GetType(), fields))
	}

}

func ProcessData(bsonDataArray []bson.M) ([]data.Datum, []error) {

	start := time.Now()

	datumArray := []data.Datum{}

	jsonData, _ := json.Marshal(bsonDataArray)
	converted := []map[string]interface{}{}
	preprocessedDatumArray := []interface{}{}

	json.Unmarshal(jsonData, &converted)

	for _, item := range converted {

		dType := fmt.Sprintf("%v", item["type"])

		// FIXES
		if dType == pump.Type {
			if boluses := item["bolus"]; boluses != nil {
				item["boluses"] = boluses
				delete(item, "bolus")
			}
		}

		if payload := item["payload"]; payload != nil {
			if payloadMetadata, ok := payload.(*metadata.Metadata); ok {
				item["payload"] = payloadMetadata
			}
		}
		if annotations := item["annotations"]; annotations != nil {
			if metadataArray, ok := annotations.(*metadata.MetadataArray); ok {
				item["annotations"] = metadataArray
			}
		}

		//log.Printf("[%d] [%v]\n\n", i, item)

		//
		// switch dType {
		// case pump.Type:
		// 	var datum *pump.Pump
		// 	dataBytes, err := bson.Marshal(item)
		// 	if err != nil {
		// 		log.Printf("%s %s", dType, err)
		// 		break
		// 	}
		// 	err = bson.Unmarshal(dataBytes, &datum)
		// 	if err != nil {
		// 		log.Printf("%s %s", dType, err)
		// 		break
		// 	}
		// 	datumArray = append(datumArray, datum)
		// case continuous.Type:
		// 	var datum *continuous.Continuous
		// 	dataBytes, err := bson.Marshal(item)
		// 	if err != nil {
		// 		log.Printf("%s %s", dType, err)
		// 		break
		// 	}
		// 	err = bson.Unmarshal(dataBytes, &datum)
		// 	if err != nil {
		// 		log.Printf("%s %s", dType, err)
		// 		break
		// 	}
		// 	datumArray = append(datumArray, datum)

		// case selfmonitored.Type:
		// 	var datum *selfmonitored.SelfMonitored
		// 	dataBytes, err := bson.Marshal(item)
		// 	if err != nil {
		// 		log.Printf("%s %s", dType, err)
		// 		break
		// 	}
		// 	err = bson.Unmarshal(dataBytes, &datum)
		// 	if err != nil {
		// 		log.Printf("%s %s", dType, err)
		// 		break
		// 	}
		// 	datumArray = append(datumArray, datum)
		// default:

		// }
		preprocessedDatumArray = append(preprocessedDatumArray, item)
	}

	errs := []error{}
	parser := structureParser.NewArray(&preprocessedDatumArray)
	validator := structureValidator.New()
	normalizer := dataNormalizer.New()

	for _, reference := range parser.References() {
		if datum := dataTypesFactory.ParseDatum(parser.WithReferenceObjectParser(reference)); datum != nil && *datum != nil {
			(*datum).Validate(validator.WithReference(strconv.Itoa(reference)))
			(*datum).Normalize(normalizer.WithReference(strconv.Itoa(reference)))
			datumArray = append(datumArray, *datum)
		}
	}

	// for i, datum := range datumArray {
	// 	(datum).Validate(validator.WithReference(strconv.Itoa(i)))
	// 	(datum).Normalize(normalizer.WithReference(strconv.Itoa(i)))
	// }

	// parser.NotParsed()

	if err := parser.Error(); err != nil {
		log.Println("Parser errors")
		errs = append(errs, err)
	}

	if err := validator.Error(); err != nil {
		log.Println("Validator errors")
		errs = append(errs, err)
	}

	if err := normalizer.Error(); err != nil {
		log.Println("Normalizer errors")
		errs = append(errs, err)
	}

	// for debug
	logUpdates(datumArray)

	log.Printf("fetched [%d] processed [%d]  in [%s] [%t]", len(bsonDataArray), len(datumArray), time.Since(start).Truncate(time.Millisecond), len(errs) > 0)

	return datumArray, errs
}

func GetDatumUpdates(bsonData bson.M) (string, []bson.M, error) {
	updates := []bson.M{}
	set := bson.M{}
	var rename bson.M
	var identityFields []string

	datumID, ok := bsonData["_id"].(string)
	if !ok {
		return "", nil, errors.New("cannot get the datum id")
	}

	datumType, ok := bsonData["type"].(string)
	if !ok {
		return datumID, nil, errors.New("cannot get the datum type")
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
		datum.Normalize(normalizer.New())
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
		datum.Normalize(normalizer.New())
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
		datum.Normalize(normalizer.New())
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
