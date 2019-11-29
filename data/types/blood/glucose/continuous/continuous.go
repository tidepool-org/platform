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
)

type Continuous struct {
	Trend           *Trend `json:"trend,omitempty" bson:"trend,omitempty"`
	glucose.Glucose `bson:",inline"`
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

	c.Trend = ParseTrend(parser.WithReferenceObjectParser("trend"))
}

func (c *Continuous) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(c.Meta())
	}

	c.Glucose.Validate(validator)

	if c.Type != "" {
		validator.String("type", &c.Type).EqualTo(Type)
	}

}

func (c *Continuous) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(c.Meta())
	}

	c.Glucose.Normalize(normalizer)
}
