package test

import (
	"fmt"

	dataTypesCommon "github.com/tidepool-org/platform/data/types/common"
	dataTypesSettingsPump "github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func SleepScheduleName(index int) string {
	return fmt.Sprintf("schedule-%d", index)
}

func RandomSleepScheduleMap(count int) *dataTypesSettingsPump.SleepScheduleMap {
	datum := dataTypesSettingsPump.NewSleepScheduleMap()
	for i := 0; i < count; i++ {
		(*datum)[SleepScheduleName(i)] = RandomSleepSchedule()
	}
	return datum
}

func CloneSleepScheduleMap(datum *dataTypesSettingsPump.SleepScheduleMap) *dataTypesSettingsPump.SleepScheduleMap {
	if datum == nil {
		return nil
	}
	clone := make(dataTypesSettingsPump.SleepScheduleMap, len(*datum))
	for k, v := range *datum {
		clone[k] = CloneSleepSchedule(v)
	}
	return &clone
}

func NewObjectFromSleepScheduleMap(datum *dataTypesSettingsPump.SleepScheduleMap, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	for k, v := range *datum {
		object[k] = NewObjectFromSleepSchedule(v, objectFormat)
	}
	return object
}

func RandomSleepSchedule() *dataTypesSettingsPump.SleepSchedule {
	datum := dataTypesSettingsPump.NewSleepSchedule()
	// enabled by default, if not enabled days, start and end not required
	datum.Enabled = pointer.FromBool(true)
	datum.Days = pointer.FromStringArray(test.RandomStringArrayFromRangeAndArrayWithoutDuplicates(1, len(dataTypesCommon.DaysOfWeek()), dataTypesCommon.DaysOfWeek()))
	datum.Start = pointer.FromInt(test.RandomIntFromRange(dataTypesSettingsPump.SleepSchedulesMidnightOffsetMinimum, dataTypesSettingsPump.SleepSchedulesMidnightOffsetMaximum))
	datum.End = pointer.FromInt(test.RandomIntFromRange(*datum.Start, dataTypesSettingsPump.SleepSchedulesMidnightOffsetMaximum))
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
