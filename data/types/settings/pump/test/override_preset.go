package test

import (
	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataTypesSettingsPump "github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomOverridePreset(unitsBloodGlucose *string) *dataTypesSettingsPump.OverridePreset {
	datum := dataTypesSettingsPump.NewOverridePreset()
	datum.Abbreviation = pointer.FromString(RandomAbbreviation())
	datum.Duration = pointer.FromInt(RandomDuration())
	datum.BloodGlucoseTarget = dataBloodGlucoseTest.RandomTarget(unitsBloodGlucose)
	datum.BasalRateScaleFactor = pointer.FromFloat64(RandomBasalRateScaleFactor())
	datum.CarbohydrateRatioScaleFactor = pointer.FromFloat64(RandomCarbohydrateRatioScaleFactor())
	datum.InsulinSensitivityScaleFactor = pointer.FromFloat64(RandomInsulinSensitivityScaleFactor())
	return datum
}

func RandomAbbreviation() string {
	return test.RandomStringFromRange(1, dataTypesSettingsPump.AbbreviationLengthMaximum)
}

func RandomDuration() int {
	return test.RandomIntFromRange(dataTypesSettingsPump.DurationMinimum, dataTypesSettingsPump.DurationMaximum)
}

func RandomBasalRateScaleFactor() float64 {
	return test.RandomFloat64FromRange(dataTypesSettingsPump.BasalRateScaleFactorMinimum, dataTypesSettingsPump.BasalRateScaleFactorMaximum)
}

func RandomCarbohydrateRatioScaleFactor() float64 {
	return test.RandomFloat64FromRange(dataTypesSettingsPump.CarbohydrateRatioScaleFactorMinimum, dataTypesSettingsPump.CarbohydrateRatioScaleFactorMaximum)
}

func RandomInsulinSensitivityScaleFactor() float64 {
	return test.RandomFloat64FromRange(dataTypesSettingsPump.InsulinSensitivityScaleFactorMinimum, dataTypesSettingsPump.InsulinSensitivityScaleFactorMaximum)
}

func CloneOverridePreset(datum *dataTypesSettingsPump.OverridePreset) *dataTypesSettingsPump.OverridePreset {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsPump.NewOverridePreset()
	clone.Abbreviation = pointer.CloneString(datum.Abbreviation)
	clone.Duration = pointer.CloneInt(datum.Duration)
	clone.BloodGlucoseTarget = dataBloodGlucoseTest.CloneTarget(datum.BloodGlucoseTarget)
	clone.BasalRateScaleFactor = pointer.CloneFloat64(datum.BasalRateScaleFactor)
	clone.CarbohydrateRatioScaleFactor = pointer.CloneFloat64(datum.CarbohydrateRatioScaleFactor)
	clone.InsulinSensitivityScaleFactor = pointer.CloneFloat64(datum.InsulinSensitivityScaleFactor)
	return clone
}

func NewObjectFromOverridePreset(datum *dataTypesSettingsPump.OverridePreset, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Abbreviation != nil {
		object["abbreviation"] = test.NewObjectFromString(*datum.Abbreviation, objectFormat)
	}
	if datum.Duration != nil {
		object["duration"] = test.NewObjectFromInt(*datum.Duration, objectFormat)
	}
	if datum.BloodGlucoseTarget != nil {
		object["bgTarget"] = dataBloodGlucoseTest.NewObjectFromTarget(datum.BloodGlucoseTarget, objectFormat)
	}
	if datum.BasalRateScaleFactor != nil {
		object["basalRateScaleFactor"] = test.NewObjectFromFloat64(*datum.BasalRateScaleFactor, objectFormat)
	}
	if datum.CarbohydrateRatioScaleFactor != nil {
		object["carbRatioScaleFactor"] = test.NewObjectFromFloat64(*datum.CarbohydrateRatioScaleFactor, objectFormat)
	}
	if datum.InsulinSensitivityScaleFactor != nil {
		object["insulinSensitivityScaleFactor"] = test.NewObjectFromFloat64(*datum.InsulinSensitivityScaleFactor, objectFormat)
	}
	return object
}

func RandomOverridePresetMap(unitsBloodGlucose *string) *dataTypesSettingsPump.OverridePresetMap {
	datumMap := dataTypesSettingsPump.NewOverridePresetMap()
	for count := test.RandomIntFromRange(2, 3); count > 0; count-- {
		datumMap.Set(RandomOverridePresetName(), RandomOverridePreset(unitsBloodGlucose))
	}
	return datumMap
}

func RandomOverridePresetName() string {
	return test.RandomStringFromRange(1, dataTypesSettingsPump.OverridePresetNameLengthMaximum)
}

func CloneOverridePresetMap(datumMap *dataTypesSettingsPump.OverridePresetMap) *dataTypesSettingsPump.OverridePresetMap {
	if datumMap == nil {
		return nil
	}
	clone := dataTypesSettingsPump.NewOverridePresetMap()
	for datumName, datum := range *datumMap {
		clone.Set(datumName, CloneOverridePreset(datum))
	}
	return clone
}

func NewObjectFromOverridePresetMap(datumMap *dataTypesSettingsPump.OverridePresetMap, objectFormat test.ObjectFormat) map[string]interface{} {
	if datumMap == nil {
		return nil
	}
	object := map[string]interface{}{}
	for datumName, datum := range *datumMap {
		object[datumName] = NewObjectFromOverridePreset(datum, objectFormat)
	}
	return object
}
