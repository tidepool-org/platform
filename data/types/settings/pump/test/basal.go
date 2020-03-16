package test

import "github.com/tidepool-org/platform/data/types/settings/pump"

func NewBasal() *pump.Basal {
	datum := pump.NewBasal()
	datum.RateMaximum = NewBasalRateMaximum()
	datum.Temporary = NewBasalTemporary()
	return datum
}

func CloneBasal(datum *pump.Basal) *pump.Basal {
	if datum == nil {
		return nil
	}
	clone := pump.NewBasal()
	clone.RateMaximum = CloneBasalRateMaximum(datum.RateMaximum)
	clone.Temporary = CloneBasalTemporary(datum.Temporary)
	return clone
}
