package food

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	ProteinTotalGramsMaximum = 1000
	ProteinTotalGramsMinimum = 0
	ProteinUnitsGrams        = "grams"
)

func ProteinUnits() []string {
	return []string{
		ProteinUnitsGrams,
	}
}

type Protein struct {
	Total *int    `json:"total,omitempty" bson:"total,omitempty"`
	Units *string `json:"units,omitempty" bson:"units,omitempty"`
}

func ParseProtein(parser data.ObjectParser) *Protein {
	if parser.Object() == nil {
		return nil
	}
	datum := NewProtein()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewProtein() *Protein {
	return &Protein{}
}

func (p *Protein) Parse(parser data.ObjectParser) {
	p.Total = parser.ParseInteger("total")
	p.Units = parser.ParseString("units")
}

func (p *Protein) Validate(validator structure.Validator) {
	validator.Int("total", p.Total).Exists().InRange(ProteinTotalGramsMinimum, ProteinTotalGramsMaximum)
	validator.String("units", p.Units).Exists().OneOf(ProteinUnits()...)
}

func (p *Protein) Normalize(normalizer data.Normalizer) {}
