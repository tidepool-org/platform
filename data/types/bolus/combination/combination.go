package combination

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
	NormalMaximum   = 100.0
	NormalMinimum   = 0.0
)

type Combination struct {
	bolus.Bolus `bson:",inline"`

	Duration         *int     `json:"duration,omitempty" bson:"duration,omitempty"`
	DurationExpected *int     `json:"expectedDuration,omitempty" bson:"expectedDuration,omitempty"`
	Extended         *float64 `json:"extended,omitempty" bson:"extended,omitempty"`
	ExtendedExpected *float64 `json:"expectedExtended,omitempty" bson:"expectedExtended,omitempty"`
	Normal           *float64 `json:"normal,omitempty" bson:"normal,omitempty"`
	NormalExpected   *float64 `json:"expectedNormal,omitempty" bson:"expectedNormal,omitempty"`
}

func SubType() string {
	return "dual/square" // TODO: Rename Type to "bolus/combination"; remove SubType
}

func NewDatum() data.Datum {
	return New()
}

func New() *Combination {
	return &Combination{}
}

func Init() *Combination {
	combination := New()
	combination.Init()
	return combination
}

func (c *Combination) Init() {
	c.Bolus.Init()
	c.SubType = SubType()

	c.Duration = nil
	c.DurationExpected = nil
	c.Extended = nil
	c.ExtendedExpected = nil
	c.Normal = nil
	c.NormalExpected = nil
}

func (c *Combination) Parse(parser data.ObjectParser) error {
	if err := c.Bolus.Parse(parser); err != nil {
		return err
	}

	c.Duration = parser.ParseInteger("duration")
	c.DurationExpected = parser.ParseInteger("expectedDuration")
	c.Extended = parser.ParseFloat("extended")
	c.ExtendedExpected = parser.ParseFloat("expectedExtended")
	c.Normal = parser.ParseFloat("normal")
	c.NormalExpected = parser.ParseFloat("expectedNormal")

	return nil
}

func (c *Combination) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(c.Meta())
	}

	c.Bolus.Validate(validator)

	if c.SubType != "" {
		validator.String("subType", &c.SubType).EqualTo(SubType())
	}

	if c.NormalExpected != nil {
		validator.Int("duration", c.Duration).Exists().EqualTo(DurationMinimum)
		validator.Int("expectedDuration", c.DurationExpected).Exists().InRange(DurationMinimum, DurationMaximum)
		validator.Float64("extended", c.Extended).Exists().EqualTo(ExtendedMinimum)
		validator.Float64("expectedExtended", c.ExtendedExpected).Exists().InRange(ExtendedMinimum, ExtendedMaximum)
	} else {
		validator.Int("duration", c.Duration).Exists().InRange(DurationMinimum, DurationMaximum)
		expectedDurationValidator := validator.Int("expectedDuration", c.DurationExpected)
		if c.Duration != nil && *c.Duration >= DurationMinimum && *c.Duration <= DurationMaximum {
			expectedDurationValidator.InRange(*c.Duration, DurationMaximum)
		} else {
			expectedDurationValidator.InRange(DurationMinimum, DurationMaximum)
		}
		if c.ExtendedExpected != nil {
			expectedDurationValidator.Exists()
		} else {
			expectedDurationValidator.NotExists()
		}
		validator.Float64("extended", c.Extended).Exists().InRange(ExtendedMinimum, ExtendedMaximum)
		expectedExtendedValidator := validator.Float64("expectedExtended", c.ExtendedExpected)
		if c.Extended != nil && *c.Extended >= ExtendedMinimum && *c.Extended <= ExtendedMaximum {
			if *c.Extended == ExtendedMinimum {
				expectedExtendedValidator.Exists()
			}
			expectedExtendedValidator.InRange(*c.Extended, ExtendedMaximum)
		} else {
			expectedExtendedValidator.InRange(ExtendedMinimum, ExtendedMaximum)
		}
	}
	validator.Float64("normal", c.Normal).Exists().InRange(NormalMinimum, NormalMaximum)
	expectedNormalValidator := validator.Float64("expectedNormal", c.NormalExpected)
	if c.Normal != nil && *c.Normal >= NormalMinimum && *c.Normal <= NormalMaximum {
		if *c.Normal == NormalMinimum {
			expectedNormalValidator.Exists()
		}
		expectedNormalValidator.InRange(*c.Normal, NormalMaximum)
	} else {
		expectedNormalValidator.InRange(NormalMinimum, NormalMaximum)
	}
}

func (c *Combination) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(c.Meta())
	}

	c.Bolus.Normalize(normalizer)
}
