package scheduled

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/basal"
)

type Scheduled struct {
	basal.Basal `bson:",inline"`

	Duration         *int     `json:"duration,omitempty" bson:"duration,omitempty"`
	ExpectedDuration *int     `json:"expectedDuration,omitempty" bson:"expectedDuration,omitempty"`
	Rate             *float64 `json:"rate,omitempty" bson:"rate,omitempty"`
	ScheduleName     *string  `json:"scheduleName,omitempty" bson:"scheduleName,omitempty"`
}

func DeliveryType() string {
	return "scheduled"
}

func NewDatum() data.Datum {
	return New()
}

func New() *Scheduled {
	return &Scheduled{}
}

func Init() *Scheduled {
	scheduled := New()
	scheduled.Init()
	return scheduled
}

func (s *Scheduled) Init() {
	s.Basal.Init()
	s.DeliveryType = DeliveryType()

	s.Duration = nil
	s.ExpectedDuration = nil
	s.Rate = nil
	s.ScheduleName = nil
}

func (s *Scheduled) Parse(parser data.ObjectParser) error {
	if err := s.Basal.Parse(parser); err != nil {
		return err
	}

	s.Duration = parser.ParseInteger("duration")
	s.ExpectedDuration = parser.ParseInteger("expectedDuration")
	s.Rate = parser.ParseFloat("rate")
	s.ScheduleName = parser.ParseString("scheduleName")

	return nil
}

func (s *Scheduled) Validate(validator data.Validator) error {
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

	validator.ValidateFloat("rate", s.Rate).Exists().InRange(0.0, 100.0)

	validator.ValidateString("scheduleName", s.ScheduleName).NotEmpty()

	return nil
}
