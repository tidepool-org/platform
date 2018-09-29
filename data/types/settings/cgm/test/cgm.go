package test

import (
	dataTypesSettingsCgm "github.com/tidepool-org/platform/data/types/settings/cgm"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

const CharsetTransmitterID = test.CharsetNumeric + test.CharsetUppercase

func RandomCGM(units *string) *dataTypesSettingsCgm.CGM {
	datum := dataTypesSettingsCgm.New()
	datum.Base = *dataTypesTest.NewBase()
	datum.Type = "cgmSettings"
	datum.Manufacturers = pointer.FromStringArray(RandomManufacturersFromRange(1, 3))
	datum.Model = pointer.FromString(test.RandomStringFromRange(1, dataTypesSettingsCgm.ModelLengthMaximum))
	datum.SerialNumber = pointer.FromString(test.RandomStringFromRange(1, dataTypesSettingsCgm.SerialNumberLengthMaximum))
	datum.TransmitterID = pointer.FromString(test.RandomStringFromRangeAndCharset(5, 6, CharsetTransmitterID))
	datum.Units = units
	datum.DefaultAlerts = RandomAlerts()
	datum.ScheduledAlerts = RandomScheduledAlerts(1, 3)
	datum.HighLevelAlert = RandomHighLevelAlertDEPRECATED(units)
	datum.LowLevelAlert = RandomLowLevelAlertDEPRECATED(units)
	datum.OutOfRangeAlert = RandomOutOfRangeAlertDEPRECATED()
	datum.RateAlerts = RandomRateAlertsDEPRECATED(units)
	return datum
}

func CloneCGM(datum *dataTypesSettingsCgm.CGM) *dataTypesSettingsCgm.CGM {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsCgm.New()
	clone.Base = *dataTypesTest.CloneBase(&datum.Base)
	clone.Manufacturers = test.CloneStringArray(datum.Manufacturers)
	clone.Model = test.CloneString(datum.Model)
	clone.SerialNumber = test.CloneString(datum.SerialNumber)
	clone.TransmitterID = test.CloneString(datum.TransmitterID)
	clone.Units = test.CloneString(datum.Units)
	clone.DefaultAlerts = CloneAlerts(datum.DefaultAlerts)
	clone.ScheduledAlerts = CloneScheduledAlerts(datum.ScheduledAlerts)
	clone.HighLevelAlert = CloneHighLevelAlertDEPRECATED(datum.HighLevelAlert)
	clone.LowLevelAlert = CloneLowLevelAlertDEPRECATED(datum.LowLevelAlert)
	clone.OutOfRangeAlert = CloneOutOfRangeAlertDEPRECATED(datum.OutOfRangeAlert)
	clone.RateAlerts = CloneRateAlertsDEPRECATED(datum.RateAlerts)
	return clone
}

func RandomManufacturersFromRange(minimumLength int, maximumLength int) []string {
	datum := make([]string, test.RandomIntFromRange(minimumLength, maximumLength))
	for index := range datum {
		datum[index] = test.RandomStringFromRange(1, dataTypesSettingsCgm.ManufacturerLengthMaximum)
	}
	return datum
}
