package cgm

import (
	"math"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/structure"
)

const (
	FallAlertRateMgdLMinuteMaximum  = 10.0
	FallAlertRateMgdLMinuteMinimum  = 0.0
	FallAlertRateMmolLMinuteMaximum = 0.55507
	FallAlertRateMmolLMinuteMinimum = 0.05551
	RiseAlertRateMgdLMinuteMaximum  = 10.0
	RiseAlertRateMgdLMinuteMinimum  = 0.0
	RiseAlertRateMmolLMinuteMaximum = 0.55507
	RiseAlertRateMmolLMinuteMinimum = 0.05551
)

type RateAlert struct {
	Alert `bson:",inline"`
	Rate  *float64 `json:"rate,omitempty" bson:"rate,omitempty"`
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
}

func (r *RateAlert) Parse(parser structure.ObjectParser) {
	r.Alert.Parse(parser)
	r.Rate = parser.Float64("rate")
	r.Units = parser.String("units")
}

func (r *RateAlert) Validate(validator structure.Validator) {
	r.Alert.Validate(validator)
	if unitsValidator := validator.String("units", r.Units); r.Rate != nil {
		unitsValidator.Exists().OneOf(dataBloodGlucose.RateUnits()...)
	} else {
		unitsValidator.NotExists()
	}
}

type FallAlert struct {
	RateAlert `bson:",inline"`
}

func ParseFallAlert(parser structure.ObjectParser) *FallAlert {
	if !parser.Exists() {
		return nil
	}
	datum := NewFallAlert()
	parser.Parse(datum)
	return datum
}

func NewFallAlert() *FallAlert {
	return &FallAlert{}
}

func (f *FallAlert) Validate(validator structure.Validator) {
	f.RateAlert.Validate(validator)
	validator.Float64("rate", f.Rate).InRange(FallAlertRateRangeForUnits(f.Units))
}

func FallAlertRateRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case dataBloodGlucose.MgdLMinute:
			return FallAlertRateMgdLMinuteMinimum, FallAlertRateMgdLMinuteMaximum
		case dataBloodGlucose.MmolLMinute:
			return FallAlertRateMmolLMinuteMinimum, FallAlertRateMmolLMinuteMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}

type RiseAlert struct {
	RateAlert `bson:",inline"`
}

func ParseRiseAlert(parser structure.ObjectParser) *RiseAlert {
	if !parser.Exists() {
		return nil
	}
	datum := NewRiseAlert()
	parser.Parse(datum)
	return datum
}

func NewRiseAlert() *RiseAlert {
	return &RiseAlert{}
}

func (r *RiseAlert) Validate(validator structure.Validator) {
	r.RateAlert.Validate(validator)
	validator.Float64("rate", r.Rate).InRange(RiseAlertRateRangeForUnits(r.Units))
}

func RiseAlertRateRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case dataBloodGlucose.MgdLMinute:
			return RiseAlertRateMgdLMinuteMinimum, RiseAlertRateMgdLMinuteMaximum
		case dataBloodGlucose.MmolLMinute:
			return RiseAlertRateMmolLMinuteMinimum, RiseAlertRateMmolLMinuteMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}
