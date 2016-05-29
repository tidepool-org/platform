package suspend

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
	"github.com/tidepool-org/platform/data/types/base/basal"
)

type Suspend struct {
	basal.Basal `bson:",inline"`

	Duration *int `json:"duration,omitempty" bson:"duration,omitempty"`
}

func DeliveryType() string {
	return "suspend"
}

func New() (*Suspend, error) {
	suspendBasal, err := basal.New(DeliveryType())
	if err != nil {
		return nil, err
	}

	return &Suspend{
		Basal: *suspendBasal,
	}, nil
}

func (s *Suspend) Parse(parser data.ObjectParser) error {
	if err := s.Basal.Parse(parser); err != nil {
		return err
	}

	s.Duration = parser.ParseInteger("duration")

	return nil
}

func (s *Suspend) Validate(validator data.Validator) error {
	if err := s.Basal.Validate(validator); err != nil {
		return err
	}

	// NOTE: set to a max of one week as we don't yet understand what is acceptable
	validator.ValidateInteger("duration", s.Duration).InRange(0, 604800000)

	return nil
}
