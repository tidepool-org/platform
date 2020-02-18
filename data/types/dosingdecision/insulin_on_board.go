package dosingdecision

import (
	"time"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	MinInsulinOnBoard = 0
	MaxInsulinOnBoard = 1000
	TimeFormat        = time.RFC3339Nano
)

type InsulinOnBoard struct {
	StartDate *string  `json:"startDate,omitempty" bson:"startDate,omitempty"`
	Value     *float64 `json:"value,omitempty" bson:"value,omitempty"`
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
	i.StartDate = parser.String("startDate")
	i.Value = parser.Float64("value")
}

func (i *InsulinOnBoard) Validate(validator structure.Validator) {
	validator.Float64("value", i.Value).Exists().InRange(MinInsulinOnBoard, MaxInsulinOnBoard)
	validator.String("startDate", i.StartDate).Exists().AsTime(TimeFormat)
}

func (i *InsulinOnBoard) Normalize(normalizer data.Normalizer, units *string) {
}
