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

func (s *Suspend) Parse(parser data.ObjectParser) {
	s.Basal.Parse(parser)

	s.Duration = parser.ParseInteger("duration")
}

func (s *Suspend) Validate(validator data.Validator) {
	s.Basal.Validate(validator)

	validator.ValidateInteger("duration", s.Duration).Exists().InRange(0, 86400000)
}
