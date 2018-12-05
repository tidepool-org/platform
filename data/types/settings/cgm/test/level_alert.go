package test

import (
	dataTypesSettingsCgm "github.com/tidepool-org/platform/data/types/settings/cgm"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomLevelAlert() *dataTypesSettingsCgm.LevelAlert {
	datum := &dataTypesSettingsCgm.LevelAlert{}
	datum.Alert = *RandomAlert()
	datum.Level = pointer.FromFloat64(test.RandomFloat64())
	datum.Units = pointer.FromString(test.RandomStringFromArray(dataTypesSettingsCgm.LevelAlertUnits()))
	return datum
}

func CloneLevelAlert(datum *dataTypesSettingsCgm.LevelAlert) *dataTypesSettingsCgm.LevelAlert {
	if datum == nil {
		return nil
	}
	clone := &dataTypesSettingsCgm.LevelAlert{}
	clone.Alert = *CloneAlert(&datum.Alert)
	clone.Level = test.CloneFloat64(datum.Level)
	clone.Units = test.CloneString(datum.Units)
	return clone
}

func NewObjectFromLevelAlert(datum *dataTypesSettingsCgm.LevelAlert, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := NewObjectFromAlert(&datum.Alert, objectFormat)
	if datum.Level != nil {
		object["level"] = test.NewObjectFromFloat64(*datum.Level, objectFormat)
	}
	if datum.Units != nil {
		object["units"] = test.NewObjectFromString(*datum.Units, objectFormat)
	}
	return object
}

func RandomHighAlert() *dataTypesSettingsCgm.HighAlert {
	datum := dataTypesSettingsCgm.NewHighAlert()
	datum.LevelAlert = *RandomLevelAlert()
	datum.Level = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesSettingsCgm.HighAlertLevelRangeForUnits(datum.Units)))
	return datum
}

func CloneHighAlert(datum *dataTypesSettingsCgm.HighAlert) *dataTypesSettingsCgm.HighAlert {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsCgm.NewHighAlert()
	clone.LevelAlert = *CloneLevelAlert(&datum.LevelAlert)
	return clone
}

func NewObjectFromHighAlert(datum *dataTypesSettingsCgm.HighAlert, objectFormat test.ObjectFormat) map[string]interface{} {
	return NewObjectFromLevelAlert(&datum.LevelAlert, objectFormat)
}

func RandomLowAlert() *dataTypesSettingsCgm.LowAlert {
	datum := dataTypesSettingsCgm.NewLowAlert()
	datum.LevelAlert = *RandomLevelAlert()
	datum.Level = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesSettingsCgm.LowAlertLevelRangeForUnits(datum.Units)))
	return datum
}

func CloneLowAlert(datum *dataTypesSettingsCgm.LowAlert) *dataTypesSettingsCgm.LowAlert {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsCgm.NewLowAlert()
	clone.LevelAlert = *CloneLevelAlert(&datum.LevelAlert)
	return clone
}

func NewObjectFromLowAlert(datum *dataTypesSettingsCgm.LowAlert, objectFormat test.ObjectFormat) map[string]interface{} {
	return NewObjectFromLevelAlert(&datum.LevelAlert, objectFormat)
}

func RandomUrgentLowAlert() *dataTypesSettingsCgm.UrgentLowAlert {
	datum := dataTypesSettingsCgm.NewUrgentLowAlert()
	datum.LevelAlert = *RandomLevelAlert()
	datum.Level = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesSettingsCgm.UrgentLowAlertLevelRangeForUnits(datum.Units)))
	return datum
}

func CloneUrgentLowAlert(datum *dataTypesSettingsCgm.UrgentLowAlert) *dataTypesSettingsCgm.UrgentLowAlert {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsCgm.NewUrgentLowAlert()
	clone.LevelAlert = *CloneLevelAlert(&datum.LevelAlert)
	return clone
}

func NewObjectFromUrgentLowAlert(datum *dataTypesSettingsCgm.UrgentLowAlert, objectFormat test.ObjectFormat) map[string]interface{} {
	return NewObjectFromLevelAlert(&datum.LevelAlert, objectFormat)
}
