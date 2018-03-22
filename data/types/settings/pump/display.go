package pump

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	DisplayUnitsMgPerDL  = "mg/dL"
	DisplayUnitsMmolPerL = "mmol/L"
)

func DisplayUnits() []string {
	return []string{
		DisplayUnitsMgPerDL,
		DisplayUnitsMmolPerL,
	}
}

type Display struct {
	Units *string `json:"units,omitempty" bson:"units,omitempty"`
}

func ParseDisplay(parser data.ObjectParser) *Display {
	if parser.Object() == nil {
		return nil
	}
	datum := NewDisplay()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewDisplay() *Display {
	return &Display{}
}

func (d *Display) Parse(parser data.ObjectParser) {
	d.Units = parser.ParseString("units")
}

func (d *Display) Validate(validator structure.Validator) {
	validator.String("units", d.Units).Exists().OneOf(DisplayUnits()...)
}

func (d *Display) Normalize(normalizer data.Normalizer) {}
