package test

import "github.com/tidepool-org/devices/api"

func NewBasalRateGuardRail() *api.BasalRatesGuardRail {
	return &api.BasalRatesGuardRail{AbsoluteBounds: []*api.AbsoluteBounds{
		newAbsoluteBoundsFromNanos(0, 30, 50000000),
	}}
}

func NewCorrectionRangeGuardRail() *api.CorrectionRangeGuardRail {
	return &api.CorrectionRangeGuardRail{
		AbsoluteBounds: newAbsoluteBoundsFromUnits(60, 180, 1),
	}
}

func NewCarbohydrateRatioGuardRail() *api.CarbohydrateRatioGuardRail {
	return &api.CarbohydrateRatioGuardRail{
		AbsoluteBounds: newAbsoluteBoundsFromNanos(1, 150, 10000000),
	}
}

func NewInsulinSensitivityGuardRail() *api.InsulinSensitivityGuardRail {
	return &api.InsulinSensitivityGuardRail{
		AbsoluteBounds: newAbsoluteBoundsFromUnits(10, 500, 1),
	}
}

func NewSuspendThresholdGuardRail() *api.SuspendThresholdGuardRail {
	return &api.SuspendThresholdGuardRail{
		AbsoluteBounds: newAbsoluteBoundsFromUnits(60, 180, 1),
	}
}

func NewBasalRateMaximumGuardRail() *api.BasalRateMaximumGuardRail {
	return &api.BasalRateMaximumGuardRail{
		AbsoluteBounds: newAbsoluteBoundsFromNanos(0, 30, 50000000),
	}
}

func NewBolusAmountMaximumGuardRail() *api.BolusAmountMaximumGuardRail {
	return &api.BolusAmountMaximumGuardRail{
		AbsoluteBounds: newAbsoluteBoundsFromNanos(0, 30, 50000000),
	}
}

func newAbsoluteBoundsFromNanos(from, to int32, incrementNanos int32) *api.AbsoluteBounds {
	return &api.AbsoluteBounds{
		Minimum: &api.FixedDecimal{
			Units: from,
		},
		Maximum: &api.FixedDecimal{
			Units: to,
		},
		Increment: &api.FixedDecimal{
			Nanos: incrementNanos,
		},
	}
}

func newAbsoluteBoundsFromUnits(from, to int32, units int32) *api.AbsoluteBounds {
	return &api.AbsoluteBounds{
		Minimum: &api.FixedDecimal{
			Units: from,
		},
		Maximum: &api.FixedDecimal{
			Units: to,
		},
		Increment: &api.FixedDecimal{
			Units: units,
		},
	}
}
