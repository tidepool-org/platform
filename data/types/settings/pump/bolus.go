package pump

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

type Bolus struct {
	AmountMaximum *BolusAmountMaximum `json:"amountMaximum,omitempty" bson:"amountMaximum,omitempty"`
	Calculator    *BolusCalculator    `json:"calculator,omitempty" bson:"calculator,omitempty"`
	Extended      *BolusExtended      `json:"extended,omitempty" bson:"extended,omitempty"`
}

func ParseBolus(parser structure.ObjectParser) *Bolus {
	if !parser.Exists() {
		return nil
	}
	datum := NewBolus()
	parser.Parse(datum)
	return datum
}

func NewBolus() *Bolus {
	return &Bolus{}
}

func (b *Bolus) Parse(parser structure.ObjectParser) {
	b.AmountMaximum = ParseBolusAmountMaximum(parser.WithReferenceObjectParser("amountMaximum"))
	b.Calculator = ParseBolusCalculator(parser.WithReferenceObjectParser("calculator"))
	b.Extended = ParseBolusExtended(parser.WithReferenceObjectParser("extended"))
}

func (b *Bolus) Validate(validator structure.Validator) {
	if b.AmountMaximum != nil {
		b.AmountMaximum.Validate(validator.WithReference("amountMaximum"))
	}
	if b.Calculator != nil {
		b.Calculator.Validate(validator.WithReference("calculator"))
	}
	if b.Extended != nil {
		b.Extended.Validate(validator.WithReference("extended"))
	}
}

func (b *Bolus) Normalize(normalizer data.Normalizer) {
	if b.AmountMaximum != nil {
		b.AmountMaximum.Normalize(normalizer.WithReference("amountMaximum"))
	}
	if b.Calculator != nil {
		b.Calculator.Normalize(normalizer.WithReference("calculator"))
	}
	if b.Extended != nil {
		b.Extended.Normalize(normalizer.WithReference("extended"))
	}
}
