package insulin

import "github.com/tidepool-org/platform/data"

const (
	UnitsUnits = "units"

	TotalMinimum = 0
	TotalMaximum = 100
)

type Dose struct {
	Total *float64 `json:"total,omitempty" bson:"total,omitempty"`
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
}

func NewDose() *Dose {
	return &Dose{}
}

func (d *Dose) Parse(parser data.ObjectParser) {
	d.Total = parser.ParseFloat("total")
	d.Units = parser.ParseString("units")
}

func (d *Dose) Validate(validator data.Validator) {
	validator.ValidateFloat("total", d.Total).Exists().InRange(TotalMinimum, TotalMaximum)
	validator.ValidateString("units", d.Units).Exists().OneOf([]string{UnitsUnits})
}

func (d *Dose) Normalize(normalizer data.Normalizer) {}

func ParseDose(parser data.ObjectParser) *Dose {
	if parser.Object() == nil {
		return nil
	}

	dose := NewDose()
	dose.Parse(parser)
	parser.ProcessNotParsed()

	return dose
}
