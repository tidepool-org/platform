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
	datum.Base = *dataTypesTest.RandomBase()
	datum.Type = "cgmSettings"
	datum.FirmwareVersion = pointer.FromString(test.RandomStringFromRange(1, dataTypesSettingsCgm.FirmwareVersionLengthMaximum))
	datum.HardwareVersion = pointer.FromString(test.RandomStringFromRange(1, dataTypesSettingsCgm.HardwareVersionLengthMaximum))
	datum.Manufacturers = pointer.FromStringArray(RandomManufacturersFromRange(1, 3))
	datum.Model = pointer.FromString(test.RandomStringFromRange(1, dataTypesSettingsCgm.ModelLengthMaximum))
	datum.Name = pointer.FromString(test.RandomStringFromRange(1, dataTypesSettingsCgm.NameLengthMaximum))
	datum.SerialNumber = pointer.FromString(test.RandomStringFromRange(1, dataTypesSettingsCgm.SerialNumberLengthMaximum))
	datum.SoftwareVersion = pointer.FromString(test.RandomStringFromRange(1, dataTypesSettingsCgm.SoftwareVersionLengthMaximum))
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
	clone.FirmwareVersion = pointer.CloneString(datum.FirmwareVersion)
	clone.HardwareVersion = pointer.CloneString(datum.HardwareVersion)
	clone.Manufacturers = pointer.CloneStringArray(datum.Manufacturers)
	clone.Model = pointer.CloneString(datum.Model)
	clone.Name = pointer.CloneString(datum.Name)
	clone.SerialNumber = pointer.CloneString(datum.SerialNumber)
	clone.SoftwareVersion = pointer.CloneString(datum.SoftwareVersion)
	clone.TransmitterID = pointer.CloneString(datum.TransmitterID)
	clone.Units = pointer.CloneString(datum.Units)
	clone.RawUnits = pointer.CloneString(datum.RawUnits)
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
