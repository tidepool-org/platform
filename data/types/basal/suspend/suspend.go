package suspend

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/basal"
)

type Suspend struct {
	basal.Basal `bson:",inline"`

	Duration         *int `json:"duration,omitempty" bson:"duration,omitempty"`
	ExpectedDuration *int `json:"expectedDuration,omitempty" bson:"expectedDuration,omitempty"`

	Suppressed *basal.Suppressed `json:"suppressed,omitempty" bson:"suppressed,omitempty"`
}

func DeliveryType() string {
	return "suspend"
}

func NewDatum() data.Datum {
	return New()
}

func New() *Suspend {
	return &Suspend{}
}

func Init() *Suspend {
	suspend := New()
	suspend.Init()
	return suspend
}

func (s *Suspend) Init() {
	s.Basal.Init()
	s.DeliveryType = DeliveryType()

	s.Duration = nil
	s.ExpectedDuration = nil

	s.Suppressed = nil
}

func (s *Suspend) Parse(parser data.ObjectParser) error {
	if err := s.Basal.Parse(parser); err != nil {
		return err
	}

	s.Duration = parser.ParseInteger("duration")
	s.ExpectedDuration = parser.ParseInteger("expectedDuration")

	s.Suppressed = basal.ParseSuppressed(parser.NewChildObjectParser("suppressed"))

	return nil
}

func (s *Suspend) Validate(validator data.Validator) error {
	if err := s.Basal.Validate(validator); err != nil {
		return err
	}

	validator.ValidateString("deliveryType", &s.DeliveryType).EqualTo(DeliveryType())

	validator.ValidateInteger("duration", s.Duration).Exists().InRange(0, 604800000)

	expectedDurationValidator := validator.ValidateInteger("expectedDuration", s.ExpectedDuration)
	if s.Duration != nil {
		expectedDurationValidator.InRange(*s.Duration, 604800000)
	} else {
		expectedDurationValidator.InRange(0, 604800000)
	}

	if s.Suppressed != nil {
		s.Suppressed.Validate(validator.NewChildValidator("suppressed"), []string{"scheduled", "temp"})
	}

	return nil
}
