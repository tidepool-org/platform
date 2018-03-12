package automated

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
)

const (
	DeliveryType = "automated" // TODO: Rename Type to "basal/automated"; remove DeliveryType

	DurationMaximum = 604800000
	DurationMinimum = 0
	RateMaximum     = 100.0
	RateMinimum     = 0.0
)

type Automated struct {
	basal.Basal `bson:",inline"`

	Duration         *int     `json:"duration,omitempty" bson:"duration,omitempty"`
	DurationExpected *int     `json:"expectedDuration,omitempty" bson:"expectedDuration,omitempty"`
	Rate             *float64 `json:"rate,omitempty" bson:"rate,omitempty"`
	ScheduleName     *string  `json:"scheduleName,omitempty" bson:"scheduleName,omitempty"`
}

func New() *Automated {
	return &Automated{
		Basal: basal.New(DeliveryType),
	}
}

func (a *Automated) Parse(parser data.ObjectParser) error {
	if err := a.Basal.Parse(parser); err != nil {
		return err
	}

	a.Duration = parser.ParseInteger("duration")
	a.DurationExpected = parser.ParseInteger("expectedDuration")
	a.Rate = parser.ParseFloat("rate")
	a.ScheduleName = parser.ParseString("scheduleName")

	return nil
}

func (a *Automated) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(a.Meta())
	}

	a.Basal.Validate(validator)

	if a.DeliveryType != "" {
		validator.String("deliveryType", &a.DeliveryType).EqualTo(DeliveryType)
	}

	validator.Int("duration", a.Duration).Exists().InRange(DurationMinimum, DurationMaximum)
	expectedDurationValidator := validator.Int("expectedDuration", a.DurationExpected)
	if a.Duration != nil && *a.Duration >= DurationMinimum && *a.Duration <= DurationMaximum {
		expectedDurationValidator.InRange(*a.Duration, DurationMaximum)
	} else {
		expectedDurationValidator.InRange(DurationMinimum, DurationMaximum)
	}
	validator.Float64("rate", a.Rate).Exists().InRange(RateMinimum, RateMaximum)
	validator.String("scheduleName", a.ScheduleName).NotEmpty()
}

func (a *Automated) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(a.Meta())
	}

	a.Basal.Normalize(normalizer)
}

type SuppressedAutomated struct {
	Type         *string `json:"type,omitempty" bson:"type,omitempty"`
	DeliveryType *string `json:"deliveryType,omitempty" bson:"deliveryType,omitempty"`

	Annotations  *data.BlobArray `json:"annotations,omitempty" bson:"annotations,omitempty"`
	Rate         *float64        `json:"rate,omitempty" bson:"rate,omitempty"`
	ScheduleName *string         `json:"scheduleName,omitempty" bson:"scheduleName,omitempty"`
}

func ParseSuppressedAutomated(parser data.ObjectParser) *SuppressedAutomated {
	if parser.Object() == nil {
		return nil
	}
	suppressed := NewSuppressedAutomated()
	suppressed.Parse(parser)
	parser.ProcessNotParsed()
	return suppressed
}

func NewSuppressedAutomated() *SuppressedAutomated {
	return &SuppressedAutomated{
		Type:         pointer.String(basal.Type),
		DeliveryType: pointer.String(DeliveryType),
	}
}

func (s *SuppressedAutomated) Parse(parser data.ObjectParser) error {
	s.Type = parser.ParseString("type")
	s.DeliveryType = parser.ParseString("deliveryType")

	s.Annotations = data.ParseBlobArray(parser.NewChildArrayParser("annotations"))
	s.Rate = parser.ParseFloat("rate")
	s.ScheduleName = parser.ParseString("scheduleName")

	return nil
}

func (s *SuppressedAutomated) Validate(validator structure.Validator) {
	validator.String("type", s.Type).Exists().EqualTo(basal.Type)
	validator.String("deliveryType", s.DeliveryType).Exists().EqualTo(DeliveryType)

	if s.Annotations != nil {
		s.Annotations.Validate(validator.WithReference("annotations"))
	}
	validator.Float64("rate", s.Rate).Exists().InRange(RateMinimum, RateMaximum)
	validator.String("scheduleName", s.ScheduleName).NotEmpty()
}

func (s *SuppressedAutomated) Normalize(normalizer data.Normalizer) {
	if s.Annotations != nil {
		s.Annotations.Normalize(normalizer.WithReference("annotations"))
	}
}
