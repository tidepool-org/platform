package test

import (
	"math"

	dataTypesSettingsCgm "github.com/tidepool-org/platform/data/types/settings/cgm"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomFallRateAlertDEPRECATED(units *string) *dataTypesSettingsCgm.FallRateAlertDEPRECATED {
	datum := dataTypesSettingsCgm.NewFallRateAlertDEPRECATED()
	datum.Enabled = pointer.FromBool(test.RandomBool())
	if rates := datum.RatesForUnits(units); len(rates) > 0 {
		datum.Rate = pointer.FromFloat64(test.RandomFloat64FromArray(rates))
	} else {
		datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(-math.MaxFloat64, math.MaxFloat64))
	}
	return datum
}

func CloneFallRateAlertDEPRECATED(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED) *dataTypesSettingsCgm.FallRateAlertDEPRECATED {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsCgm.NewFallRateAlertDEPRECATED()
	clone.Enabled = pointer.CloneBool(datum.Enabled)
	clone.Rate = pointer.CloneFloat64(datum.Rate)
	return clone
}

func RandomRiseRateAlertDEPRECATED(units *string) *dataTypesSettingsCgm.RiseRateAlertDEPRECATED {
	datum := dataTypesSettingsCgm.NewRiseRateAlertDEPRECATED()
	datum.Enabled = pointer.FromBool(test.RandomBool())
	if rates := datum.RatesForUnits(units); len(rates) > 0 {
		datum.Rate = pointer.FromFloat64(test.RandomFloat64FromArray(rates))
	} else {
		datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(-math.MaxFloat64, math.MaxFloat64))
	}
	return datum
}

func CloneRiseRateAlertDEPRECATED(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED) *dataTypesSettingsCgm.RiseRateAlertDEPRECATED {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsCgm.NewRiseRateAlertDEPRECATED()
	clone.Enabled = pointer.CloneBool(datum.Enabled)
	clone.Rate = pointer.CloneFloat64(datum.Rate)
	return clone
}

func RandomRateAlertsDEPRECATED(units *string) *dataTypesSettingsCgm.RateAlertsDEPRECATED {
	datum := dataTypesSettingsCgm.NewRateAlertsDEPRECATED()
	datum.FallRateAlert = RandomFallRateAlertDEPRECATED(units)
	datum.RiseRateAlert = RandomRiseRateAlertDEPRECATED(units)
	return datum
}

func CloneRateAlertsDEPRECATED(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED) *dataTypesSettingsCgm.RateAlertsDEPRECATED {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsCgm.NewRateAlertsDEPRECATED()
	clone.FallRateAlert = CloneFallRateAlertDEPRECATED(datum.FallRateAlert)
	clone.RiseRateAlert = CloneRiseRateAlertDEPRECATED(datum.RiseRateAlert)
	return clone
}
