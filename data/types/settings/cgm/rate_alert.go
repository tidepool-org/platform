package cgm

import (
	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	RateMgdLThree  = 3.0
	RateMgdLTwo    = 2.0
	RateMmolLThree = 0.16652243973136602
	RateMmolLTwo   = 0.11101495982091067
)

type RateAlert struct {
	Enabled *bool    `json:"enabled,omitempty" bson:"enabled,omitempty"`
	Rate    *float64 `json:"rate,omitempty" bson:"rate,omitempty"` // TODO: Make always positive
}

func (r *RateAlert) Parse(parser data.ObjectParser) {
	r.Enabled = parser.ParseBoolean("enabled")
	r.Rate = parser.ParseFloat("rate")
}

func (r *RateAlert) Validate(validator structure.Validator, units *string) {
	validator.Bool("enabled", r.Enabled).Exists()
	validator.Float64("rate", r.Rate).Exists()
}

func (r *RateAlert) Normalize(normalizer data.Normalizer, units *string) {
	if normalizer.Origin() == structure.OriginExternal {
		r.Rate = dataBloodGlucose.NormalizeValueForUnits(r.Rate, units)
	}
}

type FallRateAlert struct {
	RateAlert `bson:",inline"`
}

func ParseFallRateAlert(parser data.ObjectParser) *FallRateAlert {
	if parser.Object() == nil {
		return nil
	}
	fallRateAlert := NewFallRateAlert()
	fallRateAlert.Parse(parser)
	parser.ProcessNotParsed()
	return fallRateAlert
}

func NewFallRateAlert() *FallRateAlert {
	return &FallRateAlert{}
}

func (f *FallRateAlert) Validate(validator structure.Validator, units *string) {
	f.RateAlert.Validate(validator, units)

	if rates := f.RatesForUnits(units); len(rates) > 0 {
		validator.Float64("rate", f.Rate).OneOf(rates...)
	}
}

func (f *FallRateAlert) RatesForUnits(units *string) []float64 {
	if units != nil {
		switch *units {
		case dataBloodGlucose.MgdL, dataBloodGlucose.Mgdl:
			return []float64{-RateMgdLThree, -RateMgdLTwo}
		case dataBloodGlucose.MmolL, dataBloodGlucose.Mmoll:
			return []float64{-RateMmolLThree, -RateMmolLTwo}
		}
	}
	return nil
}

type RiseRateAlert struct {
	RateAlert `bson:",inline"`
}

func ParseRiseRateAlert(parser data.ObjectParser) *RiseRateAlert {
	if parser.Object() == nil {
		return nil
	}
	riseRateAlert := NewRiseRateAlert()
	riseRateAlert.Parse(parser)
	parser.ProcessNotParsed()
	return riseRateAlert
}

func NewRiseRateAlert() *RiseRateAlert {
	return &RiseRateAlert{}
}

func (r *RiseRateAlert) Validate(validator structure.Validator, units *string) {
	r.RateAlert.Validate(validator, units)

	if rates := r.RatesForUnits(units); len(rates) > 0 {
		validator.Float64("rate", r.Rate).OneOf(rates...)
	}
}

func (r *RiseRateAlert) RatesForUnits(units *string) []float64 {
	if units != nil {
		switch *units {
		case dataBloodGlucose.MgdL, dataBloodGlucose.Mgdl:
			return []float64{RateMgdLTwo, RateMgdLThree}
		case dataBloodGlucose.MmolL, dataBloodGlucose.Mmoll:
			return []float64{RateMmolLTwo, RateMmolLThree}
		}
	}
	return nil
}

type RateAlerts struct {
	FallRateAlert *FallRateAlert `json:"fallRate,omitempty" bson:"fallRate,omitempty"`
	RiseRateAlert *RiseRateAlert `json:"riseRate,omitempty" bson:"riseRate,omitempty"`
}

func ParseRateAlerts(parser data.ObjectParser) *RateAlerts {
	if parser.Object() == nil {
		return nil
	}
	rateAlerts := NewRateAlerts()
	rateAlerts.Parse(parser)
	parser.ProcessNotParsed()
	return rateAlerts
}

func NewRateAlerts() *RateAlerts {
	return &RateAlerts{}
}

func (r *RateAlerts) Parse(parser data.ObjectParser) {
	r.FallRateAlert = ParseFallRateAlert(parser.NewChildObjectParser("fallRate"))
	r.RiseRateAlert = ParseRiseRateAlert(parser.NewChildObjectParser("riseRate"))
}

func (r *RateAlerts) Validate(validator structure.Validator, units *string) {
	if r.FallRateAlert != nil {
		r.FallRateAlert.Validate(validator.WithReference("fallRate"), units)
	} else {
		validator.WithReference("fallRate").ReportError(structureValidator.ErrorValueNotExists())
	}
	if r.RiseRateAlert != nil {
		r.RiseRateAlert.Validate(validator.WithReference("riseRate"), units)
	} else {
		validator.WithReference("riseRate").ReportError(structureValidator.ErrorValueNotExists())
	}
}

func (r *RateAlerts) Normalize(normalizer data.Normalizer, units *string) {
	if normalizer.Origin() == structure.OriginExternal {
		if r.FallRateAlert != nil {
			r.FallRateAlert.Normalize(normalizer, units)
		}
		if r.RiseRateAlert != nil {
			r.RiseRateAlert.Normalize(normalizer, units)
		}
	}
}
