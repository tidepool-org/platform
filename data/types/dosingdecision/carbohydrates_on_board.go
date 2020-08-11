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
	StartTime *time.Time `json:"startTime,omitempty" bson:"startTime,omitempty"`
	EndTime   *time.Time `json:"endTime,omitempty" bson:"endTime,omitempty"`
	Amount    *float64   `json:"amount,omitempty" bson:"amount,omitempty"`
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
	c.StartTime = parser.Time("startTime", TimeFormat)
	c.EndTime = parser.Time("endTime", TimeFormat)
	c.Amount = parser.Float64("amount")
}

func (c *CarbohydratesOnBoard) Validate(validator structure.Validator) {
	if c.StartTime != nil {
		validator.Time("endTime", c.EndTime).After(*c.StartTime)
	}
	validator.Float64("amount", c.Amount).Exists().InRange(CarbohydratesOnBoardAmountMinimum, CarbohydratesOnBoardAmountMaximum)
}
