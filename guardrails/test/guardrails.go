package test

import "github.com/tidepool-org/devices/api"

func NewBasalRatesGuardRail() *api.BasalRatesGuardRail {
	return &api.BasalRatesGuardRail{AbsoluteBounds: []*api.AbsoluteBounds{
		newAbsoluteBoundsFromNanos(0, 50000000, 30, 0, 50000000),
	}}
}

func NewCorrectionRangeGuardRail() *api.CorrectionRangeGuardRail {
	return &api.CorrectionRangeGuardRail{
		AbsoluteBounds: newAbsoluteBoundsFromUnits(87, 180, 1),
	}
}

func NewPremealCorrectionRangeGuardRail() *api.CorrectionRangeGuardRail {
	return &api.CorrectionRangeGuardRail{
		AbsoluteBounds: newAbsoluteBoundsFromUnits(87, 130, 1),
	}
}

func NewWorkoutCorrectionRangeGuardRail() *api.CorrectionRangeGuardRail {
	return &api.CorrectionRangeGuardRail{
		AbsoluteBounds: newAbsoluteBoundsFromUnits(85, 250, 1),
	}
}

func NewCarbohydrateRatioGuardRail() *api.CarbohydrateRatioGuardRail {
	return &api.CarbohydrateRatioGuardRail{
		AbsoluteBounds: newAbsoluteBoundsFromNanos(2, 0, 150, 0, 10000000),
	}
}

func NewInsulinSensitivityGuardRail() *api.InsulinSensitivityGuardRail {
	return &api.InsulinSensitivityGuardRail{
		AbsoluteBounds: newAbsoluteBoundsFromUnits(10, 500, 1),
	}
}

func NewGlucoseSafetyLimitGuardRail() *api.GlucoseSafetyLimitGuardRail {
	return &api.GlucoseSafetyLimitGuardRail{
		AbsoluteBounds: newAbsoluteBoundsFromUnits(67, 110, 1),
	}
}

func NewBasalRateMaximumGuardRail() *api.BasalRateMaximumGuardRail {
	return &api.BasalRateMaximumGuardRail{
		AbsoluteBounds: newAbsoluteBoundsFromNanos(0, 50000000, 30, 0, 50000000),
	}
}

func NewBolusAmountMaximumGuardRail() *api.BolusAmountMaximumGuardRail {
	return &api.BolusAmountMaximumGuardRail{
		AbsoluteBounds: newAbsoluteBoundsFromNanos(0, 50000000, 30, 0, 50000000),
	}
}

func newAbsoluteBoundsFromNanos(from, fromNanos, to, toNanos int32, incrementNanos int32) *api.AbsoluteBounds {
	return &api.AbsoluteBounds{
		Minimum: &api.FixedDecimal{
			Units: from,
			Nanos: fromNanos,
		},
		Maximum: &api.FixedDecimal{
			Units: to,
			Nanos: toNanos,
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
