package pump

import (
	"sort"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
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

type BolusMap map[string]*Bolus

func ParseBolusMap(parser structure.ObjectParser) *BolusMap {
	if !parser.Exists() {
		return nil
	}
	datum := NewBolusMap()
	parser.Parse(datum)
	return datum
}

func NewBolusMap() *BolusMap {
	return &BolusMap{}
}

func (b *BolusMap) Parse(parser structure.ObjectParser) {
	for _, reference := range parser.References() {
		b.Set(reference, ParseBolus(parser.WithReferenceObjectParser(reference)))
	}
}

func (b *BolusMap) Normalize(normalizer data.Normalizer) {
	for _, name := range b.sortedNames() {
		if datum := b.Get(name); datum != nil {
			datum.Normalize(normalizer.WithReference(name))
		}
	}
}

func (b *BolusMap) Validate(validator structure.Validator) {
	for _, name := range b.sortedNames() {
		datumValidator := validator.WithReference(name)
		if datum := b.Get(name); datum != nil {
			datum.Validate(datumValidator)
		} else {
			datumValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (b *BolusMap) Get(name string) *Bolus {
	if datumArray, exists := (*b)[name]; exists {
		return datumArray
	}
	return nil
}

func (b *BolusMap) Set(name string, datum *Bolus) {
	(*b)[name] = datum
}

func (b *BolusMap) sortedNames() []string {
	names := []string{}
	for name := range *b {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
