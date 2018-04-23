package pump

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

type BolusCalculator struct {
	Enabled *bool `json:"enabled,omitempty" bson:"enabled,omitempty"`
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
}

func (b *BolusCalculator) Validate(validator structure.Validator) {
	validator.Bool("enabled", b.Enabled).Exists()
}

func (b *BolusCalculator) Normalize(normalizer data.Normalizer) {}
