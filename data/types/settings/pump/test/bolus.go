package test

import (
	"fmt"

	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/test"
)

func NewRandomBolus() *pump.Bolus {
	datum := pump.NewBolus()
	datum.AmountMaximum = NewBolusAmountMaximum()
	datum.Extended = NewBolusExtended()
	return datum
}

func CloneBolus(datum *pump.Bolus) *pump.Bolus {
	if datum == nil {
		return nil
	}
	clone := pump.NewBolus()
	clone.AmountMaximum = CloneBolusAmountMaximum(datum.AmountMaximum)
	clone.Extended = CloneBolusExtended(datum.Extended)
	return clone
}

func BolusName(index int) string {
	return fmt.Sprintf("bolus-%d", index)
}

func NewRandomBolusMap(minimumLength int, maximumLength int) *pump.BolusMap {
	datum := pump.NewBolusMap()
	for count := test.RandomIntFromRange(minimumLength, maximumLength); count > 0; count-- {
		datum.Set(BolusName(count), NewRandomBolus())
	}
	return datum
}

func CloneBolusMap(datum *pump.BolusMap) *pump.BolusMap {
	if datum == nil {
		return nil
	}
	clone := pump.NewBolusMap()
	for k, v := range *datum {
		(*clone)[k] = CloneBolus(v)
	}
	return clone
}
