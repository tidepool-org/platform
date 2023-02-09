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
	BloodGlucoseArrayLengthMaximum = 60 * 24
)

type BloodGlucose struct {
	Time  *time.Time `json:"time,omitempty" bson:"time,omitempty"`
	Value *float64   `json:"value,omitempty" bson:"value,omitempty"`
}

func ParseBloodGlucose(parser structure.ObjectParser) *BloodGlucose {
	if !parser.Exists() {
		return nil
	}
	datum := NewBloodGlucose()
	parser.Parse(datum)
	return datum
}

func NewBloodGlucose() *BloodGlucose {
	return &BloodGlucose{}
}

func (b *BloodGlucose) Parse(parser structure.ObjectParser) {
	b.Time = parser.Time("time", time.RFC3339Nano)
	b.Value = parser.Float64("value")
}

func (b *BloodGlucose) Validate(validator structure.Validator, units *string) {
	validator.Float64("value", b.Value).Exists().InRange(dataBloodGlucose.ValueRangeForUnits(units))
}

func (b *BloodGlucose) Normalize(normalizer data.Normalizer, units *string) {
	if normalizer.Origin() == structure.OriginExternal {
		b.Value = dataBloodGlucose.NormalizeValueForUnits(b.Value, units)
	}
}

type BloodGlucoseArray []*BloodGlucose

func ParseBloodGlucoseArray(parser structure.ArrayParser) *BloodGlucoseArray {
	if !parser.Exists() {
		return nil
	}
	datum := NewBloodGlucoseArray()
	parser.Parse(datum)
	return datum
}

func NewBloodGlucoseArray() *BloodGlucoseArray {
	return &BloodGlucoseArray{}
}

func (b *BloodGlucoseArray) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*b = append(*b, ParseBloodGlucose(parser.WithReferenceObjectParser(reference)))
	}
}

func (b *BloodGlucoseArray) Validate(validator structure.Validator, units *string) {
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

func (b *BloodGlucoseArray) Normalize(normalizer data.Normalizer, units *string) {
	for index, datum := range *b {
		if datum != nil {
			datum.Normalize(normalizer.WithReference(strconv.Itoa(index)), units)
		}
	}
}
