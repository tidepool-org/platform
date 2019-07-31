package test

import (
	dataTypesSettingsCgm "github.com/tidepool-org/platform/data/types/settings/cgm"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomSnooze() *dataTypesSettingsCgm.Snooze {
	units := pointer.FromString(test.RandomStringFromArray(dataTypesSettingsCgm.SnoozeUnits()))
	datum := dataTypesSettingsCgm.NewSnooze()
	datum.Duration = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesSettingsCgm.SnoozeDurationRangeForUnits(units)))
	datum.Units = units
	return datum
}

func CloneSnooze(datum *dataTypesSettingsCgm.Snooze) *dataTypesSettingsCgm.Snooze {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsCgm.NewSnooze()
	clone.Duration = pointer.CloneFloat64(datum.Duration)
	clone.Units = pointer.CloneString(datum.Units)
	return clone
}

func NewObjectFromSnooze(datum *dataTypesSettingsCgm.Snooze, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Duration != nil {
		object["duration"] = test.NewObjectFromFloat64(*datum.Duration, objectFormat)
	}
	if datum.Units != nil {
		object["units"] = test.NewObjectFromString(*datum.Units, objectFormat)
	}
	return object
}
