package test

import (
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataTypesSettingsPump "github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomSuspendThreshold() *dataTypesSettingsPump.SuspendThreshold {
	datum := dataTypesSettingsPump.NewSuspendThreshold()
	datum.Units = pointer.FromString(test.RandomStringFromArray(dataBloodGlucose.Units()))
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(datum.Units)))
	return datum
}

func CloneSuspendThreshold(datum *dataTypesSettingsPump.SuspendThreshold) *dataTypesSettingsPump.SuspendThreshold {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsPump.NewSuspendThreshold()
	clone.Units = pointer.CloneString(datum.Units)
	clone.Value = pointer.CloneFloat64(datum.Value)
	return clone
}
