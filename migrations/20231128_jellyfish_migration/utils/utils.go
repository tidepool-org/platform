package utils

import (
	"encoding/json"
	"fmt"
	"log"
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

// NOTE: required to ensure consitent precision of bg values in the platform
func getBGValuePrecision(val float64) float64 {
	//log.Printf("In VAL %v", val)
	floatStr := fmt.Sprintf("%v", val)
	if _, floatParts, found := strings.Cut(floatStr, "."); found {
		if len(floatParts) > 5 {
			mgdlVal := val * glucose.MmolLToMgdLConversionFactor
			intValue := int(mgdlVal/glucose.MmolLToMgdLConversionFactor*glucose.MmolLToMgdLPrecisionFactor + 0.5)
			floatValue := float64(intValue) / glucose.MmolLToMgdLPrecisionFactor
			//log.Printf("Out VAL %v", floatValue)
			return floatValue
		}
	}
	return val
}

// func getTarget(bgTarget interface{}) (*glucose.Target, error) {
// 	dataBytes, err := json.Marshal(bgTarget)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var target glucose.Target
// 	err = json.Unmarshal(dataBytes, &target)
// 	if err != nil {
// 		return nil, errorsP.Newf("bgTarget %s", string(dataBytes))
// 	}
// 	return &target, nil
// }

// func setGlucoseTargetPrecision(target *glucose.Target) *glucose.Target {
// 	if bg := target.High; bg != nil {
// 		if val := getBGValuePrecision(*bg); val != nil {
// 			target.High = val
// 		}
// 	}
// 	if bg := target.Low; bg != nil {
// 		if val := getBGValuePrecision(*bg); val != nil {
// 			target.Low = val
// 		}
// 	}
// 	if bg := target.Range; bg != nil {
// 		if val := getBGValuePrecision(*bg); val != nil {
// 			target.Range = val
// 		}
// 	}
// 	if low := target.Target; low != nil {
// 		if val := getBGValuePrecision(*low); val != nil {
// 			target.Target = val
// 		}
// 	}
// 	return target
// }

func updateTragetPrecision(targetObj map[string]interface{}) map[string]interface{} {
	if targetObj["high"] != nil {
		if highVal, ok := targetObj["high"].(float64); ok {
			targetObj["high"] = getBGValuePrecision(highVal)
		}
	}
	if targetObj["low"] != nil {
		if lowVal, ok := targetObj["low"].(float64); ok {
			targetObj["low"] = getBGValuePrecision(lowVal)
		}
	}
	if targetObj["range"] != nil {
		if rangeVal, ok := targetObj["range"].(float64); ok {
			targetObj["range"] = getBGValuePrecision(rangeVal)
		}
	}
	if targetObj["low"] != nil {
		if targetVal, ok := targetObj["target"].(float64); ok {
			targetObj["low"] = getBGValuePrecision(targetVal)
		}
	}
	return targetObj
}

func (b *builder) applyBaseUpdates(incomingObject map[string]interface{}) (map[string]interface{}, error) {

	updatedObject := incomingObject

	switch b.datumType {
	case pump.Type:

		if boluses := updatedObject["bolus"]; boluses != nil {
			// NOTE: fix mis-named boluses which were saved in jellyfish as a `bolus`
			updatedObject["boluses"] = boluses
			delete(updatedObject, "bolus")
		}
		if schedules := updatedObject["sleepSchedules"]; schedules != nil {
			// NOTE: this is to fix sleepSchedules so they are in the required map format
			scheduleNames := map[int]string{0: "1", 1: "2"}
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
			updatedObject["sleepSchedules"] = &sleepScheduleMap
		}
		if bgTargetPhysicalActivity := updatedObject["bgTargetPhysicalActivity"]; bgTargetPhysicalActivity != nil {
			if targetObj, ok := bgTargetPhysicalActivity.(map[string]interface{}); ok {
				updatedObject["bgTargetPhysicalActivity"] = updateTragetPrecision(targetObj)
			}
		}
		if bgTargetPreprandial := updatedObject["bgTargetPreprandial"]; bgTargetPreprandial != nil {
			if targetObj, ok := bgTargetPreprandial.(map[string]interface{}); ok {
				updatedObject["bgTargetPreprandial"] = updateTragetPrecision(targetObj)
			}
		}
		if bgTarget := updatedObject["bgTarget"]; bgTarget != nil {
			//var bgTargetStartArry pump.BloodGlucoseTargetStartArray
			dataBytes, err := json.Marshal(bgTarget)
			if err != nil {
				return nil, err
			}
			log.Printf("## bgTarget %s", string(dataBytes))
			// err = json.Unmarshal(dataBytes, &bgTargetStartArry)
			// if err != nil {
			// 	return nil, errorsP.Newf("bgTarget %s", string(dataBytes))
			// }

			// for i, item := range bgTargetStartArry {
			// 	log.Printf("## bgTargetStartArray %d %v", i, item.Target)

			// 	target := setGlucoseTargetPrecision(&item.Target)

			// 	log.Printf("## updated target %v", *target.Target)

			// 	bgTargetStartArry[i].Target = *target

			// 	log.Printf("## updated target %v", *bgTargetStartArry[i].Target.Target)
			// }

			// updatedObject["bgTarget"] = bgTargetStartArry
		}
		if bgTargets := updatedObject["bgTargets"]; bgTargets != nil {
			//var data pump.BloodGlucoseTargetStartArrayMap
			dataBytes, err := json.Marshal(bgTargets)
			if err != nil {
				return nil, err
			}
			log.Printf("## bgTargets %s", string(dataBytes))
			// err = json.Unmarshal(dataBytes, &data)
			// if err != nil {
			// 	return nil, err
			// }

			// for i, d := range data {
			// 	for x, t := range *d {
			// 		t.Target = *setGlucoseTargetPrecision(&t.Target)
			// 		(*d)[x] = t
			// 	}
			// 	data[i] = d
			// }
			// log.Print("## setting updated targets")
			// updatedObject["bgTargets"] = data
		}
		if overridePresets := updatedObject["overridePresets"]; overridePresets != nil {
			// var overridePresetMap pump.OverridePresetMap
			// dataBytes, err := json.Marshal(overridePresets)
			// if err != nil {
			// 	return nil, err
			// }
			// err = json.Unmarshal(dataBytes, &overridePresetMap)
			// if err != nil {
			// 	return nil, err
			// }
			// for i, p := range overridePresetMap {
			// 	overridePresetMap[i].BloodGlucoseTarget = setGlucoseTargetPrecision(p.BloodGlucoseTarget)
			// }
			// updatedObject["overridePresets"] = &overridePresetMap
		}

	case selfmonitored.Type, ketone.Type, continuous.Type:
		units := fmt.Sprintf("%v", updatedObject["units"])
		if units == glucose.MmolL || units == glucose.Mmoll {
			if bgVal, ok := updatedObject["value"].(float64); ok {
				updatedObject["value"] = getBGValuePrecision(bgVal)
			}
		}
	case cgm.Type:
		units := fmt.Sprintf("%v", updatedObject["units"])
		if units == glucose.MmolL || units == glucose.Mmoll {
			if lowAlerts, ok := updatedObject["lowAlerts"].(bson.M); ok {
				if bgVal, ok := lowAlerts["level"].(float64); ok {
					lowAlerts["level"] = getBGValuePrecision(bgVal)
					updatedObject["lowAlerts"] = lowAlerts
				}
			}
			if highAlerts, ok := updatedObject["highAlerts"].(bson.M); ok {
				if bgVal, ok := highAlerts["level"].(float64); ok {
					highAlerts["level"] = getBGValuePrecision(bgVal)
					updatedObject["highAlerts"] = highAlerts
				}
			}
		}
	case calculator.Type:
		if bolus := updatedObject["bolus"]; bolus != nil {
			// NOTE: we are doing this to ensure that the `bolus` is just a string reference
			if _, ok := bolus.(string); ok {
				delete(updatedObject, "bolus")
			}
		}
		if bgTarget := updatedObject["bgTarget"]; bgTarget != nil {
			if targetObj, ok := bgTarget.(map[string]interface{}); ok {
				updatedObject["bgTarget"] = updateTragetPrecision(targetObj)
			}
		}
		if bgInput, ok := updatedObject["bgInput"].(float64); ok {
			updatedObject["bgInput"] = getBGValuePrecision(bgInput)
		}
	case device.Type:
		subType := fmt.Sprintf("%v", updatedObject["subType"])
		switch subType {
		case reservoirchange.SubType, alarm.SubType:
			// NOTE: we are doing this to ensure that the `status` is just a string reference and then setting the `statusId` with it
			if status := updatedObject["status"]; status != nil {
				if statusID, ok := status.(string); ok {
					updatedObject["statusId"] = statusID
					delete(updatedObject, "status")
				}
			}
		}
	}

	if payload := updatedObject["payload"]; payload != nil {

		if m, ok := payload.(bson.M); ok {
			if length := len(m); length == 0 {
				delete(updatedObject, "payload")
			}
		}

		if strPayload, ok := payload.(string); ok {
			var payloadMetadata metadata.Metadata
			err := json.Unmarshal(json.RawMessage(strPayload), &payloadMetadata)
			if err != nil {
				return nil, errorsP.Newf("payload could not be set from %s", strPayload)
			}
			updatedObject["payload"] = &payloadMetadata
		}

	}
	if annotations := updatedObject["annotations"]; annotations != nil {
		if strAnnotations, ok := annotations.(string); ok {
			var metadataArray metadata.MetadataArray
			if err := json.Unmarshal(json.RawMessage(strAnnotations), &metadataArray); err != nil {
				return nil, errorsP.Newf("annotations could not be set from %s", strAnnotations)
			}
			updatedObject["annotations"] = &metadataArray
		}
	}
	return updatedObject, nil
}

func (b *builder) buildDatum(obj map[string]interface{}) error {
	parser := structureParser.NewObject(&obj)
	validator := structureValidator.New()
	normalizer := dataNormalizer.New()

	datum := dataTypesFactory.ParseDatum(parser)
	if datum != nil && *datum != nil {
		(*datum).SetUserID(parser.String("_userId"))
		(*datum).SetCreatedTime(parser.Time("createdTime", time.RFC3339Nano))
		(*datum).Validate(validator)
		(*datum).Normalize(normalizer)
	} else {
		return errorsP.Newf("no datum returned for id=[%s]", b.datumID)
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

	switch b.datumType {
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
		return err
	}

	if err := validator.Error(); err != nil {
		return err
	}

	if err := normalizer.Error(); err != nil {
		return err
	}

	fields, err := (*datum).IdentityFields()
	if err != nil {
		return errorsP.Wrap(err, "unable to gather identity fields for datum")
	}

	hash, err := deduplicator.GenerateIdentityHash(fields)
	if err != nil {
		return errorsP.Wrap(err, "unable to generate identity hash for datum")
	}

	deduplicator := (*datum).DeduplicatorDescriptor()
	if deduplicator == nil {
		deduplicator = data.NewDeduplicatorDescriptor()
	}
	deduplicator.Hash = pointer.FromString(hash)

	(*datum).SetDeduplicatorDescriptor(deduplicator)

	b.datum = *datum
	return nil
}

func (b *builder) datumChanges(storedObj map[string]interface{}) ([]bson.M, []bson.M, error) {

	datumJSON, err := json.Marshal(b.datum)
	if err != nil {
		return nil, nil, err
	}

	// log.Printf("datum: %s", string(datumJSON))

	datumObject := map[string]interface{}{}
	if err := json.Unmarshal(datumJSON, &datumObject); err != nil {
		return nil, nil, err
	}

	if deduplicator := datumObject["deduplicator"]; deduplicator != nil {
		datumObject["_deduplicator"] = deduplicator
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
		"time",
	}

	for _, key := range notRequired {
		delete(storedObj, key)
		delete(datumObject, key)
	}

	changelog, err := diff.Diff(storedObj, datumObject, diff.StructMapKeySupport())
	if err != nil {
		return nil, nil, err
	}

	applySet := bson.M{}
	revertSet := bson.M{}
	applyUnset := bson.M{}
	revertUnset := bson.M{}

	for _, change := range changelog.FilterOut([]string{"payload"}) {
		switch change.Type {
		case diff.CREATE:
			applySet[strings.Join(change.Path, ".")] = change.To
			revertUnset[strings.Join(change.Path, ".")] = ""
		case diff.UPDATE:
			applySet[strings.Join(change.Path, ".")] = change.To
			revertSet[strings.Join(change.Path, ".")] = change.From
		case diff.DELETE:
			applyUnset[strings.Join(change.Path, ".")] = ""
			revertSet[strings.Join(change.Path, ".")] = change.From
		}
	}

	apply := []bson.M{}
	revert := []bson.M{}
	if len(applySet) > 0 {
		apply = append(apply, bson.M{"$set": applySet})
	}
	if len(revertUnset) > 0 {
		revert = append(revert, bson.M{"$unset": revertUnset})
	}
	if len(applyUnset) > 0 {
		apply = append(apply, bson.M{"$unset": applyUnset})
	}
	if len(revertSet) > 0 {
		revert = append(revert, bson.M{"$set": revertSet})
	}
	return apply, revert, nil
}

type builder struct {
	datumType string
	datumID   string
	datum     data.Datum
}

func ProcessDatum(dataID string, dataType string, bsonData bson.M) ([]bson.M, []bson.M, error) {

	b := &builder{
		datumType: dataType,
		datumID:   dataID,
	}

	storedJSON, err := json.Marshal(bsonData)
	if err != nil {
		return nil, nil, err
	}

	// log.Printf("# FROM BSON %s", string(storedJSON))

	storedData := map[string]interface{}{}
	if err := json.Unmarshal(storedJSON, &storedData); err != nil {
		return nil, nil, err
	}

	updatedData, err := b.applyBaseUpdates(storedData)
	if err != nil {
		return nil, nil, err
	}

	if err := b.buildDatum(updatedData); err != nil {
		return nil, nil, err
	}

	apply, revert, err := b.datumChanges(storedData)
	if err != nil {
		return nil, nil, err
	}
	return apply, revert, nil
}
