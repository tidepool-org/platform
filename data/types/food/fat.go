package food

import (
	"github.com/tidepool-org/platform/structure"
)

const (
	FatTotalGramsMaximum = 1000.0
	FatTotalGramsMinimum = 0.0
	FatUnitsGrams        = "grams"
)

func FatUnits() []string {
	return []string{
		FatUnitsGrams,
	}
}

type Fat struct {
	Total *float64 `json:"total,omitempty" bson:"total,omitempty"`
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
}

func ParseFat(parser structure.ObjectParser) *Fat {
	if !parser.Exists() {
		return nil
	}
	datum := NewFat()
	parser.Parse(datum)
	return datum
}

func NewFat() *Fat {
	return &Fat{}
}

func (f *Fat) Parse(parser structure.ObjectParser) {
	f.Total = parser.Float64("total")
	f.Units = parser.String("units")
}

func (f *Fat) Validate(validator structure.Validator) {
	validator.Float64("total", f.Total).Exists().InRange(FatTotalGramsMinimum, FatTotalGramsMaximum)
	validator.String("units", f.Units).Exists().OneOf(FatUnits()...)
}
