package dosingdecision

import (
	"time"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	MinCarbsOnBoard = 0
	MaxCarbsOnBoard = 1000
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
	var startDate time.Time

	validator.Float64("quantity", i.Quantity).Exists().InRange(MinCarbsOnBoard, MaxCarbsOnBoard)
	validator.String("startDate", i.StartDate).Exists().AsTime(TimeFormat)
	val := validator.String("endDate", i.EndDate).Exists().AsTime(TimeFormat)

	if i.StartDate != nil {
		startDate, _ = time.Parse(time.RFC3339Nano, *i.StartDate)
		val.After(startDate)
	}

}

func (i *CarbsOnBoard) Normalize(normalizer data.Normalizer) {
}
