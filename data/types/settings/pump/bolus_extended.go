package pump

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

type BolusExtended struct {
	Enabled *bool `json:"enabled,omitempty" bson:"enabled,omitempty"`
}

func ParseBolusExtended(parser data.ObjectParser) *BolusExtended {
	if parser.Object() == nil {
		return nil
	}
	datum := NewBolusExtended()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewBolusExtended() *BolusExtended {
	return &BolusExtended{}
}

func (b *BolusExtended) Parse(parser data.ObjectParser) {
	b.Enabled = parser.ParseBoolean("enabled")
}

func (b *BolusExtended) Validate(validator structure.Validator) {
	validator.Bool("enabled", b.Enabled).Exists()
}

func (b *BolusExtended) Normalize(normalizer data.Normalizer) {}
