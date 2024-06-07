package test

import (
	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataTypesDeviceOverrideSettingsPump "github.com/tidepool-org/platform/data/types/device/override/settings/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomUnits() *dataTypesDeviceOverrideSettingsPump.Units {
	datum := dataTypesDeviceOverrideSettingsPump.NewUnits()
	datum.BloodGlucose = pointer.FromString(dataBloodGlucoseTest.RandomUnits())
	return datum
}

func CloneUnits(datum *dataTypesDeviceOverrideSettingsPump.Units) *dataTypesDeviceOverrideSettingsPump.Units {
	if datum == nil {
		return nil
	}
	clone := dataTypesDeviceOverrideSettingsPump.NewUnits()
	clone.BloodGlucose = pointer.CloneString(datum.BloodGlucose)
	clone.RawBloodGlucose = pointer.CloneString(datum.RawBloodGlucose)
	return clone
}

func SetUnitsRaw(datum *dataTypesDeviceOverrideSettingsPump.Units, normalized *dataTypesDeviceOverrideSettingsPump.Units) {
	if normalized != nil {
		datum.RawBloodGlucose = pointer.CloneString(normalized.RawBloodGlucose)
	}
}

func NewObjectFromUnits(datum *dataTypesDeviceOverrideSettingsPump.Units, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.BloodGlucose != nil {
		object["bg"] = test.NewObjectFromString(*datum.BloodGlucose, objectFormat)
	}
	return object
}
