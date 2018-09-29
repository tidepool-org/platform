package test

import (
	dataTypesSettingsCgm "github.com/tidepool-org/platform/data/types/settings/cgm"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomOutOfRangeAlertDEPRECATED() *dataTypesSettingsCgm.OutOfRangeAlertDEPRECATED {
	datum := dataTypesSettingsCgm.NewOutOfRangeAlertDEPRECATED()
	datum.Enabled = pointer.FromBool(test.RandomBool())
	datum.Threshold = pointer.FromInt(test.RandomIntFromArray(dataTypesSettingsCgm.OutOfRangeAlertDEPRECATEDThresholds()))
	return datum
}

func CloneOutOfRangeAlertDEPRECATED(datum *dataTypesSettingsCgm.OutOfRangeAlertDEPRECATED) *dataTypesSettingsCgm.OutOfRangeAlertDEPRECATED {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsCgm.NewOutOfRangeAlertDEPRECATED()
	clone.Enabled = test.CloneBool(datum.Enabled)
	clone.Threshold = test.CloneInt(datum.Threshold)
	return clone
}
