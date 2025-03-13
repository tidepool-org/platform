package test

import (
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataTypesSettingsCgm "github.com/tidepool-org/platform/data/types/settings/cgm"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomRateAlert() *dataTypesSettingsCgm.RateAlert {
	datum := &dataTypesSettingsCgm.RateAlert{}
	datum.Alert = *RandomAlert()
	datum.Rate = pointer.FromFloat64(test.RandomFloat64())
	datum.Units = pointer.FromString(test.RandomStringFromArray(dataBloodGlucose.RateUnits()))
	return datum
}

func CloneRateAlert(datum *dataTypesSettingsCgm.RateAlert) *dataTypesSettingsCgm.RateAlert {
	if datum == nil {
		return nil
	}
	clone := &dataTypesSettingsCgm.RateAlert{}
	clone.Alert = *CloneAlert(&datum.Alert)
	clone.Rate = pointer.CloneFloat64(datum.Rate)
	clone.Units = pointer.CloneString(datum.Units)
	return clone
}

func NewObjectFromRateAlert(datum *dataTypesSettingsCgm.RateAlert, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := NewObjectFromAlert(&datum.Alert, objectFormat)
	if datum.Rate != nil {
		object["rate"] = test.NewObjectFromFloat64(*datum.Rate, objectFormat)
	}
	if datum.Units != nil {
		object["units"] = test.NewObjectFromString(*datum.Units, objectFormat)
	}
	return object
}

func RandomFallAlert() *dataTypesSettingsCgm.FallAlert {
	datum := dataTypesSettingsCgm.NewFallAlert()
	datum.RateAlert = *RandomRateAlert()
	datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesSettingsCgm.FallAlertRateRangeForUnits(datum.Units)))
	return datum
}

func CloneFallAlert(datum *dataTypesSettingsCgm.FallAlert) *dataTypesSettingsCgm.FallAlert {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsCgm.NewFallAlert()
	clone.RateAlert = *CloneRateAlert(&datum.RateAlert)
	return clone
}

func NewObjectFromFallAlert(datum *dataTypesSettingsCgm.FallAlert, objectFormat test.ObjectFormat) map[string]interface{} {
	return NewObjectFromRateAlert(&datum.RateAlert, objectFormat)
}

func RandomRiseAlert() *dataTypesSettingsCgm.RiseAlert {
	datum := dataTypesSettingsCgm.NewRiseAlert()
	datum.RateAlert = *RandomRateAlert()
	datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesSettingsCgm.RiseAlertRateRangeForUnits(datum.Units)))
	return datum
}

func CloneRiseAlert(datum *dataTypesSettingsCgm.RiseAlert) *dataTypesSettingsCgm.RiseAlert {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsCgm.NewRiseAlert()
	clone.RateAlert = *CloneRateAlert(&datum.RateAlert)
	return clone
}

func NewObjectFromRiseAlert(datum *dataTypesSettingsCgm.RiseAlert, objectFormat test.ObjectFormat) map[string]interface{} {
	return NewObjectFromRateAlert(&datum.RateAlert, objectFormat)
}
