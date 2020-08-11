package dosingdecision

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
	BloodGlucose *string `json:"bloodGlucose,omitempty" bson:"bloodGlucose,omitempty"`
	Carbohydrate *string `json:"carbohydrate,omitempty" bson:"carbohydrate,omitempty"`
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
	u.BloodGlucose = parser.String("bloodGlucose")
	u.Carbohydrate = parser.String("carbohydrate")
	u.Insulin = parser.String("insulin")
}

func (u *Units) Validate(validator structure.Validator) {
	validator.String("bloodGlucose", u.BloodGlucose).Exists().OneOf(dataBloodGlucose.Units()...)
	validator.String("carbohydrate", u.Carbohydrate).Exists().OneOf(CarbohydrateUnits()...)
	validator.String("insulin", u.Insulin).Exists().OneOf(InsulinUnits()...)
}

func (u *Units) Normalize(normalizer data.Normalizer) {
	if normalizer.Origin() == structure.OriginExternal {
		u.BloodGlucose = dataBloodGlucose.NormalizeUnits(u.BloodGlucose)
	}
}
