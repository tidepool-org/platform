package extended

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

type Extended struct {
	bolus.Bolus `bson:",inline"`

	Extended         *float64 `json:"extended,omitempty" bson:"extended,omitempty"`
	ExpectedExtended *float64 `json:"expectedExtended,omitempty" bson:"expectedExtended,omitempty"`
	Duration         *int     `json:"duration,omitempty" bson:"duration,omitempty"`
	ExpectedDuration *int     `json:"expectedDuration,omitempty" bson:"expectedDuration,omitempty"`
}

func SubType() string {
	return "square"
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

	e.Extended = nil
	e.ExpectedExtended = nil
	e.Duration = nil
	e.ExpectedDuration = nil
}

func (e *Extended) Parse(parser data.ObjectParser) error {
	if err := e.Bolus.Parse(parser); err != nil {
		return err
	}

	e.Extended = parser.ParseFloat("extended")
	e.ExpectedExtended = parser.ParseFloat("expectedExtended")
	e.Duration = parser.ParseInteger("duration")
	e.ExpectedDuration = parser.ParseInteger("expectedDuration")

	return nil
}

func (e *Extended) Validate(validator data.Validator) error {
	if err := e.Bolus.Validate(validator); err != nil {
		return err
	}

	validator.ValidateString("subType", &e.SubType).EqualTo(SubType())

	validator.ValidateFloat("extended", e.Extended).Exists().InRange(0.0, 100.0)

	expectedExtendedValidator := validator.ValidateFloat("expectedExtended", e.ExpectedExtended)
	if e.Extended != nil {
		if *e.Extended == 0.0 {
			expectedExtendedValidator.Exists()
		}
		expectedExtendedValidator.InRange(*e.Extended, 100.0)
	} else {
		expectedExtendedValidator.InRange(0.0, 100.0)
	}

	validator.ValidateInteger("duration", e.Duration).Exists().InRange(0, 86400000)

	expectedDurationValidator := validator.ValidateInteger("expectedDuration", e.ExpectedDuration)
	if e.Duration != nil {
		expectedDurationValidator.InRange(*e.Duration, 86400000)
	} else {
		expectedDurationValidator.InRange(0, 86400000)
	}
	if e.ExpectedExtended != nil {
		expectedDurationValidator.Exists()
	} else {
		expectedDurationValidator.NotExists()
	}

	return nil
}
