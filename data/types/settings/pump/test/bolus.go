package test

import "github.com/tidepool-org/platform/data/types/settings/pump"

func NewBolus() *pump.Bolus {
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
