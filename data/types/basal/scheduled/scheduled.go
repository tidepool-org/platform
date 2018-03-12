package scheduled

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
)

const (
	DeliveryType = "scheduled" // TODO: Rename Type to "basal/scheduled"; remove DeliveryType

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

func New() *Scheduled {
	return &Scheduled{
		Basal: basal.New(DeliveryType),
	}
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
		validator.String("deliveryType", &s.DeliveryType).EqualTo(DeliveryType)
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

type SuppressedScheduled struct {
	Type         *string `json:"type,omitempty" bson:"type,omitempty"`
	DeliveryType *string `json:"deliveryType,omitempty" bson:"deliveryType,omitempty"`

	Annotations  *data.BlobArray `json:"annotations,omitempty" bson:"annotations,omitempty"`
	Rate         *float64        `json:"rate,omitempty" bson:"rate,omitempty"`
	ScheduleName *string         `json:"scheduleName,omitempty" bson:"scheduleName,omitempty"`
}

func ParseSuppressedScheduled(parser data.ObjectParser) *SuppressedScheduled {
	if parser.Object() == nil {
		return nil
	}
	suppressed := NewSuppressedScheduled()
	suppressed.Parse(parser)
	parser.ProcessNotParsed()
	return suppressed
}

func NewSuppressedScheduled() *SuppressedScheduled {
	return &SuppressedScheduled{
		Type:         pointer.String(basal.Type),
		DeliveryType: pointer.String(DeliveryType),
	}
}

func (s *SuppressedScheduled) Parse(parser data.ObjectParser) error {
	s.Type = parser.ParseString("type")
	s.DeliveryType = parser.ParseString("deliveryType")

	s.Annotations = data.ParseBlobArray(parser.NewChildArrayParser("annotations"))
	s.Rate = parser.ParseFloat("rate")
	s.ScheduleName = parser.ParseString("scheduleName")

	return nil
}

func (s *SuppressedScheduled) Validate(validator structure.Validator) {
	validator.String("type", s.Type).Exists().EqualTo(basal.Type)
	validator.String("deliveryType", s.DeliveryType).Exists().EqualTo(DeliveryType)

	if s.Annotations != nil {
		s.Annotations.Validate(validator.WithReference("annotations"))
	}
	validator.Float64("rate", s.Rate).Exists().InRange(RateMinimum, RateMaximum)
	validator.String("scheduleName", s.ScheduleName).NotEmpty()
}

func (s *SuppressedScheduled) Normalize(normalizer data.Normalizer) {
	if s.Annotations != nil {
		s.Annotations.Normalize(normalizer.WithReference("annotations"))
	}
}
