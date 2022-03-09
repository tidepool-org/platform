package food

import (
	"github.com/tidepool-org/platform/structure"
)

const (
	ProteinTotalGramsMaximum = 1000.0
	ProteinTotalGramsMinimum = 0.0
	ProteinUnitsGrams        = "grams"
)

func ProteinUnits() []string {
	return []string{
		ProteinUnitsGrams,
	}
}

type Protein struct {
	Total *float64 `json:"total,omitempty" bson:"total,omitempty"`
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
}

func ParseProtein(parser structure.ObjectParser) *Protein {
	if !parser.Exists() {
		return nil
	}
	datum := NewProtein()
	parser.Parse(datum)
	return datum
}

func NewProtein() *Protein {
	return &Protein{}
}

func (p *Protein) Parse(parser structure.ObjectParser) {
	p.Total = parser.Float64("total")
	p.Units = parser.String("units")
}

func (p *Protein) Validate(validator structure.Validator) {
	validator.Float64("total", p.Total).Exists().InRange(ProteinTotalGramsMinimum, ProteinTotalGramsMaximum)
	validator.String("units", p.Units).Exists().OneOf(ProteinUnits()...)
}
