package continuous

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

func Trends() []string {
	return []string{ConstantRate, SlowFall, SlowRise, ModerateFall, ModerateRise, RapidFall, RapidRise}
}

type Trend struct {
	Category *string `json:"category,omitempty" bson:"category,omitempty"`
	Rate     *Rate   `json:"rate,omitempty" bson:"rate,omitempty"`
}

func NewTrend() *Trend {
	return &Trend{}
}

func (c *Trend) Parse(parser structure.ObjectParser) {

	c.Category = parser.String("category")
	c.Rate = ParseRate(parser.WithReferenceObjectParser("rate"))
}

func (c *Trend) Validate(validator structure.Validator) {

	if c.Category != nil {
		validator.String("units", c.Category).Exists().OneOf(Trends()...)
	}

}

func (c *Trend) Normalize(normalizer data.Normalizer) {
}

func ParseTrend(parser structure.ObjectParser) *Trend {
	if !parser.Exists() {
		return nil
	}
	datum := NewTrend()
	parser.Parse(datum)
	return datum
}
