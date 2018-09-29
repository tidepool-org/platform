package test

import (
	dataTypesSettingsCgm "github.com/tidepool-org/platform/data/types/settings/cgm"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomDurationAlert() *dataTypesSettingsCgm.DurationAlert {
	datum := &dataTypesSettingsCgm.DurationAlert{}
	datum.Alert = *RandomAlert()
	datum.Duration = pointer.FromFloat64(test.RandomFloat64())
	datum.Units = pointer.FromString(test.RandomStringFromArray(dataTypesSettingsCgm.DurationAlertUnits()))
	return datum
}

func CloneDurationAlert(datum *dataTypesSettingsCgm.DurationAlert) *dataTypesSettingsCgm.DurationAlert {
	if datum == nil {
		return nil
	}
	clone := &dataTypesSettingsCgm.DurationAlert{}
	clone.Alert = *CloneAlert(&datum.Alert)
	clone.Duration = test.CloneFloat64(datum.Duration)
	clone.Units = test.CloneString(datum.Units)
	return clone
}

func NewObjectFromDurationAlert(datum *dataTypesSettingsCgm.DurationAlert, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := NewObjectFromAlert(&datum.Alert, objectFormat)
	if datum.Duration != nil {
		object["duration"] = test.NewObjectFromFloat64(*datum.Duration, objectFormat)
	}
	if datum.Units != nil {
		object["units"] = test.NewObjectFromString(*datum.Units, objectFormat)
	}
	return object
}

func RandomNoDataAlert() *dataTypesSettingsCgm.NoDataAlert {
	datum := dataTypesSettingsCgm.NewNoDataAlert()
	datum.DurationAlert = *RandomDurationAlert()
	datum.Duration = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesSettingsCgm.NoDataAlertDurationRangeForUnits(datum.Units)))
	return datum
}

func CloneNoDataAlert(datum *dataTypesSettingsCgm.NoDataAlert) *dataTypesSettingsCgm.NoDataAlert {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsCgm.NewNoDataAlert()
	clone.DurationAlert = *CloneDurationAlert(&datum.DurationAlert)
	return clone
}

func NewObjectFromNoDataAlert(datum *dataTypesSettingsCgm.NoDataAlert, objectFormat test.ObjectFormat) map[string]interface{} {
	return NewObjectFromDurationAlert(&datum.DurationAlert, objectFormat)
}

func RandomOutOfRangeAlert() *dataTypesSettingsCgm.OutOfRangeAlert {
	datum := dataTypesSettingsCgm.NewOutOfRangeAlert()
	datum.DurationAlert = *RandomDurationAlert()
	datum.Duration = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesSettingsCgm.OutOfRangeAlertDurationRangeForUnits(datum.Units)))
	return datum
}

func CloneOutOfRangeAlert(datum *dataTypesSettingsCgm.OutOfRangeAlert) *dataTypesSettingsCgm.OutOfRangeAlert {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsCgm.NewOutOfRangeAlert()
	clone.DurationAlert = *CloneDurationAlert(&datum.DurationAlert)
	return clone
}

func NewObjectFromOutOfRangeAlert(datum *dataTypesSettingsCgm.OutOfRangeAlert, objectFormat test.ObjectFormat) map[string]interface{} {
	return NewObjectFromDurationAlert(&datum.DurationAlert, objectFormat)
}
