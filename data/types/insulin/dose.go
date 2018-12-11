package insulin

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	DoseActiveUnitsMaximum     = 250.0
	DoseActiveUnitsMinimum     = 0.0
	DoseCorrectionUnitsMaximum = 250.0
	DoseCorrectionUnitsMinimum = -250.0
	DoseFoodUnitsMaximum       = 250.0
	DoseFoodUnitsMinimum       = 0.0
	DoseTotalUnitsMaximum      = 250.0
	DoseTotalUnitsMinimum      = 0.0
	DoseUnitsUnits             = "Units"
)

func DoseUnits() []string {
	return []string{
		DoseUnitsUnits,
	}
}

type Dose struct {
	Active     *float64 `json:"active,omitempty" bson:"active,omitempty"`
	Correction *float64 `json:"correction,omitempty" bson:"correction,omitempty"`
	Food       *float64 `json:"food,omitempty" bson:"food,omitempty"`
	Total      *float64 `json:"total,omitempty" bson:"total,omitempty"`
	Units      *string  `json:"units,omitempty" bson:"units,omitempty"`
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
	d.Active = parser.ParseFloat("active")
	d.Correction = parser.ParseFloat("correction")
	d.Food = parser.ParseFloat("food")
	d.Total = parser.ParseFloat("total")
	d.Units = parser.ParseString("units")
}

func (d *Dose) Validate(validator structure.Validator) {
	validator.Float64("active", d.Active).InRange(DoseActiveUnitsMinimum, DoseActiveUnitsMaximum)
	validator.Float64("correction", d.Correction).InRange(DoseCorrectionUnitsMinimum, DoseCorrectionUnitsMaximum)
	validator.Float64("food", d.Food).InRange(DoseFoodUnitsMinimum, DoseFoodUnitsMaximum)
	validator.Float64("total", d.Total).Exists().InRange(DoseTotalUnitsMinimum, DoseTotalUnitsMaximum)
	validator.String("units", d.Units).Exists().OneOf(DoseUnits()...)
}

func (d *Dose) Normalize(normalizer data.Normalizer) {}
