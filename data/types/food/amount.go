package food

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	UnitsLengthMaximum = 100
	ValueMinimum       = 0.0
)

type Amount struct {
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

func ParseAmount(parser data.ObjectParser) *Amount {
	if parser.Object() == nil {
		return nil
	}
	datum := NewAmount()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewAmount() *Amount {
	return &Amount{}
}

func (a *Amount) Parse(parser data.ObjectParser) {
	a.Units = parser.ParseString("units")
	a.Value = parser.ParseFloat("value")
}

func (a *Amount) Validate(validator structure.Validator) {
	validator.String("units", a.Units).Exists().NotEmpty().LengthLessThanOrEqualTo(UnitsLengthMaximum)
	validator.Float64("value", a.Value).Exists().GreaterThanOrEqualTo(ValueMinimum)
}

func (a *Amount) Normalize(normalizer data.Normalizer) {}
