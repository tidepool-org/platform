package dosingdecision

import (
	"strconv"
	"time"

	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	ForecastBloodGlucoseArrayLengthMaximum = 60 * 24
)

type ForecastBloodGlucose struct {
	Time  *time.Time `json:"time,omitempty" bson:"time,omitempty"`
	Value *float64   `json:"value,omitempty" bson:"value,omitempty"`
}

func ParseForecastBloodGlucose(parser structure.ObjectParser) *ForecastBloodGlucose {
	if !parser.Exists() {
		return nil
	}
	datum := NewForecastBloodGlucose()
	parser.Parse(datum)
	return datum
}

func NewForecastBloodGlucose() *ForecastBloodGlucose {
	return &ForecastBloodGlucose{}
}

func (b *ForecastBloodGlucose) Parse(parser structure.ObjectParser) {
	b.Time = parser.Time("time", time.RFC3339Nano)
	b.Value = parser.Float64("value")
}

func (b *ForecastBloodGlucose) Validate(validator structure.Validator, units *string) {
	// TODO: https://tidepool.atlassian.net/browse/BACK-3842: validator.Time("time", b.Time).Exists()
	validator.Float64("value", b.Value).Exists() // No range validation as this is a forecast value
}

func (b *ForecastBloodGlucose) Normalize(normalizer data.Normalizer, units *string) {
	if normalizer.Origin() == structure.OriginExternal {
		b.Value = dataBloodGlucose.NormalizeValueForUnits(b.Value, units)
	}
}

type ForecastBloodGlucoseArray []*ForecastBloodGlucose

func ParseForecastBloodGlucoseArray(parser structure.ArrayParser) *ForecastBloodGlucoseArray {
	if !parser.Exists() {
		return nil
	}
	datum := NewForecastBloodGlucoseArray()
	parser.Parse(datum)
	return datum
}

func NewForecastBloodGlucoseArray() *ForecastBloodGlucoseArray {
	return &ForecastBloodGlucoseArray{}
}

func (b *ForecastBloodGlucoseArray) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*b = append(*b, ParseForecastBloodGlucose(parser.WithReferenceObjectParser(reference)))
	}
}

func (b *ForecastBloodGlucoseArray) Validate(validator structure.Validator, units *string) {
	if length := len(*b); length == 0 {
		validator.ReportError(structureValidator.ErrorValueEmpty())
	} else if length > BloodGlucoseArrayLengthMaximum {
		validator.ReportError(structureValidator.ErrorLengthNotLessThanOrEqualTo(length, BloodGlucoseArrayLengthMaximum))
	}
	for index, datum := range *b {
		if datumValidator := validator.WithReference(strconv.Itoa(index)); datum != nil {
			datum.Validate(datumValidator, units)
		} else {
			datumValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (b *ForecastBloodGlucoseArray) Normalize(normalizer data.Normalizer, units *string) {
	for index, datum := range *b {
		if datum != nil {
			datum.Normalize(normalizer.WithReference(strconv.Itoa(index)), units)
		}
	}
}
