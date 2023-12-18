package utils

import (
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

func updateIfExistsPumpSettingsSleepSchedules(datum *pump.Pump) (*pump.SleepScheduleMap, error) {
	sleepSchedules := datum.SleepSchedules
	if sleepSchedules == nil {
		return nil, nil
	}
	for key := range *sleepSchedules {
		days := (*sleepSchedules)[key].Days
		updatedDays := []string{}
		for _, day := range *days {
			if !slices.Contains(common.DaysOfWeek(), strings.ToLower(day)) {
				return nil, errors.Newf("pumpSettings.sleepSchedules has an invalid day of week %s", day)
			}
			updatedDays = append(updatedDays, strings.ToLower(day))
		}
		(*sleepSchedules)[key].Days = &updatedDays
	}
	//sorts schedules based on day
	sleepSchedules.Normalize(normalizer.New())
	return sleepSchedules, nil

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
	datumID := ""
	var identityFields []string

	// while doing test runs
	var errorDebug = func(err error) (string, bson.M, error) {
		log.Printf("[%s] error [%s] creating hash for datum %v", datumID, err, bsonData)
		return datumID, nil, err
	}

	datumType, ok := bsonData["type"].(string)
	if !ok {
		return errorDebug(errors.New("cannot get the datum type"))
	}

	dataBytes, err := bson.Marshal(bsonData)
	if err != nil {
		return errorDebug(err)
	}

	switch datumType {
	case basal.Type:
		var datum *basal.Basal
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return errorDebug(err)
		}
		datumID = *datum.ID
		identityFields, err = datum.IdentityFields()
		log.Printf("basal %s id  %v", datumID, identityFields)
		if err != nil {
			return errorDebug(err)
		}
	case bolus.Type:
		var datum *bolus.Bolus
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return errorDebug(err)
		}
		datumID = *datum.ID
		identityFields, err = datum.IdentityFields()
		log.Printf("bolus %s id  %v", datumID, identityFields)
		if err != nil {
			return errorDebug(err)
		}
	case device.Type:
		var datum *bolus.Bolus
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return errorDebug(err)
		}
		datumID = *datum.ID
		identityFields, err = datum.IdentityFields()
		log.Printf("device %s id  %v", datumID, identityFields)
		if err != nil {
			return errorDebug(err)
		}
	case pump.Type:
		var datum *pump.Pump
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return errorDebug(err)
		}
		datumID = *datum.ID
		identityFields, err = datum.IdentityFields()
		log.Printf("pump %s id  %v", datumID, identityFields)
		if err != nil {
			return errorDebug(err)
		}

		boluses, err := updateIfExistsPumpSettingsBolus(bsonData)
		if err != nil {
			return errorDebug(err)
		} else if boluses != nil {
			updates["boluses"] = boluses
		}

		sleepSchedules, err := updateIfExistsPumpSettingsSleepSchedules(datum)
		if err != nil {
			return errorDebug(err)
		} else if sleepSchedules != nil {
			updates["sleepSchedules"] = sleepSchedules
		}

	case selfmonitored.Type:
		var datum *selfmonitored.SelfMonitored
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return errorDebug(err)
		}
		datumID = *datum.ID
		if *datum.Units != glucose.MgdL && *datum.Units != glucose.Mgdl {
			// NOTE: we need to ensure the same precision for the
			// converted value as it is used to calculate the hash
			val := GetBGValuePlatformPrecision(*datum.Value)
			datum.Value = &val
		}
		identityFields, err = datum.IdentityFields()

		log.Printf("smbg %s id  %v", datumID, identityFields)
		if err != nil {
			return errorDebug(err)
		}
	case ketone.Type:
		var datum *ketone.Ketone
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return errorDebug(err)
		}
		datumID = *datum.ID
		if *datum.Units != glucose.MgdL && *datum.Units != glucose.Mgdl {
			// NOTE: we need to ensure the same precision for the
			// converted value as it is used to calculate the hash
			val := GetBGValuePlatformPrecision(*datum.Value)
			datum.Value = &val
		}
		identityFields, err = datum.IdentityFields()
		log.Printf("ketone %s id  %v", datumID, identityFields)
		if err != nil {
			return errorDebug(err)
		}
	case continuous.Type:
		var datum *continuous.Continuous
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return errorDebug(err)
		}
		datumID = *datum.ID
		if *datum.Units != glucose.MgdL && *datum.Units != glucose.Mgdl {
			// NOTE: we need to ensure the same precision for the
			// converted value as it is used to calculate the hash
			val := GetBGValuePlatformPrecision(*datum.Value)
			datum.Value = &val
		}
		identityFields, err = datum.IdentityFields()
		log.Printf("cbg %s id  %v", datumID, identityFields)
		if err != nil {
			return errorDebug(err)
		}
	default:
		var datum *types.Base
		err = bson.Unmarshal(dataBytes, &datum)
		if err != nil {
			return errorDebug(err)
		}
		datumID = *datum.ID
		identityFields, err = datum.IdentityFields()

		log.Printf("default %s id  %v", datumID, identityFields)
		if err != nil {
			return errorDebug(err)
		}
	}

	hash, err := deduplicator.GenerateIdentityHash(identityFields)
	if err != nil {
		return errorDebug(err)
	}
	updates["_deduplicator"] = bson.M{"hash": hash}

	log.Printf("datum %s updates  %v", datumID, updates)

	return datumID, updates, nil
}
