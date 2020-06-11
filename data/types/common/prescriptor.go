package common

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	ManualPrescriptor = "manual"
	AutoPrescriptor   = "auto"
	HybridPrescriptor = "hybrid"
)

func Presciptors() []string {
	return []string{
		AutoPrescriptor,
		ManualPrescriptor,
		HybridPrescriptor,
	}
}

type Prescriptor struct {
	Prescriptor *string `json:"prescriptor,omitempty" bson:"prescriptor,omitempty"`
}

func ParsePrescriptor(parser structure.ObjectParser) *Prescriptor {
	if !parser.Exists() {
		return nil
	}
	datum := NewPrescriptor()
	parser.Parse(datum)
	return datum
}

func NewPrescriptor() *Prescriptor {
	return &Prescriptor{}
}

func (p *Prescriptor) Parse(parser structure.ObjectParser) {
	p.Prescriptor = parser.String("prescriptor")
}

func (p *Prescriptor) Validate(validator structure.Validator) {
	validator.String("prescriptor", p.Prescriptor).OneOf(Presciptors()...)
}

func (p *Prescriptor) Normalize(normalizer data.Normalizer) {
}
