package scheduled

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

type Scheduled struct {
	basal.Basal `bson:",inline"`

	Duration *int     `json:"duration,omitempty" bson:"duration,omitempty"`
	Name     *string  `json:"scheduleName,omitempty" bson:"scheduleName,omitempty"` // TODO: Data model name UPDATE
	Rate     *float64 `json:"rate,omitempty" bson:"rate,omitempty"`
}

func DeliveryType() string {
	return "scheduled"
}

func New() (*Scheduled, error) {
	scheduledBasal, err := basal.New(DeliveryType())
	if err != nil {
		return nil, err
	}

	return &Scheduled{
		Basal: *scheduledBasal,
	}, nil
}

func (s *Scheduled) Parse(parser data.ObjectParser) error {
	if err := s.Basal.Parse(parser); err != nil {
		return err
	}

	s.Duration = parser.ParseInteger("duration")
	s.Name = parser.ParseString("scheduleName")
	s.Rate = parser.ParseFloat("rate")

	return nil
}

func (s *Scheduled) Validate(validator data.Validator) error {
	if err := s.Basal.Validate(validator); err != nil {
		return err
	}

	validator.ValidateInteger("duration", s.Duration).Exists().InRange(0, 432000000)
	validator.ValidateFloat("rate", s.Rate).Exists().InRange(0.0, 20.0)
	validator.ValidateString("scheduleName", s.Name).LengthGreaterThan(1)

	return nil
}
