package data

import (
	"time"

	"github.com/tidepool-org/platform/structure"
)

const (
	MinimumTimeScale = 0
	MaximumTimeScale = 3600000
	BloodGlucose     = "bloodGlucose"
	CarbsOnBoard     = "carbsOnBoard"
	InsulinOnBoard   = "insulinOnBoard"
)

type Forecast struct {
	StartTime *string    `json:"startTime,omitempty" bson:"startTime,omitempty"`
	TimeScale *int       `json:"timeScale,omitempty" bson:"timeScale,omitempty"`
	Type      *string    `json:"type,omitempty" bson:"type,omitempty"`
	Unit      *string    `json:"unit,omitempty" bson:"unit,omitempty"`
	Values    *[]float64 `json:"values,omitempty" bson:"values,omitempty"`
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
	f.StartTime = parser.String("startTime")
	f.TimeScale = parser.Int("timeScale")
	f.Type = parser.String("type")
	f.Unit = parser.String("unit")
	f.Values = parser.Float64Array("values")
}

func (f *Forecast) Validate(validator structure.Validator) {
	if f.StartTime != nil {
		validator.String("startTime", f.StartTime).AsTime(time.RFC3339Nano)
	}
	validator.Int("timeScale", f.TimeScale).Exists().InRange(MinimumTimeScale, MaximumTimeScale)
	validator.String("type", f.Type).Exists().OneOf(Types()...)
	validator.String("unit", f.Unit).Exists()
}

func (f *Forecast) Normalize(normalizer Normalizer) {}
