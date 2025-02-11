package continuous

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
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

	TrendRateMaximum      = 100
	TrendRateMinimum      = -100
	SampleIntervalMinimum = 0
	SampleIntervalMaximum = 24 * 60 * 60 * 1000
)

func Trends() []string {
	return []string{ConstantRate, SlowFall, SlowRise, ModerateFall, ModerateRise, RapidFall, RapidRise}
}

type Continuous struct {
	glucose.Glucose `bson:",inline"`
	Trend           *string  `json:"trend,omitempty" bson:"trend,omitempty"`
	TrendRate       *float64 `json:"trendRate,omitempty" bson:"trendRate,omitempty"`
	SampleInterval  *int     `json:"sampleInterval,omitempty" bson:"sampleInterval,omitempty"`
	Backfilled      *bool    `json:"backfilled,omitempty" bson:"backfilled,omitempty"`
}

func New() *Continuous {
	return &Continuous{
		Glucose: glucose.New(Type),
	}
}

func (c *Continuous) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(c.Meta())
	}

	c.Glucose.Parse(parser)

	c.Trend = parser.String("trend")
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
	validator.Float64("trendRate", c.TrendRate).InRange(TrendRateMinimum, TrendRateMaximum)
	validator.Int("sampleInterval", c.SampleInterval).InRange(SampleIntervalMinimum, SampleIntervalMaximum)
}

func (c *Continuous) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(c.Meta())
	}

	c.Glucose.Normalize(normalizer)
}
