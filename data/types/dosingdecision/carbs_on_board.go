package dosingdecision

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	CarbsOnBoardStartMaximum = 86400000
	CarbsOnBoardStartMinimum = 0
)

type CarbsOnBoard struct {
	StartDate *string  `json:"startDate,omitempty" bson:"startDate,omitempty"`
	EndDate   *string  `json:"endDate,omitempty" bson:"endDate,omitempty"`
	Quantity  *float64 `json:"quantity,omitempty" bson:"quantity,omitempty"`
}

func ParseCarbsOnBoard(parser structure.ObjectParser) *CarbsOnBoard {
	if !parser.Exists() {
		return nil
	}
	datum := NewCarbsOnBoard()
	parser.Parse(datum)
	return datum
}

func NewCarbsOnBoard() *CarbsOnBoard {
	return &CarbsOnBoard{}
}

func (i *CarbsOnBoard) Parse(parser structure.ObjectParser) {
	i.StartDate = parser.String("startDate")
	i.EndDate = parser.String("endDate")
	i.Quantity = parser.Float64("quantity")
}

func (i *CarbsOnBoard) Validate(validator structure.Validator) {
	validator.Float64("quantity", i.Quantity).Exists()
	validator.String("startDate", i.StartDate).Exists().AsTime(TimeFormat)
	validator.String("endDate", i.EndDate).Exists().AsTime(TimeFormat)
}

func (i *CarbsOnBoard) Normalize(normalizer data.Normalizer) {
	//if normalizer.Origin() == structure.OriginExternal {
	//	i.Amount = dataBloodGlucose.NormalizeValueForUnits(i.Amount, units)
	//}
}
