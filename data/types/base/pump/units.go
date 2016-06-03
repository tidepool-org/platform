package pump

import "github.com/tidepool-org/platform/data"

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
	validator.ValidateStringAsBloodGlucoseUnits("bg", u.BloodGlucose).Exists()
}

func (u *Units) Normalize(normalizer data.Normalizer) {
	u.BloodGlucose = normalizer.NormalizeBloodGlucose(u.BloodGlucose).Units()
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
