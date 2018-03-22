package pump

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

type BolusCombination struct {
	Enabled *bool `json:"enabled,omitempty" bson:"enabled,omitempty"`
}

func ParseBolusCombination(parser data.ObjectParser) *BolusCombination {
	if parser.Object() == nil {
		return nil
	}
	datum := NewBolusCombination()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewBolusCombination() *BolusCombination {
	return &BolusCombination{}
}

func (b *BolusCombination) Parse(parser data.ObjectParser) {
	b.Enabled = parser.ParseBoolean("enabled")
}

func (b *BolusCombination) Validate(validator structure.Validator) {
	validator.Bool("enabled", b.Enabled).Exists()
}

func (b *BolusCombination) Normalize(normalizer data.Normalizer) {}
