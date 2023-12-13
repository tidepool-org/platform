package utils

import (
	"fmt"
	"log"
	"slices"
	"strconv"
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

func GetValidatedString(bsonData bson.M, fieldName string) (string, error) {
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

func getValidatedTime(bsonData bson.M, fieldName string) (string, error) {
	if valRaw, ok := bsonData[fieldName]; !ok {
		return "", errors.Newf("%s is missing", fieldName)
	} else if ms, ok := valRaw.(int64); !ok {
		if t := time.UnixMilli(ms); !t.IsZero() {
			return t.Format(types.TimeFormat), nil
		}
	}
	log.Printf("invalid data %#v", bsonData)
	return "", errors.Newf("%s is missing", fieldName)
}

func datumHash_1(bsonData bson.M) (string, error) {

	identityFields := []string{}
	if datumUserID, err := GetValidatedString(bsonData, "_userId"); err != nil {
		log.Printf("invalid data _userId: %#v", bsonData)
		return "", err
	} else {
		identityFields = append(identityFields, datumUserID)
	}
	if deviceID, err := GetValidatedString(bsonData, "deviceId"); err != nil {
		log.Printf("invalid data deviceId: %#v", bsonData)
		return "", err
	} else {
		identityFields = append(identityFields, deviceID)
	}
	if datumTime, err := getValidatedTime(bsonData, "time"); err != nil {
		log.Printf("invalid data time: %#v", bsonData)
		return "", err
	} else {
		identityFields = append(identityFields, datumTime)
	}
	datumType, err := GetValidatedString(bsonData, "type")
	if err != nil {
		log.Printf("invalid data type: %#v", bsonData)
		return "", err
	}
	identityFields = append(identityFields, datumType)

	switch datumType {
	case basal.Type:
		if deliveryType, err := GetValidatedString(bsonData, "deliveryType"); err != nil {
			return "", err
		} else {
			identityFields = append(identityFields, deliveryType)
		}
	case bolus.Type, device.Type:
		if subType, err := GetValidatedString(bsonData, "subType"); err != nil {
			return "", err
		} else {
			identityFields = append(identityFields, subType)
		}
	case selfmonitored.Type, ketone.Type, continuous.Type:
		units, err := GetValidatedString(bsonData, "units")
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
				val = GetBGValuePlatformPrecision(val)
			}
			identityFields = append(identityFields, strconv.FormatFloat(val, 'f', -1, 64))
		}
	}
	return deduplicator.GenerateIdentityHash(identityFields)
}

