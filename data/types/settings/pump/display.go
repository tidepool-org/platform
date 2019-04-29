package pump

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

type Display struct {
	BloodGlucose *DisplayBloodGlucose `json:"bloodGlucose,omitempty" bson:"bloodGlucose,omitempty"`
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
	d.BloodGlucose = ParseDisplayBloodGlucose(parser.NewChildObjectParser("bloodGlucose"))
}

func (d *Display) Validate(validator structure.Validator) {
	if d.BloodGlucose != nil {
		d.BloodGlucose.Validate(validator.WithReference("bloodGlucose"))
	}
}

func (d *Display) Normalize(normalizer data.Normalizer) {
	if d.BloodGlucose != nil {
		d.BloodGlucose.Normalize(normalizer.WithReference("bloodGlucose"))
	}
}
