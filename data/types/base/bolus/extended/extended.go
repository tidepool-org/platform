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
	"github.com/tidepool-org/platform/data/types/base/bolus"
)

type Extended struct {
	bolus.Bolus `bson:",inline"`

	Duration *int     `json:"duration,omitempty" bson:"duration,omitempty"`
	Extended *float64 `json:"extended,omitempty" bson:"extended,omitempty"`
}

func SubType() string {
	return "square"
}

func New() (*Extended, error) {
	extendedBolus, err := bolus.New(SubType())
	if err != nil {
		return nil, err
	}

	return &Extended{
		Bolus: *extendedBolus,
	}, nil
}

func (e *Extended) Parse(parser data.ObjectParser) error {
	if err := e.Bolus.Parse(parser); err != nil {
		return err
	}

	e.Duration = parser.ParseInteger("duration")
	e.Extended = parser.ParseFloat("extended")

	return nil
}

func (e *Extended) Validate(validator data.Validator) error {
	if err := e.Bolus.Validate(validator); err != nil {
		return err
	}

	validator.ValidateInteger("duration", e.Duration).Exists().InRange(0, 86400000)
	validator.ValidateFloat("extended", e.Extended).Exists().InRange(0.0, 100.0)

	return nil
}
