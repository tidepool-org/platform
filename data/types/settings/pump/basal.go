package pump

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

type Basal struct {
	RateMaximum *BasalRateMaximum `json:"rateMaximum,omitempty" bson:"rateMaximum,omitempty"`
	Temporary   *BasalTemporary   `json:"temporary,omitempty" bson:"temporary,omitempty"`
}

func ParseBasal(parser data.ObjectParser) *Basal {
	if parser.Object() == nil {
		return nil
	}
	datum := NewBasal()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewBasal() *Basal {
	return &Basal{}
}

func (b *Basal) Parse(parser data.ObjectParser) {
	b.RateMaximum = ParseBasalRateMaximum(parser.NewChildObjectParser("rateMaximum"))
	b.Temporary = ParseBasalTemporary(parser.NewChildObjectParser("temporary"))
}

func (b *Basal) Validate(validator structure.Validator) {
	if b.RateMaximum != nil {
		b.RateMaximum.Validate(validator.WithReference("rateMaximum"))
	}
	if b.Temporary != nil {
		b.Temporary.Validate(validator.WithReference("temporary"))
	}
}

func (b *Basal) Normalize(normalizer data.Normalizer) {
	if b.RateMaximum != nil {
		b.RateMaximum.Normalize(normalizer.WithReference("rateMaximum"))
	}
	if b.Temporary != nil {
		b.Temporary.Normalize(normalizer.WithReference("temporary"))
	}
}
