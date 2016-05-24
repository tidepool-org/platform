package pump

import (
	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/types/common/bloodglucose"
)

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
	validator.ValidateString("bg", u.BloodGlucose).Exists().OneOf([]string{common.Mmoll, common.MmolL, common.Mgdl, common.MgdL})
	validator.ValidateString("carb", u.Carbohydrate).Exists().LengthGreaterThanOrEqualTo(1)
}

func (u *Units) Normalize(normalizer data.Normalizer) {
	u.BloodGlucose = normalizer.NormalizeBloodGlucose("bg", u.BloodGlucose).NormalizeUnits()
}

func ParseUnits(parser data.ObjectParser) *Units {
	var units *Units
	if parser.Object() != nil {
		units = NewUnits()
		units.Parse(parser)
	}
	return units
}
