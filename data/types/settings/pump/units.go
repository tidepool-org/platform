package pump

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/blood/glucose"
)

type Units struct {
	Carbohydrate *string `json:"carb,omitempty" bson:"carb,omitempty"`
	BloodGlucose *string `json:"bg,omitempty" bson:"bg,omitempty"`
}

func NewUnits() *Units {
	return &Units{}
}

func (u *Units) Parse(parser data.ObjectParser) {
	u.Carbohydrate = parser.ParseString("carb")
	u.BloodGlucose = parser.ParseString("bg")
}

func (u *Units) Validate(validator data.Validator) {
	validator.ValidateString("carb", u.Carbohydrate).Exists().NotEmpty()
	validator.ValidateString("bg", u.BloodGlucose).Exists().OneOf(glucose.Units())
}

func (u *Units) Normalize(normalizer data.Normalizer) {
	u.BloodGlucose = glucose.NormalizeUnits(u.BloodGlucose)
}

func ParseUnits(parser data.ObjectParser) *Units {
	var units *Units
	if parser.Object() != nil {
		units = NewUnits()
		units.Parse(parser)
		parser.ProcessNotParsed()
	}
	return units
}
