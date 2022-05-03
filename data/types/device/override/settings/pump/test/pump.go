package test

import (
	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataTypesDeviceOverrideSettingsPump "github.com/tidepool-org/platform/data/types/device/override/settings/pump"
	dataTypesDeviceTest "github.com/tidepool-org/platform/data/types/device/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomPump(unitsBloodGlucose *string) *dataTypesDeviceOverrideSettingsPump.Pump {
	datum := randomPump(unitsBloodGlucose)
	datum.Device = *dataTypesDeviceTest.RandomDevice()
	datum.SubType = "pumpSettingsOverride"
	return datum
}

func RandomPumpForParser(unitsBloodGlucose *string) *dataTypesDeviceOverrideSettingsPump.Pump {
	datum := randomPump(unitsBloodGlucose)
	datum.Device = *dataTypesDeviceTest.RandomDeviceForParser()
	datum.SubType = "pumpSettingsOverride"
	return datum
}

func randomPump(unitsBloodGlucose *string) *dataTypesDeviceOverrideSettingsPump.Pump {
	datum := dataTypesDeviceOverrideSettingsPump.New()
	datum.OverrideType = pointer.FromString(RandomOverrideType())
	if *datum.OverrideType == dataTypesDeviceOverrideSettingsPump.OverrideTypePreset {
		datum.OverridePreset = pointer.FromString(test.RandomStringFromRange(1, dataTypesDeviceOverrideSettingsPump.OverridePresetLengthMaximum))
	}
	datum.Method = pointer.FromString(RandomMethod())
	datum.Duration = pointer.FromInt(test.RandomIntFromRange(dataTypesDeviceOverrideSettingsPump.DurationMinimum, dataTypesDeviceOverrideSettingsPump.DurationMaximum-1))
	if test.RandomBool() {
		datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesDeviceOverrideSettingsPump.DurationMaximum))
	}
	datum.BloodGlucoseTarget = dataBloodGlucoseTest.RandomTarget(unitsBloodGlucose)
	datum.BasalRateScaleFactor = pointer.FromFloat64(RandomBasalRateScaleFactor())
	datum.CarbohydrateRatioScaleFactor = pointer.FromFloat64(RandomCarbohydrateRatioScaleFactor())
	datum.InsulinSensitivityScaleFactor = pointer.FromFloat64(RandomInsulinSensitivityScaleFactor())
	if unitsBloodGlucose != nil {
		datum.Units = RandomUnits()
		datum.Units.BloodGlucose = unitsBloodGlucose
	}
	return datum
}

func RandomOverrideType() string {
	return test.RandomStringFromArray(dataTypesDeviceOverrideSettingsPump.OverrideTypes())
}

func RandomMethod() string {
	return test.RandomStringFromArray(dataTypesDeviceOverrideSettingsPump.Methods())
}

func RandomBasalRateScaleFactor() float64 {
	return test.RandomFloat64FromRange(dataTypesDeviceOverrideSettingsPump.BasalRateScaleFactorMinimum, dataTypesDeviceOverrideSettingsPump.BasalRateScaleFactorMaximum)
}

func RandomCarbohydrateRatioScaleFactor() float64 {
	return test.RandomFloat64FromRange(dataTypesDeviceOverrideSettingsPump.CarbohydrateRatioScaleFactorMinimum, dataTypesDeviceOverrideSettingsPump.CarbohydrateRatioScaleFactorMaximum)
}

func RandomInsulinSensitivityScaleFactor() float64 {
	return test.RandomFloat64FromRange(dataTypesDeviceOverrideSettingsPump.InsulinSensitivityScaleFactorMinimum, dataTypesDeviceOverrideSettingsPump.InsulinSensitivityScaleFactorMaximum)
}

func ClonePump(datum *dataTypesDeviceOverrideSettingsPump.Pump) *dataTypesDeviceOverrideSettingsPump.Pump {
	if datum == nil {
		return nil
	}
	clone := dataTypesDeviceOverrideSettingsPump.New()
	clone.Device = *dataTypesDeviceTest.CloneDevice(&datum.Device)
	clone.OverrideType = pointer.CloneString(datum.OverrideType)
	clone.OverridePreset = pointer.CloneString(datum.OverridePreset)
	clone.Method = pointer.CloneString(datum.Method)
	clone.Duration = pointer.CloneInt(datum.Duration)
	clone.DurationExpected = pointer.CloneInt(datum.DurationExpected)
	clone.BloodGlucoseTarget = dataBloodGlucoseTest.CloneTarget(datum.BloodGlucoseTarget)
	clone.BasalRateScaleFactor = pointer.CloneFloat64(datum.BasalRateScaleFactor)
	clone.CarbohydrateRatioScaleFactor = pointer.CloneFloat64(datum.CarbohydrateRatioScaleFactor)
	clone.InsulinSensitivityScaleFactor = pointer.CloneFloat64(datum.InsulinSensitivityScaleFactor)
	clone.Units = CloneUnits(datum.Units)
	return clone
}

func NewObjectFromPump(datum *dataTypesDeviceOverrideSettingsPump.Pump, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := dataTypesDeviceTest.NewObjectFromDevice(&datum.Device, objectFormat)
	if datum.OverrideType != nil {
		object["overrideType"] = test.NewObjectFromString(*datum.OverrideType, objectFormat)
	}
	if datum.OverridePreset != nil {
		object["overridePreset"] = test.NewObjectFromString(*datum.OverridePreset, objectFormat)
	}
	if datum.Method != nil {
		object["method"] = test.NewObjectFromString(*datum.Method, objectFormat)
	}
	if datum.Duration != nil {
		object["duration"] = test.NewObjectFromInt(*datum.Duration, objectFormat)
	}
	if datum.DurationExpected != nil {
		object["expectedDuration"] = test.NewObjectFromInt(*datum.DurationExpected, objectFormat)
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
	if datum.Units != nil {
		object["units"] = NewObjectFromUnits(datum.Units, objectFormat)
	}
	return object
}
