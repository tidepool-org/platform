package test

import (
	dataTypesSettingsCgm "github.com/tidepool-org/platform/data/types/settings/cgm"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomHighLevelAlertDEPRECATED(units *string) *dataTypesSettingsCgm.HighLevelAlertDEPRECATED {
	datum := dataTypesSettingsCgm.NewHighLevelAlertDEPRECATED()
	datum.Enabled = pointer.FromBool(test.RandomBool())
	datum.Level = pointer.FromFloat64(test.RandomFloat64FromRange(datum.LevelRangeForUnits(units)))
	datum.Snooze = pointer.FromInt(test.RandomIntFromArray(dataTypesSettingsCgm.LevelAlertDEPRECATEDSnoozes()))
	return datum
}

func CloneHighLevelAlertDEPRECATED(datum *dataTypesSettingsCgm.HighLevelAlertDEPRECATED) *dataTypesSettingsCgm.HighLevelAlertDEPRECATED {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsCgm.NewHighLevelAlertDEPRECATED()
	clone.Enabled = pointer.CloneBool(datum.Enabled)
	clone.Level = pointer.CloneFloat64(datum.Level)
	clone.Snooze = pointer.CloneInt(datum.Snooze)
	return clone
}

func RandomLowLevelAlertDEPRECATED(units *string) *dataTypesSettingsCgm.LowLevelAlertDEPRECATED {
	datum := dataTypesSettingsCgm.NewLowLevelAlertDEPRECATED()
	datum.Enabled = pointer.FromBool(test.RandomBool())
	datum.Level = pointer.FromFloat64(test.RandomFloat64FromRange(datum.LevelRangeForUnits(units)))
	datum.Snooze = pointer.FromInt(test.RandomIntFromArray(dataTypesSettingsCgm.LevelAlertDEPRECATEDSnoozes()))
	return datum
}

func CloneLowLevelAlertDEPRECATED(datum *dataTypesSettingsCgm.LowLevelAlertDEPRECATED) *dataTypesSettingsCgm.LowLevelAlertDEPRECATED {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsCgm.NewLowLevelAlertDEPRECATED()
	clone.Enabled = pointer.CloneBool(datum.Enabled)
	clone.Level = pointer.CloneFloat64(datum.Level)
	clone.Snooze = pointer.CloneInt(datum.Snooze)
	return clone
}
