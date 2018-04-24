package extended

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/structure"
)

const (
	DurationMaximum = 86400000
	DurationMinimum = 0
	ExtendedMaximum = 100.0
	ExtendedMinimum = 0.0
)

type Extended struct {
	bolus.Bolus `bson:",inline"`

	Duration         *int     `json:"duration,omitempty" bson:"duration,omitempty"`
	DurationExpected *int     `json:"expectedDuration,omitempty" bson:"expectedDuration,omitempty"`
	Extended         *float64 `json:"extended,omitempty" bson:"extended,omitempty"`
	ExtendedExpected *float64 `json:"expectedExtended,omitempty" bson:"expectedExtended,omitempty"`
}

func SubType() string {
	return "square" // TODO: Rename Type to "bolus/extended"; remove SubType
}

func NewDatum() data.Datum {
	return New()
}

func New() *Extended {
	return &Extended{}
}

func Init() *Extended {
	extended := New()
	extended.Init()
	return extended
}

func (e *Extended) Init() {
	e.Bolus.Init()
	e.SubType = SubType()

	e.Duration = nil
	e.DurationExpected = nil
	e.Extended = nil
	e.ExtendedExpected = nil
}

func (e *Extended) Parse(parser data.ObjectParser) error {
	if err := e.Bolus.Parse(parser); err != nil {
		return err
	}

	e.Duration = parser.ParseInteger("duration")
	e.DurationExpected = parser.ParseInteger("expectedDuration")
	e.Extended = parser.ParseFloat("extended")
	e.ExtendedExpected = parser.ParseFloat("expectedExtended")

	return nil
}

func (e *Extended) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(e.Meta())
	}

	e.Bolus.Validate(validator)

	if e.SubType != "" {
		validator.String("subType", &e.SubType).EqualTo(SubType())
	}

	validator.Int("duration", e.Duration).Exists().InRange(DurationMinimum, DurationMaximum)
	durationExpectedValidator := validator.Int("expectedDuration", e.DurationExpected)
	if e.Duration != nil && *e.Duration >= DurationMinimum && *e.Duration <= DurationMaximum {
		durationExpectedValidator.InRange(*e.Duration, DurationMaximum)
	} else {
		durationExpectedValidator.InRange(DurationMinimum, DurationMaximum)
	}
	if e.ExtendedExpected != nil {
		durationExpectedValidator.Exists()
	} else {
		durationExpectedValidator.NotExists()
	}
	validator.Float64("extended", e.Extended).Exists().InRange(ExtendedMinimum, ExtendedMaximum)
	extendedExpectedValidator := validator.Float64("expectedExtended", e.ExtendedExpected)
	if e.Extended != nil && *e.Extended >= ExtendedMinimum && *e.Extended <= ExtendedMaximum {
		if *e.Extended == ExtendedMinimum {
			extendedExpectedValidator.Exists()
		}
		extendedExpectedValidator.InRange(*e.Extended, ExtendedMaximum)
	} else {
		extendedExpectedValidator.InRange(ExtendedMinimum, ExtendedMaximum)
	}
}

func (e *Extended) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(e.Meta())
	}

	e.Bolus.Normalize(normalizer)
}
