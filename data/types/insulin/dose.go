package insulin

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	TotalMaximum = 100.0
	TotalMinimum = 0.0
	UnitsUnits   = "units"
)

func Units() []string {
	return []string{
		UnitsUnits,
	}
}

type Dose struct {
	Total *float64 `json:"total,omitempty" bson:"total,omitempty"`
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
}

func ParseDose(parser data.ObjectParser) *Dose {
	if parser.Object() == nil {
		return nil
	}
	dose := NewDose()
	dose.Parse(parser)
	parser.ProcessNotParsed()
	return dose
}

func NewDose() *Dose {
	return &Dose{}
}

func (d *Dose) Parse(parser data.ObjectParser) {
	d.Total = parser.ParseFloat("total")
	d.Units = parser.ParseString("units")
}

func (d *Dose) Validate(validator structure.Validator) {
	validator.Float64("total", d.Total).Exists().InRange(TotalMinimum, TotalMaximum)
	validator.String("units", d.Units).Exists().OneOf(Units()...)
}

func (d *Dose) Normalize(normalizer data.Normalizer) {}
