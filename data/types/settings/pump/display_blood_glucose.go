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

func ParseDisplayBloodGlucose(parser data.ObjectParser) *DisplayBloodGlucose {
	if parser.Object() == nil {
		return nil
	}
	datum := NewDisplayBloodGlucose()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewDisplayBloodGlucose() *DisplayBloodGlucose {
	return &DisplayBloodGlucose{}
}

func (d *DisplayBloodGlucose) Parse(parser data.ObjectParser) {
	d.Units = parser.ParseString("units")
}

func (d *DisplayBloodGlucose) Validate(validator structure.Validator) {
	validator.String("units", d.Units).Exists().OneOf(DisplayBloodGlucoseUnits()...)
}

func (d *DisplayBloodGlucose) Normalize(normalizer data.Normalizer) {}
