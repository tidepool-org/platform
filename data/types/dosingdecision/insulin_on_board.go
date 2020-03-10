package dosingdecision

import (
	"github.com/tidepool-org/platform/structure"
)

const (
	InsulinOnBoardAmountMaximum = 1000
	InsulinOnBoardAmountMinimum = 0
)

type InsulinOnBoard struct {
	StartTime *string  `json:"startTime,omitempty" bson:"startTime,omitempty"`
	Amount    *float64 `json:"amount,omitempty" bson:"amount,omitempty"`
}

func ParseInsulinOnBoard(parser structure.ObjectParser) *InsulinOnBoard {
	if !parser.Exists() {
		return nil
	}
	datum := NewInsulinOnBoard()
	parser.Parse(datum)
	return datum
}

func NewInsulinOnBoard() *InsulinOnBoard {
	return &InsulinOnBoard{}
}

func (i *InsulinOnBoard) Parse(parser structure.ObjectParser) {
	i.StartTime = parser.String("startTime")
	i.Amount = parser.Float64("amount")
}

func (i *InsulinOnBoard) Validate(validator structure.Validator) {
	validator.String("startTime", i.StartTime).AsTime(TimeFormat)
	validator.Float64("amount", i.Amount).Exists().InRange(InsulinOnBoardAmountMinimum, InsulinOnBoardAmountMaximum)
}
