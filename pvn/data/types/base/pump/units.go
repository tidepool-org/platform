package pump

import "github.com/tidepool-org/platform/pvn/data"

type Units struct {
	Carbohydrate *string `json:"carb" bson:"carb"`
	BloodGlucose *string `json:"bg" bson:"bg"`
}

func NewUnits() *Units {
	return &Units{}
}

func (u *Units) Parse(parser data.ObjectParser) {
	u.Carbohydrate = parser.ParseString("carb")
	u.BloodGlucose = parser.ParseString("bg")
}

func (u *Units) Validate(validator data.Validator) {
	validator.ValidateString("bg", u.BloodGlucose).Exists().OneOf([]string{"mmol/l", "mmol/L", "mg/dl", "mg/dL"})
	validator.ValidateString("carb", u.Carbohydrate).Exists().LengthGreaterThanOrEqualTo(1)
}

func (u *Units) Normalize(normalizer data.Normalizer) {
}

func ParseUnits(parser data.ObjectParser) *Units {
	var units *Units
	if parser.Object() != nil {
		units = NewUnits()
		units.Parse(parser)
	}
	return units
}
