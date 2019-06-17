package test

import (
	dataTypesSettingsCgm "github.com/tidepool-org/platform/data/types/settings/cgm"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomAlerts() *dataTypesSettingsCgm.Alerts {
	datum := dataTypesSettingsCgm.NewAlerts()
	datum.Enabled = pointer.FromBool(test.RandomBool())
	datum.UrgentLow = RandomUrgentLowAlert()
	datum.UrgentLowPredicted = RandomUrgentLowAlert()
	datum.Low = RandomLowAlert()
	datum.LowPredicted = RandomLowAlert()
	datum.High = RandomHighAlert()
	datum.HighPredicted = RandomHighAlert()
	datum.Fall = RandomFallAlert()
	datum.Rise = RandomRiseAlert()
	datum.NoData = RandomNoDataAlert()
	datum.OutOfRange = RandomOutOfRangeAlert()
	return datum
}

func CloneAlerts(datum *dataTypesSettingsCgm.Alerts) *dataTypesSettingsCgm.Alerts {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsCgm.NewAlerts()
	clone.Enabled = pointer.CloneBool(datum.Enabled)
	clone.UrgentLow = CloneUrgentLowAlert(datum.UrgentLow)
	clone.UrgentLowPredicted = CloneUrgentLowAlert(datum.UrgentLowPredicted)
	clone.Low = CloneLowAlert(datum.Low)
	clone.LowPredicted = CloneLowAlert(datum.LowPredicted)
	clone.High = CloneHighAlert(datum.High)
	clone.HighPredicted = CloneHighAlert(datum.HighPredicted)
	clone.Fall = CloneFallAlert(datum.Fall)
	clone.Rise = CloneRiseAlert(datum.Rise)
	clone.NoData = CloneNoDataAlert(datum.NoData)
	clone.OutOfRange = CloneOutOfRangeAlert(datum.OutOfRange)
	return clone
}

func NewObjectFromAlerts(datum *dataTypesSettingsCgm.Alerts, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Enabled != nil {
		object["enabled"] = test.NewObjectFromBool(*datum.Enabled, objectFormat)
	}
	if datum.UrgentLow != nil {
		object["urgentLow"] = NewObjectFromUrgentLowAlert(datum.UrgentLow, objectFormat)
	}
	if datum.UrgentLowPredicted != nil {
		object["urgentLowPredicted"] = NewObjectFromUrgentLowAlert(datum.UrgentLowPredicted, objectFormat)
	}
	if datum.Low != nil {
		object["low"] = NewObjectFromLowAlert(datum.Low, objectFormat)
	}
	if datum.LowPredicted != nil {
		object["lowPredicted"] = NewObjectFromLowAlert(datum.LowPredicted, objectFormat)
	}
	if datum.High != nil {
		object["high"] = NewObjectFromHighAlert(datum.High, objectFormat)
	}
	if datum.HighPredicted != nil {
		object["highPredicted"] = NewObjectFromHighAlert(datum.HighPredicted, objectFormat)
	}
	if datum.Fall != nil {
		object["fall"] = NewObjectFromFallAlert(datum.Fall, objectFormat)
	}
	if datum.Rise != nil {
		object["rise"] = NewObjectFromRiseAlert(datum.Rise, objectFormat)
	}
	if datum.NoData != nil {
		object["noData"] = NewObjectFromNoDataAlert(datum.NoData, objectFormat)
	}
	if datum.OutOfRange != nil {
		object["outOfRange"] = NewObjectFromOutOfRangeAlert(datum.OutOfRange, objectFormat)
	}
	return object
}

func RandomAlert() *dataTypesSettingsCgm.Alert {
	datum := &dataTypesSettingsCgm.Alert{}
	datum.Enabled = pointer.FromBool(test.RandomBool())
	datum.Snooze = RandomSnooze()
	return datum
}

func CloneAlert(datum *dataTypesSettingsCgm.Alert) *dataTypesSettingsCgm.Alert {
	if datum == nil {
		return nil
	}
	clone := &dataTypesSettingsCgm.Alert{}
	clone.Enabled = pointer.CloneBool(datum.Enabled)
	clone.Snooze = CloneSnooze(datum.Snooze)
	return clone
}

func NewObjectFromAlert(datum *dataTypesSettingsCgm.Alert, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Enabled != nil {
		object["enabled"] = test.NewObjectFromBool(*datum.Enabled, objectFormat)
	}
	if datum.Snooze != nil {
		object["snooze"] = NewObjectFromSnooze(datum.Snooze, objectFormat)
	}
	return object
}