func datumHash(bsonData bson.M) (string, error) {

	datumType, err := GetValidatedString(bsonData, "type")
	if err != nil {
		log.Printf("invalid data type: %#v", bsonData)
		return "", err
	}
	identityFields := []string{}

	switch datumType {
	case basal.Type:
		var basalDatum *basal.Basal
		bsonBytes, err := bson.Marshal(bsonData)
		if err != nil {
			return "", err
		}
		bson.Unmarshal(bsonBytes, &basalDatum)
		identityFields, err = basalDatum.IdentityFields()
		if err != nil {
			return "", err
		}
	case bolus.Type:
		var bolusDatum *bolus.Bolus
		bsonBytes, err := bson.Marshal(bsonData)
		if err != nil {
			return "", err
		}
		bson.Unmarshal(bsonBytes, &bolusDatum)
		identityFields, err = bolusDatum.IdentityFields()
		if err != nil {
			return "", err
		}
	case device.Type:
		var deviceDatum *device.Device
		bsonBytes, err := bson.Marshal(bsonData)
		if err != nil {
			return "", err
		}
		bson.Unmarshal(bsonBytes, &deviceDatum)
		identityFields, err = deviceDatum.IdentityFields()
		if err != nil {
			return "", err
		}
	case selfmonitored.Type:
		var smbgDatum *selfmonitored.SelfMonitored
		bsonBytes, err := bson.Marshal(bsonData)
		if err != nil {
			return "", err
		}
		bson.Unmarshal(bsonBytes, &smbgDatum)
		if *smbgDatum.Units != glucose.MgdL && *smbgDatum.Units != glucose.Mgdl {
			// NOTE: we need to ensure the same precision for the
			// converted value as it is used to calculate the hash
			val := GetBGValuePlatformPrecision(*smbgDatum.Value)
			smbgDatum.Value = &val
		}
		identityFields, err = smbgDatum.IdentityFields()
		if err != nil {
			return "", err
		}
	case ketone.Type:
		var ketoneDatum *ketone.Ketone
		bsonBytes, err := bson.Marshal(bsonData)
		if err != nil {
			return "", err
		}
		bson.Unmarshal(bsonBytes, &ketoneDatum)
		if *ketoneDatum.Units != glucose.MgdL && *ketoneDatum.Units != glucose.Mgdl {
			// NOTE: we need to ensure the same precision for the
			// converted value as it is used to calculate the hash
			val := GetBGValuePlatformPrecision(*ketoneDatum.Value)
			ketoneDatum.Value = &val
		}

		identityFields, err = ketoneDatum.IdentityFields()
		if err != nil {
			return "", err
		}
	case continuous.Type:
		var cbgDatum *continuous.Continuous
		bsonBytes, err := bson.Marshal(bsonData)
		if err != nil {
			return "", err
		}
		bson.Unmarshal(bsonBytes, &cbgDatum)

		if *cbgDatum.Units != glucose.MgdL && *cbgDatum.Units != glucose.Mgdl {
			// NOTE: we need to ensure the same precision for the
			// converted value as it is used to calculate the hash
			val := GetBGValuePlatformPrecision(*cbgDatum.Value)
			cbgDatum.Value = &val
		}

		identityFields, err = cbgDatum.IdentityFields()
		if err != nil {
			return "", err
		}

	}
	return deduplicator.GenerateIdentityHash(identityFields)
}

func updateIfExistsPumpSettingsSleepSchedules(bsonData bson.M) (*pump.SleepScheduleMap, error) {
	dataType, err := GetValidatedString(bsonData, "type")
	if err != nil {
		return nil, err
	}

	if dataType == pump.Type {
		if sleepSchedules := bsonData["sleepSchedules"]; sleepSchedules != nil {
			schedules, ok := sleepSchedules.(*pump.SleepScheduleMap)
			if !ok {
				return nil, errors.Newf("pumpSettings.sleepSchedules is not the expected type %s", sleepSchedules)
			}
			for key := range *schedules {
				days := (*schedules)[key].Days
				updatedDays := []string{}
				for _, day := range *days {
					if !slices.Contains(common.DaysOfWeek(), strings.ToLower(day)) {
						return nil, errors.Newf("pumpSettings.sleepSchedules has an invalid day of week %s", day)
					}
					updatedDays = append(updatedDays, strings.ToLower(day))
				}
				(*schedules)[key].Days = &updatedDays
			}
			//sorts schedules based on day
			schedules.Normalize(normalizer.New())
			return schedules, nil
		}
	}
	return nil, nil
}

func updateIfExistsPumpSettingsBolus(bsonData bson.M) (*pump.BolusMap, error) {
	dataType, err := GetValidatedString(bsonData, "type")
	if err != nil {
		return nil, err
	}
	if dataType == pump.Type {
		if bolus := bsonData["bolus"]; bolus != nil {
			boluses, ok := bolus.(*pump.BolusMap)
			if !ok {
				return nil, errors.Newf("pumpSettings.bolus is not the expected type %v", bolus)
			}
			return boluses, nil
		}
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

func GetDatumUpdates(bsonData bson.M) (bson.M, error) {
	updates := bson.M{}

	hash, err := datumHash(bsonData)
	if err != nil {
		return nil, err
	}
	updates["_deduplicator"] = bson.M{"hash": hash}

	boluses, err := updateIfExistsPumpSettingsBolus(bsonData)
	if err != nil {
		return nil, err
	} else if boluses != nil {
		updates["boluses"] = boluses
	}

	sleepSchedules, err := updateIfExistsPumpSettingsSleepSchedules(bsonData)
	if err != nil {
		return nil, err
	} else if sleepSchedules != nil {
		updates["sleepSchedules"] = sleepSchedules
	}

	return updates, nil
}
