package pump

import (
	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/structure"
)

type Units struct {
	BloodGlucose *string `json:"bg,omitempty" bson:"bg,omitempty"` // TODO: Rename "bloodGlucose"
}

func ParseUnits(parser structure.ObjectParser) *Units {
	if !parser.Exists() {
		return nil
	}
	datum := NewUnits()
	parser.Parse(datum)
	return datum
}

func NewUnits() *Units {
	return &Units{}
}

func (u *Units) Parse(parser structure.ObjectParser) {
	u.BloodGlucose = parser.String("bg")
}

func (u *Units) Validate(validator structure.Validator) {
	validator.String("bg", u.BloodGlucose).Exists().OneOf(dataBloodGlucose.Units()...)
}

func (u *Units) Normalize(normalizer data.Normalizer) {
	if normalizer.Origin() == structure.OriginExternal {
		u.BloodGlucose = dataBloodGlucose.NormalizeUnits(u.BloodGlucose)
	}
}
