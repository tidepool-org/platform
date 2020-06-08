package iob

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	InsulinOnBoardMaximum = 250.0
	InsulinOnBoardMinimum = 0.0
)

type Iob struct {
	InsulinOnBoard *float64 `json:"insulinOnBoard,omitempty" bson:"insulinOnBoard,omitempty"`
}

func ParseIob(parser structure.ObjectParser) *Iob {
	if !parser.Exists() {
		return nil
	}
	datum := NewIob()
	parser.Parse(datum)
	return datum
}

func NewIob() *Iob {
	return &Iob{}
}

func (i *Iob) Parse(parser structure.ObjectParser) {
	i.InsulinOnBoard = parser.Float64("insulinOnBoard")
}

func (i *Iob) Validate(validator structure.Validator) {
	validator.Float64("insulinOnBoard", i.InsulinOnBoard).InRange(InsulinOnBoardMinimum, InsulinOnBoardMaximum)
}

func (i *Iob) Normalize(normalizer data.Normalizer) {
}
