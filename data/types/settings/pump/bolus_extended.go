package pump

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

type BolusExtended struct {
	Enabled *bool `json:"enabled,omitempty" bson:"enabled,omitempty"`
}

func ParseBolusExtended(parser structure.ObjectParser) *BolusExtended {
	if !parser.Exists() {
		return nil
	}
	datum := NewBolusExtended()
	parser.Parse(datum)
	return datum
}

func NewBolusExtended() *BolusExtended {
	return &BolusExtended{}
}

func (b *BolusExtended) Parse(parser structure.ObjectParser) {
	b.Enabled = parser.Bool("enabled")
}

func (b *BolusExtended) Validate(validator structure.Validator) {
	validator.Bool("enabled", b.Enabled).Exists()
}

func (b *BolusExtended) Normalize(normalizer data.Normalizer) {}
