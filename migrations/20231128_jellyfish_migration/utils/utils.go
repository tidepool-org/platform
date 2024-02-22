package utils

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/r3labs/diff/v3"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	"github.com/tidepool-org/platform/data/types/blood/ketone"
	"github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/data/types/calculator"
	"github.com/tidepool-org/platform/data/types/common"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/device/alarm"
	"github.com/tidepool-org/platform/data/types/device/reservoirchange"
	dataTypesFactory "github.com/tidepool-org/platform/data/types/factory"
	"github.com/tidepool-org/platform/data/types/settings/cgm"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	errorsP "github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func getBGValuePrecision(val interface{}) *float64 {
	floatStr := fmt.Sprintf("%v", val)
	floatParts := strings.Split(floatStr, ".")
	if len(floatParts) == 2 {
		if len(floatParts[1]) > 5 {
			if floatVal, ok := val.(float64); ok {
				mgdlVal := floatVal * glucose.MmolLToMgdLConversionFactor
				intValue := int(mgdlVal/glucose.MmolLToMgdLConversionFactor*glucose.MmolLToMgdLPrecisionFactor + 0.5)
				floatValue := float64(intValue) / glucose.MmolLToMgdLPrecisionFactor
				return &floatValue
			}
		}
	}
	return nil
}

