package test

import (
	dataTypesSettingsPump "github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomInsulinModel() *dataTypesSettingsPump.InsulinModel {
	datum := dataTypesSettingsPump.NewInsulinModel()
	datum.ModelType = pointer.FromString(test.RandomStringFromArray(dataTypesSettingsPump.InsulinModelModelTypes()))
	if *datum.ModelType == dataTypesSettingsPump.InsulinModelModelTypeOther {
		datum.ModelTypeOther = pointer.FromString(test.RandomStringFromRange(1, dataTypesSettingsPump.InsulinModelModelTypeOtherLengthMaximum))
	}
	datum.ActionDuration = pointer.FromInt(test.RandomIntFromRange(dataTypesSettingsPump.InsulinModelActionDurationMinimum, dataTypesSettingsPump.InsulinModelActionDurationMaximum))
	datum.ActionPeakOffset = pointer.FromInt(test.RandomIntFromRange(dataTypesSettingsPump.InsulinModelActionPeakOffsetMinimum, *datum.ActionDuration))
	return datum
}

func CloneInsulinModel(datum *dataTypesSettingsPump.InsulinModel) *dataTypesSettingsPump.InsulinModel {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsPump.NewInsulinModel()
	clone.ModelType = pointer.CloneString(datum.ModelType)
	clone.ModelTypeOther = pointer.CloneString(datum.ModelTypeOther)
	clone.ActionDuration = pointer.CloneInt(datum.ActionDuration)
	clone.ActionPeakOffset = pointer.CloneInt(datum.ActionPeakOffset)
	return clone
}
