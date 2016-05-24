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
	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/types/base/basal"
)

type Suspend struct {
	basal.Basal

	Duration *int `json:"duration" bson:"duration"`
}

func Type() string {
	return basal.Type()
}

func DeliveryType() string {
	return "suspend"
}

func New() *Suspend {
	suspendType := Type()
	suspendSubType := DeliveryType()

	suspend := &Suspend{}
	suspend.Type = &suspendType
	suspend.DeliveryType = &suspendSubType
	return suspend
}

func (s *Suspend) Parse(parser data.ObjectParser) {
	s.Basal.Parse(parser)
	s.Duration = parser.ParseInteger("duration")
}

func (s *Suspend) Validate(validator data.Validator) {
	s.Basal.Validate(validator)
	validator.ValidateInteger("duration", s.Duration).Exists().InRange(0, 86400000)
}

func (s *Suspend) Normalize(normalizer data.Normalizer) {
	s.Basal.Normalize(normalizer)
}
