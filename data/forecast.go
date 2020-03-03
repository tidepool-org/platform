package data

import (
	"time"

	"github.com/tidepool-org/platform/structure"
)

const (
	BloodGlucose   = "bloodGlucose"
	CarbsOnBoard   = "carbsOnBoard"
	InsulinOnBoard = "insulinOnBoard"
)

type Forecast struct {
	Date  *string  `json:"date,omitempty" bson:"date,omitempty"`
	Type  *string  `json:"type,omitempty" bson:"type,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

func Types() []string {
	return []string{
		BloodGlucose,
		CarbsOnBoard,
		InsulinOnBoard,
	}
}

func ParseForecast(parser structure.ObjectParser) *Forecast {
	if !parser.Exists() {
		return nil
	}
	datum := NewForecast()
	parser.Parse(datum)
	return datum
}

func NewForecast() *Forecast {
	return &Forecast{}
}

func (f *Forecast) Parse(parser structure.ObjectParser) {
	f.Date = parser.String("date")
	f.Type = parser.String("type")
	f.Value = parser.Float64("value")
}

func (f *Forecast) Validate(validator structure.Validator) {
	validator.String("date", f.Date).AsTime(time.RFC3339Nano)
	validator.String("type", f.Type).Exists().OneOf(Types()...)
}

func (f *Forecast) Normalize(normalizer Normalizer) {}
