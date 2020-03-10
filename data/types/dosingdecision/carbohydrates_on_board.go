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
	StartTime *string  `json:"startTime,omitempty" bson:"startTime,omitempty"`
	EndTime   *string  `json:"endTime,omitempty" bson:"endTime,omitempty"`
	Amount    *float64 `json:"amount,omitempty" bson:"amount,omitempty"`
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
	c.StartTime = parser.String("startTime")
	c.EndTime = parser.String("endTime")
	c.Amount = parser.Float64("amount")
}

func (c *CarbohydratesOnBoard) Validate(validator structure.Validator) {
	var startTime time.Time

	if c.StartTime != nil {
		startTime, _ = time.Parse(time.RFC3339Nano, *c.StartTime)
	}

	validator.String("startTime", c.StartTime).AsTime(TimeFormat)
	validator.String("endTime", c.EndTime).AsTime(TimeFormat).After(startTime)
	validator.Float64("amount", c.Amount).Exists().InRange(CarbohydratesOnBoardAmountMinimum, CarbohydratesOnBoardAmountMaximum)
}
