package test

import (
	dataTypesSettingsCgm "github.com/tidepool-org/platform/data/types/settings/cgm"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomScheduledAlerts(minimumLength int, maximumLength int) *dataTypesSettingsCgm.ScheduledAlerts {
	datum := make(dataTypesSettingsCgm.ScheduledAlerts, test.RandomIntFromRange(minimumLength, maximumLength))
	for index := range datum {
		datum[index] = RandomScheduledAlert()
	}
	return &datum
}

func CloneScheduledAlerts(datum *dataTypesSettingsCgm.ScheduledAlerts) *dataTypesSettingsCgm.ScheduledAlerts {
	if datum == nil {
		return nil
	}
	clone := make(dataTypesSettingsCgm.ScheduledAlerts, len(*datum))
	for index, d := range *datum {
		clone[index] = CloneScheduledAlert(d)
	}
	return &clone
}

func NewArrayFromScheduledAlerts(datum *dataTypesSettingsCgm.ScheduledAlerts, objectFormat test.ObjectFormat) []interface{} {
	if datum == nil {
		return nil
	}
	array := []interface{}{}
	for _, d := range *datum {
		array = append(array, NewObjectFromScheduledAlert(d, objectFormat))
	}
	return array
}

func RandomScheduledAlert() *dataTypesSettingsCgm.ScheduledAlert {
	datum := dataTypesSettingsCgm.NewScheduledAlert()
	datum.Name = pointer.FromString(test.RandomStringFromRange(1, dataTypesSettingsCgm.ScheduledAlertNameLengthMaximum))
	datum.Days = pointer.FromStringArray(test.RandomStringArrayFromRangeAndArrayWithoutDuplicates(1, len(dataTypesSettingsCgm.ScheduledAlertDays()), dataTypesSettingsCgm.ScheduledAlertDays()))
	datum.Start = pointer.FromInt(test.RandomIntFromRange(dataTypesSettingsCgm.ScheduledAlertStartMinimum, dataTypesSettingsCgm.ScheduledAlertStartMaximum))
	datum.End = pointer.FromInt(test.RandomIntFromRange(dataTypesSettingsCgm.ScheduledAlertEndMinimum, dataTypesSettingsCgm.ScheduledAlertEndMaximum))
	datum.Alerts = RandomAlerts()
	return datum
}

func CloneScheduledAlert(datum *dataTypesSettingsCgm.ScheduledAlert) *dataTypesSettingsCgm.ScheduledAlert {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsCgm.NewScheduledAlert()
	clone.Name = test.CloneString(datum.Name)
	clone.Days = test.CloneStringArray(datum.Days)
	clone.Start = test.CloneInt(datum.Start)
	clone.End = test.CloneInt(datum.End)
	clone.Alerts = CloneAlerts(datum.Alerts)
	return clone
}

func NewObjectFromScheduledAlert(datum *dataTypesSettingsCgm.ScheduledAlert, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Name != nil {
		object["name"] = test.NewObjectFromString(*datum.Name, objectFormat)
	}
	if datum.Days != nil {
		object["days"] = test.NewObjectFromStringArray(*datum.Days, objectFormat)
	}
	if datum.Start != nil {
		object["start"] = test.NewObjectFromInt(*datum.Start, objectFormat)
	}
	if datum.End != nil {
		object["end"] = test.NewObjectFromInt(*datum.End, objectFormat)
	}
	if datum.Alerts != nil {
		object["alerts"] = NewObjectFromAlerts(datum.Alerts, objectFormat)
	}
	return object
}
