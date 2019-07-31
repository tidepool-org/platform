package pump

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	DisplayBloodGlucoseUnitsMgPerDL  = "mg/dL"
	DisplayBloodGlucoseUnitsMmolPerL = "mmol/L"
)

func DisplayBloodGlucoseUnits() []string {
	return []string{
		DisplayBloodGlucoseUnitsMgPerDL,
		DisplayBloodGlucoseUnitsMmolPerL,
	}
}

type DisplayBloodGlucose struct {
	Units *string `json:"units,omitempty" bson:"units,omitempty"`
}

func ParseDisplayBloodGlucose(parser structure.ObjectParser) *DisplayBloodGlucose {
	if !parser.Exists() {
		return nil
	}
	datum := NewDisplayBloodGlucose()
	parser.Parse(datum)
	return datum
}

func NewDisplayBloodGlucose() *DisplayBloodGlucose {
	return &DisplayBloodGlucose{}
}

func (d *DisplayBloodGlucose) Parse(parser structure.ObjectParser) {
	d.Units = parser.String("units")
}

func (d *DisplayBloodGlucose) Validate(validator structure.Validator) {
	validator.String("units", d.Units).Exists().OneOf(DisplayBloodGlucoseUnits()...)
}

func (d *DisplayBloodGlucose) Normalize(normalizer data.Normalizer) {}
