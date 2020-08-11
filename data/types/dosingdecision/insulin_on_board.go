package dosingdecision

import (
	"time"

	"github.com/tidepool-org/platform/structure"
)

const (
	InsulinOnBoardAmountMaximum = 1000
	InsulinOnBoardAmountMinimum = -1000
)

type InsulinOnBoard struct {
	StartTime *time.Time `json:"startTime,omitempty" bson:"startTime,omitempty"`
	Amount    *float64   `json:"amount,omitempty" bson:"amount,omitempty"`
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
	i.StartTime = parser.Time("startTime", TimeFormat)
	i.Amount = parser.Float64("amount")
}

func (i *InsulinOnBoard) Validate(validator structure.Validator) {
	validator.Float64("amount", i.Amount).Exists().InRange(InsulinOnBoardAmountMinimum, InsulinOnBoardAmountMaximum)
}