func ApplyBaseChanges(bsonData bson.M, dataType string) error {

	switch dataType {
	case pump.Type:

		if boluses := bsonData["bolus"]; boluses != nil {
			//fix mis-named boluses
			bsonData["boluses"] = boluses
			delete(bsonData, "bolus")
		}
		if schedules := bsonData["sleepSchedules"]; schedules != nil {
			//fix sleepSchedules to be a map
			scheduleNames := map[int]string{0: "1", 1: "2"}
			sleepScheduleMap := pump.SleepScheduleMap{}
			dataBytes, err := json.Marshal(schedules)
			if err != nil {
				return err
			}
			schedulesArray := []*pump.SleepSchedule{}
			err = json.Unmarshal(dataBytes, &schedulesArray)
			if err != nil {
				return err
			}
			for i, schedule := range schedulesArray {
				days := schedule.Days
				updatedDays := []string{}
				for _, day := range *days {
					if !slices.Contains(common.DaysOfWeek(), strings.ToLower(day)) {
						return errorsP.Newf("pumpSettings.sleepSchedules has an invalid day of week %s", day)
					}
					updatedDays = append(updatedDays, strings.ToLower(day))
				}
				schedule.Days = &updatedDays
				sleepScheduleMap[scheduleNames[i]] = schedule
			}
			bsonData["sleepSchedules"] = &sleepScheduleMap
		}

	case selfmonitored.Type, ketone.Type, continuous.Type:
		// fix BG Precision
		units := fmt.Sprintf("%v", bsonData["units"])
		if units == glucose.MmolL || units == glucose.Mmoll {
			if val := getBGValuePrecision(bsonData["value"]); val != nil {
				bsonData["value"] = *val
			}
		}
	case cgm.Type:
		units := fmt.Sprintf("%v", bsonData["units"])
		if units == glucose.MmolL || units == glucose.Mmoll {
			if lowAlerts, ok := bsonData["lowAlerts"].(bson.M); ok {
				if val := getBGValuePrecision(lowAlerts["level"]); val != nil {
					lowAlerts["level"] = *val
					bsonData["lowAlerts"] = lowAlerts
				}
			}
			if highAlerts, ok := bsonData["highAlerts"].(bson.M); ok {
				if val := getBGValuePrecision(highAlerts["level"]); val != nil {
					highAlerts["level"] = *val
					bsonData["highAlerts"] = highAlerts
				}
			}
		}
	case calculator.Type:
		if bolus := bsonData["bolus"]; bolus != nil {
			if _, ok := bolus.(string); ok {
				//if the bolus is a string reference then its ok so leave as is
				delete(bsonData, "bolus")
			}
		}
	case device.Type:
		subType := fmt.Sprintf("%v", bsonData["subType"])
		switch subType {
		case reservoirchange.SubType, alarm.SubType:
			if status := bsonData["status"]; status != nil {
				if statusID, ok := status.(string); ok {
					bsonData["statusId"] = statusID
					delete(bsonData, "status")
				}
			}
		}
	}

	if payload := bsonData["payload"]; payload != nil {

		if m, ok := payload.(bson.M); ok {
			if length := len(m); length == 0 {
				delete(bsonData, "payload")
			}
		}

		if strPayload, ok := payload.(string); ok {
			var payloadMetadata metadata.Metadata
			err := json.Unmarshal(json.RawMessage(strPayload), &payloadMetadata)
			if err != nil {
				return errorsP.Newf("payload could not be set from %s", strPayload)
			}
			bsonData["payload"] = &payloadMetadata
		}

	}
	if annotations := bsonData["annotations"]; annotations != nil {
		if strAnnotations, ok := annotations.(string); ok {
			var metadataArray metadata.MetadataArray
			if err := json.Unmarshal(json.RawMessage(strAnnotations), &metadataArray); err != nil {
				return errorsP.Newf("annotations could not be set from %s", strAnnotations)
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
		(*datum).SetUserID(parser.String("_userId"))
		(*datum).SetCreatedTime(parser.Time("createdTime", time.RFC3339Nano))
		(*datum).Validate(validator)
		(*datum).Normalize(normalizer)
	} else {
		return nil, errorsP.Newf("no datum returned for id=[%s]", objID)
	}

	validator.Bool("_active", parser.Bool("_active")).Exists()
	validator.String("_archivedTime", parser.String("_archivedTime"))
	validator.String("_groupId", parser.String("_groupId")).Exists()
	validator.String("_id", parser.String("_id")).Exists()
	validator.Int("_version", parser.Int("_version")).Exists()
	validator.Int("_schemaVersion", parser.Int("_schemaVersion"))
	validator.Object("_deduplicator", parser.Object("_deduplicator"))

	validator.String("uploadId", parser.String("uploadId")).Exists()
	validator.String("guid", parser.String("guid"))
	validator.Time("modifiedTime", parser.Time("modifiedTime", time.RFC3339Nano))
	validator.Time("localTime", parser.Time("localTime", time.RFC3339Nano))

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
		validator.Float64("rate", parser.Float64("rate"))
	case device.Type:
		validator.Object("previous", parser.Object("previous"))
		validator.String("previousOverride", parser.String("previousOverride"))
		validator.Int("index", parser.Int("index"))
		validator.String("statusId", parser.String("statusId"))
	case calculator.Type:
		validator.Float64("percent", parser.Float64("percent"))
		validator.Float64("rate", parser.Float64("rate"))
		validator.Int("duration", parser.Int("duration"))
		validator.String("bolusId", parser.String("bolusId"))
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

	fields, err := (*datum).IdentityFields()
	if err != nil {
		return nil, errorsP.Wrap(err, "unable to gather identity fields for datum")
	}

	hash, err := deduplicator.GenerateIdentityHash(fields)
	if err != nil {
		return nil, errorsP.Wrap(err, "unable to generate identity hash for datum")
	}

	deduplicator := (*datum).DeduplicatorDescriptor()
	if deduplicator == nil {
		deduplicator = data.NewDeduplicatorDescriptor()
	}
	deduplicator.Hash = pointer.FromString(hash)

	(*datum).SetDeduplicatorDescriptor(deduplicator)

	return datum, nil
}

func GetDatumChanges(id string, datum interface{}, original map[string]interface{}) ([]bson.M, error) {

	outgoingJSONData, err := json.Marshal(datum)
	if err != nil {
		return nil, err
	}

	processedObject := map[string]interface{}{}
	if err := json.Unmarshal(outgoingJSONData, &processedObject); err != nil {
		return nil, err
	}

	if deduplicator := processedObject["deduplicator"]; deduplicator != nil {
		processedObject["_deduplicator"] = deduplicator
	}

	// these are extras that we want to leave on the
	// original object so don't compare
	notRequired := []string{
		"_active",
		"_archivedTime",
		"_groupId",
		"_id",
		"_schemaVersion",
		"_userId",
		"_version",
		"createdTime",
		"guid",
		"modifiedTime",
		"uploadId",
		"deduplicator",
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

	for _, change := range changelog.FilterOut([]string{"payload"}) {
		switch change.Type {
		case diff.CREATE, diff.UPDATE:
			set[strings.Join(change.Path, ".")] = change.To
		case diff.DELETE:
			unset[strings.Join(change.Path, ".")] = ""
		}
	}

	difference := []bson.M{}
	if len(set) > 0 {
		difference = append(difference, bson.M{"$set": set})
	}
	if len(unset) > 0 {
		difference = append(difference, bson.M{"$unset": unset})
	}
	return difference, nil
}

func ProcessDatum(dataID string, dataType string, bsonData bson.M) ([]bson.M, error) {

	if err := ApplyBaseChanges(bsonData, dataType); err != nil {
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

	updates, err := GetDatumChanges(dataID, datum, ojbData)
	if err != nil {
		return nil, err
	}
	return updates, nil
}
