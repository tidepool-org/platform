package dosingdecision

import (
	"strconv"
	"time"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	ForecastArrayLengthMaximum = 10000
)

type Forecast struct {
	Time  *string  `json:"time,omitempty" bson:"time,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
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
	f.Time = parser.String("time")
	f.Value = parser.Float64("value")
}

func (f *Forecast) Validate(validator structure.Validator) {
	validator.String("time", f.Time).Exists().AsTime(time.RFC3339Nano)
	validator.Float64("value", f.Value).Exists()
}

type ForecastArray []*Forecast

func ParseForecastArray(parser structure.ArrayParser) *ForecastArray {
	if !parser.Exists() {
		return nil
	}
	datum := NewForecastArray()
	parser.Parse(datum)
	return datum
}

func NewForecastArray() *ForecastArray {
	return &ForecastArray{}
}

func (f *ForecastArray) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*f = append(*f, ParseForecast(parser.WithReferenceObjectParser(reference)))
	}
}

func (f *ForecastArray) Validate(validator structure.Validator) {
	if length := len(*f); length == 0 {
		validator.ReportError(structureValidator.ErrorValueEmpty())
	} else if length > ForecastArrayLengthMaximum {
		validator.ReportError(structureValidator.ErrorLengthNotLessThanOrEqualTo(length, ForecastArrayLengthMaximum))
	}
	for index, datum := range *f {
		if datumValidator := validator.WithReference(strconv.Itoa(index)); datum != nil {
			datum.Validate(datumValidator)
		} else {
			datumValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}
