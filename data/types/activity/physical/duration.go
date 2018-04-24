package physical

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	UnitsHours   = "hours"
	UnitsMinutes = "minutes"
	UnitsSeconds = "seconds"
	ValueMinimum = 0
)

func Units() []string {
	return []string{
		UnitsHours,
		UnitsMinutes,
		UnitsSeconds,
	}
}

type Duration struct {
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

func ParseDuration(parser data.ObjectParser) *Duration {
	if parser.Object() == nil {
		return nil
	}
	duration := NewDuration()
	duration.Parse(parser)
	parser.ProcessNotParsed()
	return duration
}

func NewDuration() *Duration {
	return &Duration{}
}

func (d *Duration) Parse(parser data.ObjectParser) {
	d.Units = parser.ParseString("units")
	d.Value = parser.ParseFloat("value")
}

func (d *Duration) Validate(validator structure.Validator) {
	validator.String("units", d.Units).Exists().OneOf(Units()...)
	validator.Float64("value", d.Value).Exists().GreaterThan(ValueMinimum)
}

func (d *Duration) Normalize(normalizer data.Normalizer) {}
