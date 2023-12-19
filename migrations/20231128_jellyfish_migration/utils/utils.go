package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strings"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/data/blood/glucose"
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
	scheduleNames := map[int]string{0: "One", 1: "Two", 2: "Three", 3: "Four", 4: "Five"}

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

func updateIfExistsPumpSettingsBolus(bsonData bson.M) (*pump.BolusMap, error) {
	if bolus := bsonData["bolus"]; bolus != nil {
		boluses, ok := bolus.(*pump.BolusMap)
		if !ok {
			return nil, errors.Newf("data %v is not the expected boluses type", bolus)
		}
		return boluses, nil
	}
	return nil, nil
}

func GetBGValuePlatformPrecision(mmolVal float64) float64 {
	if len(fmt.Sprintf("%v", mmolVal)) > 7 {
		mgdlVal := mmolVal * glucose.MmolLToMgdLConversionFactor
		mgdL := glucose.MgdL
		mmolVal = *glucose.NormalizeValueForUnits(&mgdlVal, &mgdL)
	}
	return mmolVal
}

func GetDatumUpdates(bsonData bson.M) (string, bson.M, error) {
	updates := bson.M{}
	var identityFields []string

	// while doing test runs
	var errorDebug = func(id string, err error) (string, bson.M, error) {
		log.Printf("[%s] error [%s] creating hash for datum %v", id, err, bsonData)
		return id, nil, err
	}

	datumID, ok := bsonData["_id"].(string)
	if !ok {
		return errorDebug("", errors.New("cannot get the datum id"))
	}

	datumType, ok := bsonData["type"].(string)
	if !ok {
		return errorDebug(datumID, errors.New("cannot get the datum type"))
	}

	dataBytes, err := bson.Marshal(bsonData)
	if err != nil {
		return errorDebug(datumID, err)
	}

	switch datumType {
	case basal.Type:
		var datum *basal.Basal
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return errorDebug(datumID, err)
		}
		identityFields, err = datum.IdentityFields()
		if err != nil {
			return errorDebug(datumID, err)
		}
	case bolus.Type:
		var datum *bolus.Bolus
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return errorDebug(datumID, err)
		}
		identityFields, err = datum.IdentityFields()
		if err != nil {
			return errorDebug(datumID, err)
		}
	case device.Type:
		var datum *bolus.Bolus
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return errorDebug(datumID, err)
		}
		identityFields, err = datum.IdentityFields()
		if err != nil {
			return errorDebug(datumID, err)
		}
	case pump.Type:
		var datum *types.Base
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return errorDebug(datumID, err)
		}
		identityFields, err = datum.IdentityFields()
		if err != nil {
			return errorDebug(datumID, err)
		}

		boluses, err := updateIfExistsPumpSettingsBolus(bsonData)
		if err != nil {
			return errorDebug(datumID, err)
		} else if boluses != nil {
			updates["boluses"] = boluses
		}

		sleepSchedules, err := updateIfExistsPumpSettingsSleepSchedules(bsonData)
		if err != nil {
			return errorDebug(datumID, err)
		} else if sleepSchedules != nil {
			updates["sleepSchedules"] = sleepSchedules
		}
	case selfmonitored.Type:
		var datum *selfmonitored.SelfMonitored
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return errorDebug(datumID, err)
		}
		if *datum.Units != glucose.MgdL && *datum.Units != glucose.Mgdl {
			// NOTE: we need to ensure the same precision for the
			// converted value as it is used to calculate the hash
			val := GetBGValuePlatformPrecision(*datum.Value)
			datum.Value = &val
		}
		identityFields, err = datum.IdentityFields()
		if err != nil {
			return errorDebug(datumID, err)
		}
	case ketone.Type:
		var datum *ketone.Ketone
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return errorDebug(datumID, err)
		}
		if *datum.Units != glucose.MgdL && *datum.Units != glucose.Mgdl {
			// NOTE: we need to ensure the same precision for the
			// converted value as it is used to calculate the hash
			val := GetBGValuePlatformPrecision(*datum.Value)
			datum.Value = &val
		}
		identityFields, err = datum.IdentityFields()
		if err != nil {
			return errorDebug(datumID, err)
		}
	case continuous.Type:
		var datum *continuous.Continuous
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return errorDebug(datumID, err)
		}
		if *datum.Units != glucose.MgdL && *datum.Units != glucose.Mgdl {
			// NOTE: we need to ensure the same precision for the
			// converted value as it is used to calculate the hash
			val := GetBGValuePlatformPrecision(*datum.Value)
			datum.Value = &val
		}
		identityFields, err = datum.IdentityFields()
		if err != nil {
			return errorDebug(datumID, err)
		}
	default:
		var datum *types.Base
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return errorDebug(datumID, err)
		}
		identityFields, err = datum.IdentityFields()
		if err != nil {
			return errorDebug(datumID, err)
		}
	}

	hash, err := deduplicator.GenerateIdentityHash(identityFields)
	if err != nil {
		return errorDebug(datumID, err)
	}
	updates["_deduplicator"] = bson.M{"hash": hash}
	return datumID, updates, nil
}
