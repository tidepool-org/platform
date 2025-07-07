package extended

import (
	"github.com/tidepool-org/platform/data"
	dataTypesBolus "github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/structure"
)

const (
	SubType = "square" // TODO: Rename Type to "bolus/extended"; remove SubType

	DurationMaximum = 86400000
	DurationMinimum = 0
	ExtendedMaximum = 100.0
	ExtendedMinimum = 0.0
)

type ExtendedFields struct {
	Duration         *int     `json:"duration,omitempty" bson:"duration,omitempty"`
	DurationExpected *int     `json:"expectedDuration,omitempty" bson:"expectedDuration,omitempty"`
	Extended         *float64 `json:"extended,omitempty" bson:"extended,omitempty"`
	ExtendedExpected *float64 `json:"expectedExtended,omitempty" bson:"expectedExtended,omitempty"`
}

func (e *ExtendedFields) Parse(parser structure.ObjectParser) {
	e.Duration = parser.Int("duration")
	e.DurationExpected = parser.Int("expectedDuration")
	e.Extended = parser.Float64("extended")
	e.ExtendedExpected = parser.Float64("expectedExtended")
}

func (e *ExtendedFields) Validate(validator structure.Validator) {
	durationValidator := validator.Int("duration", e.Duration).Exists()
	if e.DurationExpected != nil && structure.InRange(*e.DurationExpected, DurationMinimum, DurationMaximum) {
		durationValidator.InRange(DurationMinimum, *e.DurationExpected)
	} else {
		durationValidator.InRange(DurationMinimum, DurationMaximum)
	}
	durationExpectedValidator := validator.Int("expectedDuration", e.DurationExpected).InRange(DurationMinimum, DurationMaximum)
	if e.ExtendedExpected != nil {
		durationExpectedValidator.Exists()
	} else {
		durationExpectedValidator.NotExists()
	}

	extendedValidator := validator.Float64("extended", e.Extended).Exists()
	if e.ExtendedExpected != nil && structure.InRange(*e.ExtendedExpected, ExtendedMinimum, ExtendedMaximum) {
		extendedValidator.InRange(ExtendedMinimum, *e.ExtendedExpected)
	} else {
		extendedValidator.InRange(ExtendedMinimum, ExtendedMaximum)
	}
	validator.Float64("expectedExtended", e.ExtendedExpected).InRange(ExtendedMinimum, ExtendedMaximum)
}

type Extended struct {
	dataTypesBolus.Bolus `bson:",inline"`

	ExtendedFields `bson:",inline"`
}

func New() *Extended {
	return &Extended{
		Bolus: dataTypesBolus.New(SubType),
	}
}

func (e *Extended) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(e.Meta())
	}

	e.Bolus.Parse(parser)

	e.ExtendedFields.Parse(parser)
}

func (e *Extended) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(e.Meta())
	}

	e.Bolus.Validate(validator)

	if e.SubType != "" {
		validator.String("subType", &e.SubType).EqualTo(SubType)
	}

	e.ExtendedFields.Validate(validator)
}

func (e *Extended) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(e.Meta())
	}

	e.Bolus.Normalize(normalizer)
}
