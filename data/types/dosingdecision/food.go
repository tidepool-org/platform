package dosingdecision

import (
	"time"

	dataTypesFood "github.com/tidepool-org/platform/data/types/food"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type Food struct {
	Time      *time.Time               `json:"time,omitempty" bson:"time,omitempty"`
	Nutrition *dataTypesFood.Nutrition `json:"nutrition,omitempty" bson:"nutrition,omitempty"`
}

func ParseFood(parser structure.ObjectParser) *Food {
	if !parser.Exists() {
		return nil
	}
	datum := NewFood()
	parser.Parse(datum)
	return datum
}

func NewFood() *Food {
	return &Food{}
}

func (f *Food) Parse(parser structure.ObjectParser) {
	f.Time = parser.Time("time", time.RFC3339Nano)
	f.Nutrition = dataTypesFood.ParseNutrition(parser.WithReferenceObjectParser("nutrition"))
}

func (f *Food) Validate(validator structure.Validator) {
	if nutritionValidator := validator.WithReference("nutrition"); f.Nutrition != nil {
		f.Nutrition.Validate(nutritionValidator)
	} else {
		nutritionValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
}
