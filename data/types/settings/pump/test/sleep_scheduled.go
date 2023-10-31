package test

import (
	dataTypesSettingsPump "github.com/tidepool-org/platform/data/types/settings/pump"

	dataTypesCommon "github.com/tidepool-org/platform/data/types/common"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomSleepSchedules(minimumLength int, maximumLength int) *dataTypesSettingsPump.SleepSchedules {
	datum := make(dataTypesSettingsPump.SleepSchedules, test.RandomIntFromRange(minimumLength, maximumLength))
	for index := range datum {
		datum[index] = RandomSleepSchedule()
	}
	return &datum
}

func CloneScheduledAlerts(datum *dataTypesSettingsPump.SleepSchedules) *dataTypesSettingsPump.SleepSchedules {
	if datum == nil {
		return nil
	}
	clone := make(dataTypesSettingsPump.SleepSchedules, len(*datum))
	for index, d := range *datum {
		clone[index] = CloneSleepSchedule(d)
	}
	return &clone
}

func NewArrayFromSleepSchedules(datum *dataTypesSettingsPump.SleepSchedules, objectFormat test.ObjectFormat) []interface{} {
	if datum == nil {
		return nil
	}
	array := []interface{}{}
	for _, d := range *datum {
		array = append(array, NewObjectFromSleepSchedule(d, objectFormat))
	}
	return array
}

func RandomSleepSchedule() *dataTypesSettingsPump.SleepSchedule {
	datum := dataTypesSettingsPump.NewSleepSchedule()
	// enabled by default, if not enbaled days, start and end not required
	datum.Enabled = pointer.FromBool(true)
	datum.Days = pointer.FromStringArray(test.RandomStringArrayFromRangeAndArrayWithoutDuplicates(1, len(dataTypesCommon.Days()), dataTypesCommon.Days()))
	datum.Start = pointer.FromInt(test.RandomIntFromRange(dataTypesSettingsPump.SleepSchedulesMidnightOffsetMinimum, dataTypesSettingsPump.SleepSchedulesMidnightOffsetMaximum))
	datum.End = pointer.FromInt(test.RandomIntFromRange(dataTypesSettingsPump.SleepSchedulesMidnightOffsetMinimum, dataTypesSettingsPump.SleepSchedulesMidnightOffsetMaximum))
	return datum
}

func CloneSleepSchedule(datum *dataTypesSettingsPump.SleepSchedule) *dataTypesSettingsPump.SleepSchedule {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsPump.NewSleepSchedule()
	clone.Enabled = pointer.CloneBool(datum.Enabled)
	clone.Days = pointer.CloneStringArray(datum.Days)
	clone.Start = pointer.CloneInt(datum.Start)
	clone.End = pointer.CloneInt(datum.End)
	return clone
}

func NewObjectFromSleepSchedule(datum *dataTypesSettingsPump.SleepSchedule, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Enabled != nil {
		object["enabled"] = test.NewObjectFromBool(*datum.Enabled, objectFormat)
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

	return object
}
