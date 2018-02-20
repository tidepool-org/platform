package scheduled

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/structure"
)

const (
	DurationMaximum = 604800000
	DurationMinimum = 0
	RateMaximum     = 100.0
	RateMinimum     = 0.0
)

type Scheduled struct {
	basal.Basal `bson:",inline"`

	Duration         *int     `json:"duration,omitempty" bson:"duration,omitempty"`
	DurationExpected *int     `json:"expectedDuration,omitempty" bson:"expectedDuration,omitempty"`
	Rate             *float64 `json:"rate,omitempty" bson:"rate,omitempty"`
	ScheduleName     *string  `json:"scheduleName,omitempty" bson:"scheduleName,omitempty"`
}

func DeliveryType() string {
	return "scheduled" // TODO: Rename Type to "basal/scheduled"; remove DeliveryType
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
	s.DurationExpected = nil
	s.Rate = nil
	s.ScheduleName = nil
}

func (s *Scheduled) Parse(parser data.ObjectParser) error {
	if err := s.Basal.Parse(parser); err != nil {
		return err
	}

	s.Duration = parser.ParseInteger("duration")
	s.DurationExpected = parser.ParseInteger("expectedDuration")
	s.Rate = parser.ParseFloat("rate")
	s.ScheduleName = parser.ParseString("scheduleName")

	return nil
}

func (s *Scheduled) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(s.Meta())
	}

	s.Basal.Validate(validator)

	if s.DeliveryType != "" {
		validator.String("deliveryType", &s.DeliveryType).EqualTo(DeliveryType())
	}

	validator.Int("duration", s.Duration).Exists().InRange(DurationMinimum, DurationMaximum)
	expectedDurationValidator := validator.Int("expectedDuration", s.DurationExpected)
	if s.Duration != nil && *s.Duration >= DurationMinimum && *s.Duration <= DurationMaximum {
		expectedDurationValidator.InRange(*s.Duration, DurationMaximum)
	} else {
		expectedDurationValidator.InRange(DurationMinimum, DurationMaximum)
	}
	validator.Float64("rate", s.Rate).Exists().InRange(RateMinimum, RateMaximum)
	validator.String("scheduleName", s.ScheduleName).NotEmpty()
}

func (s *Scheduled) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(s.Meta())
	}

	s.Basal.Normalize(normalizer)
}
