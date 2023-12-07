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
		if t := time.Unix(0, ms*int64(time.Millisecond)); !t.IsZero() {
			return t.Format(types.TimeFormat), nil
		}
	}
	log.Printf("invalid data %#v", bsonData)
	return "", errors.Newf("%s is missing", fieldName)
}

func datumHash(bsonData bson.M) (string, error) {

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
