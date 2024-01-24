package utils

import (
	"encoding/json"
	"log"
	"slices"
	"strings"

	"go.mongodb.org/mongo-driver/bson"

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
