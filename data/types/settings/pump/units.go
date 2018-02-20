package pump

import (
	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/structure"
)

const (
	CarbohydrateExchanges = "exchanges"
	CarbohydrateGrams     = "grams"
)

func Carbohydrates() []string {
	return []string{
		CarbohydrateExchanges,
		CarbohydrateGrams,
	}
}

type Units struct {
	BloodGlucose *string `json:"bg,omitempty" bson:"bg,omitempty"`     // TODO: Rename "bloodGlucose"
	Carbohydrate *string `json:"carb,omitempty" bson:"carb,omitempty"` // TODO: Rename "carbohydrate"
}

func ParseUnits(parser data.ObjectParser) *Units {
	if parser.Object() == nil {
		return nil
	}
	units := NewUnits()
	units.Parse(parser)
	parser.ProcessNotParsed()
	return units
}

func NewUnits() *Units {
	return &Units{}
}

func (u *Units) Parse(parser data.ObjectParser) {
	u.BloodGlucose = parser.ParseString("bg")
	u.Carbohydrate = parser.ParseString("carb")
}

func (u *Units) Validate(validator structure.Validator) {
	validator.String("bg", u.BloodGlucose).Exists().OneOf(dataBloodGlucose.Units()...)
	validator.String("carb", u.Carbohydrate).Exists().OneOf(Carbohydrates()...)
}

func (u *Units) Normalize(normalizer data.Normalizer) {
	if normalizer.Origin() == structure.OriginExternal {
		u.BloodGlucose = dataBloodGlucose.NormalizeUnits(u.BloodGlucose)
	}
}
