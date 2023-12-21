package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

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

func GetBGValuePlatformPrecision(mmolVal float64) float64 {
	if len(fmt.Sprintf("%v", mmolVal)) > 7 {
		mgdlVal := mmolVal * glucose.MmolLToMgdLConversionFactor
		mgdL := glucose.MgdL
		mmolVal = *glucose.NormalizeValueForUnits(&mgdlVal, &mgdL)
	}
	return mmolVal
}

func GetDatumUpdates(bsonData bson.M) (string, bson.M, error) {
	start := time.Now()
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

	switch datumType {
	case basal.Type:
		//log.Printf("updating basal start %s", time.Since(start))
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
		//log.Printf("updating bolus start %s", time.Since(start))
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
		//log.Printf("updating device event start %s", time.Since(start))
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
	case pump.Type:
		//log.Printf("updating pump settings start %s", time.Since(start))
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
		//log.Printf("updating smbg start %s", time.Since(start))
		var datum *selfmonitored.SelfMonitored
		dataBytes, err := bson.Marshal(bsonData)
		if err != nil {
			return datumID, nil, err
		}
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return datumID, nil, err
		}
		if *datum.Units != glucose.MgdL && *datum.Units != glucose.Mgdl {
			// NOTE: we need to ensure the same precision for the
			// converted value as it is used to calculate the hash
			val := GetBGValuePlatformPrecision(*datum.Value)
			datum.Value = &val
		}
		identityFields, err = datum.IdentityFields()
		if err != nil {
			return datumID, nil, err
		}
	case ketone.Type:
		//log.Printf("updating ketone start %s", time.Since(start))
		var datum *ketone.Ketone
		dataBytes, err := bson.Marshal(bsonData)
		if err != nil {
			return datumID, nil, err
		}
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return datumID, nil, err
		}
		if *datum.Units != glucose.MgdL && *datum.Units != glucose.Mgdl {
			// NOTE: we need to ensure the same precision for the
			// converted value as it is used to calculate the hash
			val := GetBGValuePlatformPrecision(*datum.Value)
			datum.Value = &val
		}
		identityFields, err = datum.IdentityFields()
		if err != nil {
			return datumID, nil, err
		}
	case continuous.Type:
		//log.Printf("updating cbg start %s", time.Since(start))
		var datum *continuous.Continuous
		dataBytes, err := bson.Marshal(bsonData)
		if err != nil {
			return datumID, nil, err
		}
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return datumID, nil, err
		}
		if *datum.Units != glucose.MgdL && *datum.Units != glucose.Mgdl {
			// NOTE: we need to ensure the same precision for the
			// converted value as it is used to calculate the hash
			val := GetBGValuePlatformPrecision(*datum.Value)
			datum.Value = &val
		}
		identityFields, err = datum.IdentityFields()
		if err != nil {
			return datumID, nil, err
		}
	default:
		//log.Printf("updating generic start %s", time.Since(start))
		var datum *types.Base
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

	//log.Printf("updates made end %s", time.Since(start))
	//log.Printf("generate hash start %s", time.Since(start))
	hash, err := deduplicator.GenerateIdentityHash(identityFields)
	if err != nil {
		return datumID, nil, err
	}

	//log.Printf("generate hash end %s", time.Since(start))
	set["_deduplicator"] = bson.M{"hash": hash}

	var updates = bson.M{"$set": set}
	if rename != nil {
		updates["$rename"] = rename
	}
	duration := time.Since(start)
	if duration > (time.Millisecond * 3) {
		log.Printf("slow datum [%s] updates took %s", datumType, time.Since(start))
	}
	return datumID, updates, nil
}
