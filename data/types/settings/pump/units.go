package pump

import (
	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/structure"
)

const (
	CarbohydrateUnitsExchanges = "exchanges"
	CarbohydrateUnitsGrams     = "grams"
	InsulinUnitsUnits          = "Units"
)

func CarbohydrateUnits() []string {
	return []string{
		CarbohydrateUnitsExchanges,
		CarbohydrateUnitsGrams,
	}
}

func InsulinUnits() []string {
	return []string{
		InsulinUnitsUnits,
	}
}

type Units struct {
	BloodGlucose *string `json:"bg,omitempty" bson:"bg,omitempty"`     // TODO: Rename "bloodGlucose"
	Carbohydrate *string `json:"carb,omitempty" bson:"carb,omitempty"` // TODO: Rename "carbohydrate"
	Insulin      *string `json:"insulin,omitempty" bson:"insulin,omitempty"`
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
	u.Carbohydrate = parser.String("carb")
	u.Insulin = parser.String("insulin")
}

func (u *Units) Validate(validator structure.Validator) {
	validator.String("bg", u.BloodGlucose).Exists().OneOf(dataBloodGlucose.Units()...)
	validator.String("carb", u.Carbohydrate).Exists().OneOf(CarbohydrateUnits()...)
	validator.String("insulin", u.Insulin).OneOf(InsulinUnits()...)
}

func (u *Units) Normalize(normalizer data.Normalizer) {
	if normalizer.Origin() == structure.OriginExternal {
		u.BloodGlucose = dataBloodGlucose.NormalizeUnits(u.BloodGlucose)
	}
}
