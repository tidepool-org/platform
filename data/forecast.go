package data

import (
	"github.com/tidepool-org/platform/structure"
)

type Forecast struct {
	StartTime *string   `json:"startTime,omitempty" bson:"startTime,omitempty"`
	TimeScale *int      `json:"timeScale,omitempty" bson:"timeScale,omitempty"`
	Type      *string   `json:"type,omitempty" bson:"type,omitempty"`
	Unit      *string   `json:"unit,omitempty" bson:"unit,omitempty"`
	Values    []float64 `json:"values,omitempty" bson:"values,omitempty"`
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

func (a *Forecast) Parse(parser structure.ObjectParser) {
}

func (a *Forecast) Validate(validator structure.Validator) {
}

func (a *Forecast) Normalize(normalizer Normalizer) {}
