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

func ParseBolus(parser data.ObjectParser) *Bolus {
	if parser.Object() == nil {
		return nil
	}
	datum := NewBolus()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewBolus() *Bolus {
	return &Bolus{}
}

func (b *Bolus) Parse(parser data.ObjectParser) {
	b.AmountMaximum = ParseBolusAmountMaximum(parser.NewChildObjectParser("amountMaximum"))
	b.Calculator = ParseBolusCalculator(parser.NewChildObjectParser("calculator"))
	b.Extended = ParseBolusExtended(parser.NewChildObjectParser("extended"))
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
