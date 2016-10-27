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
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/device"
)

type Status struct {
	device.Device `bson:",inline"`

	Name     *string                 `json:"status,omitempty" bson:"status,omitempty"`
	Duration *int                    `json:"duration,omitempty" bson:"duration,omitempty"`
	Reason   *map[string]interface{} `json:"reason,omitempty" bson:"reason,omitempty"`
}

func SubType() string {
	return "status"
}

func NewDatum() data.Datum {
	return New()
}

func New() *Status {
	return &Status{}
}

func Init() *Status {
	status := New()
	status.Init()
	return status
}

func (s *Status) Init() {
	s.Device.Init()
	s.SubType = SubType()

	s.Name = nil
	s.Duration = nil
	s.Reason = nil
}

func (s *Status) Parse(parser data.ObjectParser) error {
	if err := s.Device.Parse(parser); err != nil {
		return err
	}

	s.Duration = parser.ParseInteger("duration")
	s.Name = parser.ParseString("status")
	s.Reason = parser.ParseObject("reason")

	return nil
}

func (s *Status) Validate(validator data.Validator) error {
	if err := s.Device.Validate(validator); err != nil {
		return err
	}

	validator.ValidateString("subType", &s.SubType).EqualTo(SubType())

	validator.ValidateInteger("duration", s.Duration).GreaterThanOrEqualTo(0) // TODO_DATA: .Exists() - Suspend events on Animas do not have duration?
	validator.ValidateString("status", s.Name).Exists().OneOf([]string{"resumed", "suspended"})
	validator.ValidateObject("reason", s.Reason).Exists()

	return nil
}
