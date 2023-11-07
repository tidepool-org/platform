package test

import (
	"math/rand"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"

	"github.com/tidepool-org/platform/data/types"
	dataTypesBasalTest "github.com/tidepool-org/platform/data/types/basal/test"
	dataTypesInsulinTest "github.com/tidepool-org/platform/data/types/insulin/test"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &types.Meta{
		Type: "pumpSettings",
	}
}

func NewManufacturer(minimumLength int, maximumLength int) string {
	return test.RandomStringFromRange(minimumLength, maximumLength)
}

func NewManufacturers(minimumLength int, maximumLength int) []string {
	result := make([]string, minimumLength+rand.Intn(maximumLength-minimumLength+1))
	for index := range result {
		result[index] = NewManufacturer(1, 100)
	}
	return result
}

func NewPump(unitsBloodGlucose *string) *pump.Pump {
	scheduleName := dataTypesBasalTest.RandomScheduleName()
	datum := pump.New()
	datum.Base = *dataTypesTest.RandomBase()
	datum.Type = "pumpSettings"
	datum.ActiveScheduleName = pointer.FromString(scheduleName)
	datum.AutomatedDelivery = pointer.FromBool(test.RandomBool())
	datum.Basal = NewBasal()
	datum.BasalRateSchedules = pump.NewBasalRateStartArrayMap()
	datum.BasalRateSchedules.Set(scheduleName, NewBasalRateStartArray())
	datum.BloodGlucoseSafetyLimit = pointer.FromFloat64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(unitsBloodGlucose)))
	datum.BloodGlucoseTargetPhysicalActivity = dataBloodGlucoseTest.RandomTarget(unitsBloodGlucose)
	datum.BloodGlucoseTargetPreprandial = dataBloodGlucoseTest.RandomTarget(unitsBloodGlucose)
	datum.BloodGlucoseTargetSchedules = pump.NewBloodGlucoseTargetStartArrayMap()
	datum.BloodGlucoseTargetSchedules.Set(scheduleName, RandomBloodGlucoseTargetStartArray(unitsBloodGlucose))
	datum.Boluses = NewRandomBolusMap(2, 4)
	datum.CarbohydrateRatioSchedules = pump.NewCarbohydrateRatioStartArrayMap()
	datum.CarbohydrateRatioSchedules.Set(scheduleName, NewCarbohydrateRatioStartArray())
	datum.Display = NewDisplay()
	datum.FirmwareVersion = pointer.FromString(test.RandomStringFromRange(1, pump.FirmwareVersionLengthMaximum))
	datum.HardwareVersion = pointer.FromString(test.RandomStringFromRange(1, pump.HardwareVersionLengthMaximum))
	datum.InsulinFormulation = dataTypesInsulinTest.RandomFormulation(3)
	datum.InsulinModel = RandomInsulinModel()
	datum.InsulinSensitivitySchedules = pump.NewInsulinSensitivityStartArrayMap()
	datum.InsulinSensitivitySchedules.Set(scheduleName, NewInsulinSensitivityStartArray(unitsBloodGlucose))
	datum.Manufacturers = pointer.FromStringArray(NewManufacturers(1, 10))
	datum.Model = pointer.FromString(test.RandomStringFromRange(1, 100))
	datum.Name = pointer.FromString(test.RandomStringFromRange(1, 100))
	datum.OverridePresets = RandomOverridePresetMap(unitsBloodGlucose)
	datum.ScheduleTimeZoneOffset = pointer.FromInt(test.RandomIntFromRange(pump.ScheduleTimeZoneOffsetMinimum, pump.ScheduleTimeZoneOffsetMaximum))
	datum.SerialNumber = pointer.FromString(test.RandomStringFromRange(1, 100))
	datum.SoftwareVersion = pointer.FromString(test.RandomStringFromRange(1, pump.SoftwareVersionLengthMaximum))
	datum.Units = RandomUnits(unitsBloodGlucose)
	return datum
}

func ClonePump(datum *pump.Pump) *pump.Pump {
	if datum == nil {
		return nil
	}
	clone := pump.New()
	clone.Base = *dataTypesTest.CloneBase(&datum.Base)
	clone.ActiveScheduleName = pointer.CloneString(datum.ActiveScheduleName)
	clone.AutomatedDelivery = pointer.CloneBool(datum.AutomatedDelivery)
	clone.Basal = CloneBasal(datum.Basal)
	clone.BasalRateSchedule = CloneBasalRateStartArray(datum.BasalRateSchedule)
	clone.BasalRateSchedules = CloneBasalRateStartArrayMap(datum.BasalRateSchedules)
	clone.BloodGlucoseSafetyLimit = pointer.CloneFloat64(datum.BloodGlucoseSafetyLimit)
	clone.BloodGlucoseTargetPhysicalActivity = dataBloodGlucoseTest.CloneTarget(datum.BloodGlucoseTargetPhysicalActivity)
	clone.BloodGlucoseTargetPreprandial = dataBloodGlucoseTest.CloneTarget(datum.BloodGlucoseTargetPreprandial)
	clone.BloodGlucoseTargetSchedule = CloneBloodGlucoseTargetStartArray(datum.BloodGlucoseTargetSchedule)
	clone.BloodGlucoseTargetSchedules = CloneBloodGlucoseTargetStartArrayMap(datum.BloodGlucoseTargetSchedules)
	clone.Bolus = CloneBolus(datum.Bolus)
	clone.Boluses = CloneBolusMap(datum.Boluses)
	clone.CarbohydrateRatioSchedule = CloneCarbohydrateRatioStartArray(datum.CarbohydrateRatioSchedule)
	clone.CarbohydrateRatioSchedules = CloneCarbohydrateRatioStartArrayMap(datum.CarbohydrateRatioSchedules)
	clone.Display = CloneDisplay(datum.Display)
	clone.FirmwareVersion = pointer.CloneString(datum.FirmwareVersion)
	clone.HardwareVersion = pointer.CloneString(datum.HardwareVersion)
	clone.InsulinFormulation = dataTypesInsulinTest.CloneFormulation(datum.InsulinFormulation)
	clone.InsulinModel = CloneInsulinModel(datum.InsulinModel)
	clone.InsulinSensitivitySchedule = CloneInsulinSensitivityStartArray(datum.InsulinSensitivitySchedule)
	clone.InsulinSensitivitySchedules = CloneInsulinSensitivityStartArrayMap(datum.InsulinSensitivitySchedules)
	clone.Manufacturers = pointer.CloneStringArray(datum.Manufacturers)
	clone.Model = pointer.CloneString(datum.Model)
	clone.Name = pointer.CloneString(datum.Name)
	clone.OverridePresets = CloneOverridePresetMap(datum.OverridePresets)
	clone.ScheduleTimeZoneOffset = pointer.CloneInt(datum.ScheduleTimeZoneOffset)
	clone.SerialNumber = pointer.CloneString(datum.SerialNumber)
	clone.SoftwareVersion = pointer.CloneString(datum.SoftwareVersion)
	clone.Units = CloneUnits(datum.Units)
	return clone
}
