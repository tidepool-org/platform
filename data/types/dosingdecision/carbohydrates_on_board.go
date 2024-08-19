package dosingdecision

import (
	"time"

	"github.com/tidepool-org/platform/structure"
)

const (
	CarbohydratesOnBoardAmountMaximum = 1000
	CarbohydratesOnBoardAmountMinimum = 0
)

type CarbohydratesOnBoard struct {
	Time   *time.Time `json:"time,omitempty" bson:"time,omitempty"`
	Amount *float64   `json:"amount,omitempty" bson:"amount,omitempty"`
}

func ParseCarbohydratesOnBoard(parser structure.ObjectParser) *CarbohydratesOnBoard {
	if !parser.Exists() {
		return nil
	}
	datum := NewCarbohydratesOnBoard()
	parser.Parse(datum)
	return datum
}

func NewCarbohydratesOnBoard() *CarbohydratesOnBoard {
	return &CarbohydratesOnBoard{}
}

func (c *CarbohydratesOnBoard) Parse(parser structure.ObjectParser) {
	c.Time = parser.Time("time", time.RFC3339Nano)
	c.Amount = parser.Float64("amount")
}

func (c *CarbohydratesOnBoard) Validate(validator structure.Validator) {
	validator.Float64("amount", c.Amount).Exists().InRange(CarbohydratesOnBoardAmountMinimum, CarbohydratesOnBoardAmountMaximum)
}
