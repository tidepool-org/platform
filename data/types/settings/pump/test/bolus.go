package test

import (
	"fmt"

	dataTypesSettingsPump "github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/test"
)

func NewRandomBolus() *dataTypesSettingsPump.Bolus {
	datum := dataTypesSettingsPump.NewBolus()
	datum.AmountMaximum = NewBolusAmountMaximum()
	datum.Extended = NewBolusExtended()
	datum.Calculator = NewBolusCalculator()
	return datum
}

func CloneBolus(datum *dataTypesSettingsPump.Bolus) *dataTypesSettingsPump.Bolus {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsPump.NewBolus()
	clone.AmountMaximum = CloneBolusAmountMaximum(datum.AmountMaximum)
	clone.Extended = CloneBolusExtended(datum.Extended)
	clone.Calculator = CloneBolusCalculator(datum.Calculator)
	return clone
}

func BolusName(index int) string {
	return fmt.Sprintf("bolus-%d", index)
}

func NewRandomBolusMap(minimumLength int, maximumLength int) *dataTypesSettingsPump.BolusMap {
	datum := dataTypesSettingsPump.NewBolusMap()
	count := test.RandomIntFromRange(minimumLength, maximumLength)
	if count == 0 {
		return datum
	}

	for i := 0; i < count; i++ {
		datum.Set(BolusName(count), NewRandomBolus())
	}

	return datum
}

func CloneBolusMap(datum *dataTypesSettingsPump.BolusMap) *dataTypesSettingsPump.BolusMap {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsPump.NewBolusMap()
	for k, v := range *datum {
		(*clone)[k] = CloneBolus(v)
	}
	return clone
}
