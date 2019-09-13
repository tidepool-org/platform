package status

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

type ForecastStruct struct {
	Forecast *data.Forecast `json:"forecast,omitempty" bson:"forecast,omitempty"`
}

func (f *ForecastStruct) statusObject() {
}

func ParseForecastStruct(parser structure.ObjectParser) *ForecastStruct {
	if !parser.Exists() {
		return nil
	}
	datum := NewForecastStruct()
	parser.Parse(datum)
	return datum
}

func NewForecastStruct() *ForecastStruct {
	return &ForecastStruct{}
}

func (f *ForecastStruct) Parse(parser structure.ObjectParser) {
	f.Forecast = data.ParseForecast(parser.WithReferenceObjectParser("battery"))
}
