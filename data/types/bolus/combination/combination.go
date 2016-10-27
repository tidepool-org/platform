package combination

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/bolus"
)

type Combination struct {
	bolus.Bolus `bson:",inline"`

	Normal           *float64 `json:"normal,omitempty" bson:"normal,omitempty"`
	ExpectedNormal   *float64 `json:"expectedNormal,omitempty" bson:"expectedNormal,omitempty"`
	Extended         *float64 `json:"extended,omitempty" bson:"extended,omitempty"`
	ExpectedExtended *float64 `json:"expectedExtended,omitempty" bson:"expectedExtended,omitempty"`
	Duration         *int     `json:"duration,omitempty" bson:"duration,omitempty"`
	ExpectedDuration *int     `json:"expectedDuration,omitempty" bson:"expectedDuration,omitempty"`
}

func SubType() string {
	return "dual/square"
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

	c.Normal = nil
	c.ExpectedNormal = nil
	c.Extended = nil
	c.ExpectedExtended = nil
	c.Duration = nil
	c.ExpectedDuration = nil
}

func (c *Combination) Parse(parser data.ObjectParser) error {
	if err := c.Bolus.Parse(parser); err != nil {
		return err
	}

	c.Normal = parser.ParseFloat("normal")
	c.ExpectedNormal = parser.ParseFloat("expectedNormal")
	c.Extended = parser.ParseFloat("extended")
	c.ExpectedExtended = parser.ParseFloat("expectedExtended")
	c.Duration = parser.ParseInteger("duration")
	c.ExpectedDuration = parser.ParseInteger("expectedDuration")

	return nil
}

func (c *Combination) Validate(validator data.Validator) error {
	if err := c.Bolus.Validate(validator); err != nil {
		return err
	}

	validator.ValidateString("subType", &c.SubType).EqualTo(SubType())

	validator.ValidateFloat("normal", c.Normal).Exists().InRange(0.0, 100.0)

	expectedNormalValidator := validator.ValidateFloat("expectedNormal", c.ExpectedNormal)
	if c.Normal != nil {
		if *c.Normal == 0.0 {
			expectedNormalValidator.Exists()
		}
		expectedNormalValidator.InRange(*c.Normal, 100.0)
	} else {
		expectedNormalValidator.InRange(0.0, 100.0)
	}

	if c.ExpectedNormal != nil {
		validator.ValidateFloat("extended", c.Extended).Exists().EqualTo(0.0)
		validator.ValidateFloat("expectedExtended", c.ExpectedExtended).Exists().InRange(0.0, 100.0)
		validator.ValidateInteger("duration", c.Duration).Exists().EqualTo(0)
		validator.ValidateInteger("expectedDuration", c.ExpectedDuration).Exists().InRange(0, 86400000)
	} else {
		validator.ValidateFloat("extended", c.Extended).Exists().InRange(0.0, 100.0)

		expectedExtendedValidator := validator.ValidateFloat("expectedExtended", c.ExpectedExtended)
		if c.Extended != nil {
			if *c.Extended == 0.0 {
				expectedExtendedValidator.Exists()
			}
			expectedExtendedValidator.InRange(*c.Extended, 100.0)
		} else {
			expectedExtendedValidator.InRange(0.0, 100.0)
		}

		validator.ValidateInteger("duration", c.Duration).Exists().InRange(0, 86400000)

		expectedDurationValidator := validator.ValidateInteger("expectedDuration", c.ExpectedDuration)
		if c.Duration != nil {
			expectedDurationValidator.InRange(*c.Duration, 86400000)
		} else {
			expectedDurationValidator.InRange(0, 86400000)
		}
		if c.ExpectedExtended != nil {
			expectedDurationValidator.Exists()
		} else {
			expectedDurationValidator.NotExists()
		}
	}

	return nil
}
