package continuous

import (
	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataTypesBloodGlucose "github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/structure"
)

const (
	Type = "cbg"

	ConstantRate = "constant"
	SlowFall     = "slowFall"
	SlowRise     = "slowRise"
	ModerateFall = "moderateFall"
	ModerateRise = "moderateRise"
	RapidFall    = "rapidFall"
	RapidRise    = "rapidRise"

	SampleIntervalMinimum = 0
	SampleIntervalMaximum = 24 * 60 * 60 * 1000
)

func Trends() []string {
	return []string{
		ConstantRate,
		SlowFall,
		SlowRise,
		ModerateFall,
		ModerateRise,
		RapidFall,
		RapidRise,
	}
}

type Continuous struct {
	dataTypesBloodGlucose.Glucose `bson:",inline"`
	Trend                         *string  `json:"trend,omitempty" bson:"trend,omitempty"`
	TrendRateUnits                *string  `json:"trendRateUnits,omitempty" bson:"trendRateUnits,omitempty"`
	TrendRate                     *float64 `json:"trendRate,omitempty" bson:"trendRate,omitempty"`
	SampleInterval                *int     `json:"sampleInterval,omitempty" bson:"sampleInterval,omitempty"`
	Backfilled                    *bool    `json:"backfilled,omitempty" bson:"backfilled,omitempty"`
}

func New() *Continuous {
	return &Continuous{
		Glucose: dataTypesBloodGlucose.New(Type),
	}
}

func (c *Continuous) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(c.Meta())
	}

	c.Glucose.Parse(parser)

	c.Trend = parser.String("trend")
	c.TrendRateUnits = parser.String("trendRateUnits")
	c.TrendRate = parser.Float64("trendRate")
	c.SampleInterval = parser.Int("sampleInterval")
	c.Backfilled = parser.Bool("backfilled")
}

func (c *Continuous) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(c.Meta())
	}

	c.Glucose.Validate(validator)

	if c.Type != "" {
		validator.String("type", &c.Type).EqualTo(Type)
	}

	validator.String("trend", c.Trend).OneOf(Trends()...)
	if trendRateUnitsValidator := validator.String("trendRateUnits", c.TrendRateUnits); c.TrendRate != nil {
		trendRateUnitsValidator.Exists().OneOf(dataBloodGlucose.RateUnits()...)
	} else {
		trendRateUnitsValidator.NotExists()
	}
	validator.Float64("trendRate", c.TrendRate).InRange(dataBloodGlucose.ValueRangeForRateUnits(c.TrendRateUnits))
	validator.Int("sampleInterval", c.SampleInterval).InRange(SampleIntervalMinimum, SampleIntervalMaximum)
}

func (c *Continuous) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(c.Meta())
	}

	c.Glucose.Normalize(normalizer)

	if normalizer.Origin() == structure.OriginExternal {
		rateUnits := c.TrendRateUnits
		c.TrendRateUnits = dataBloodGlucose.NormalizeRateUnits(rateUnits)
		c.TrendRate = dataBloodGlucose.NormalizeValueForRateUnits(c.TrendRate, rateUnits)
	}
}
