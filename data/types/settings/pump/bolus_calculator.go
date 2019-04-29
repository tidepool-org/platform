package pump

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

type BolusCalculator struct {
	Enabled *bool                   `json:"enabled,omitempty" bson:"enabled,omitempty"`
	Insulin *BolusCalculatorInsulin `json:"insulin,omitempty" bson:"insulin,omitempty"`
}

func ParseBolusCalculator(parser data.ObjectParser) *BolusCalculator {
	if parser.Object() == nil {
		return nil
	}
	datum := NewBolusCalculator()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewBolusCalculator() *BolusCalculator {
	return &BolusCalculator{}
}

func (b *BolusCalculator) Parse(parser data.ObjectParser) {
	b.Enabled = parser.ParseBoolean("enabled")
	b.Insulin = ParseBolusCalculatorInsulin(parser.NewChildObjectParser("insulin"))
}

func (b *BolusCalculator) Validate(validator structure.Validator) {
	validator.Bool("enabled", b.Enabled).Exists()
	if b.Insulin != nil {
		b.Insulin.Validate(validator.WithReference("insulin"))
	}
}

func (b *BolusCalculator) Normalize(normalizer data.Normalizer) {
	if b.Insulin != nil {
		b.Insulin.Normalize(normalizer.WithReference("insulin"))
	}
}
