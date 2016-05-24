package status

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
	"github.com/tidepool-org/platform/pvn/data/types/base/device"
)

type Status struct {
	device.Device `bson:",inline"`

	Name     *string                 `json:"status" bson:"status"`
	Duration *int                    `json:"duration" bson:"duration"`
	Reason   *map[string]interface{} `json:"reason" bson:"reason"`
}

func Type() string {
	return device.Type()
}

func SubType() string {
	return "status"
}

func New() *Status {
	statusType := Type()
	statusSubType := SubType()

	status := &Status{}
	status.Type = &statusType
	status.SubType = &statusSubType
	return status
}

func (s *Status) Parse(parser data.ObjectParser) {
	s.Device.Parse(parser)
	s.Duration = parser.ParseInteger("duration")
	s.Name = parser.ParseString("status")
	s.Reason = parser.ParseObject("reason")
}

func (s *Status) Validate(validator data.Validator) {
	s.Device.Validate(validator)

	validator.ValidateInteger("duration", s.Duration).Exists().GreaterThanOrEqualTo(0)
	validator.ValidateString("status", s.Name).Exists().OneOf([]string{"suspended"})
	validator.ValidateObject("reason", s.Reason).Exists()

}

func (s *Status) Normalize(normalizer data.Normalizer) {
	s.Device.Normalize(normalizer)
}
