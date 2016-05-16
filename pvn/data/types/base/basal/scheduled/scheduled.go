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
	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/types/base/basal"
)

type Scheduled struct {
	basal.Basal

	Name     *string  `json:"scheduleName,omitempty" bson:"scheduleName,omitempty"`
	Rate     *float64 `json:"rate" bson:"rate"`
	Duration *int     `json:"duration" bson:"duration"`
}

func Type() string {
	return basal.Type()
}

func DeliveryType() string {
	return "scheduled"
}

func New() *Scheduled {
	scheduledType := Type()
	scheduledSubType := DeliveryType()

	scheduled := &Scheduled{}
	scheduled.Type = &scheduledType
	scheduled.DeliveryType = &scheduledSubType
	return scheduled
}

func (s *Scheduled) Parse(parser data.ObjectParser) {
	s.Basal.Parse(parser)
	s.Duration = parser.ParseInteger("duration")
	s.Rate = parser.ParseFloat("rate")
	s.Name = parser.ParseString("scheduleName")
}

func (s *Scheduled) Validate(validator data.Validator) {
	s.Basal.Validate(validator)
	validator.ValidateInteger("duration", s.Duration).Exists().GreaterThanOrEqualTo(0).LessThanOrEqualTo(432000000)
	validator.ValidateFloat("rate", s.Rate).Exists().GreaterThanOrEqualTo(0.0).LessThanOrEqualTo(20.0)
	validator.ValidateString("scheduleName", s.Name).LengthGreaterThan(1)
}

func (s *Scheduled) Normalize(normalizer data.Normalizer) {
	s.Basal.Normalize(normalizer)
}
