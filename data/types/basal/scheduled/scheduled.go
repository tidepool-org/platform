package scheduled

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/metadata"
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

	Duration           *int                 `json:"duration,omitempty" bson:"duration,omitempty"`
	DurationExpected   *int                 `json:"expectedDuration,omitempty" bson:"expectedDuration,omitempty"`
	InsulinFormulation *insulin.Formulation `json:"insulinFormulation,omitempty" bson:"insulinFormulation,omitempty"`
	Rate               *float64             `json:"rate,omitempty" bson:"rate,omitempty"`
	ScheduleName       *string              `json:"scheduleName,omitempty" bson:"scheduleName,omitempty"`
}

func New() *Scheduled {
	return &Scheduled{
		Basal: basal.New(DeliveryType),
	}
}

func (s *Scheduled) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(s.Meta())
	}

	s.Basal.Parse(parser)

	s.Duration = parser.Int("duration")
	s.DurationExpected = parser.Int("expectedDuration")
	s.InsulinFormulation = insulin.ParseFormulation(parser.WithReferenceObjectParser("insulinFormulation"))
	s.Rate = parser.Float64("rate")
	s.ScheduleName = parser.String("scheduleName")
}

func (s *Scheduled) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(s.Meta())
	}

	s.Basal.Validate(validator)

	if s.DeliveryType != "" {
		validator.String("deliveryType", &s.DeliveryType).EqualTo(DeliveryType)
	}

	validator.Int("duration", s.Duration).Exists().InRangeWarning(DurationMinimum, DurationMaximum)
	expectedDurationValidator := validator.Int("expectedDuration", s.DurationExpected)
	if s.Duration != nil && *s.Duration >= DurationMinimum && *s.Duration <= DurationMaximum {
		expectedDurationValidator.InRangeWarning(*s.Duration, DurationMaximum)
	} else {
		expectedDurationValidator.InRangeWarning(DurationMinimum, DurationMaximum)
	}
	if s.InsulinFormulation != nil {
		s.InsulinFormulation.Validate(validator.WithReference("insulinFormulation"))
	}
	validator.Float64("rate", s.Rate).Exists().InRange(RateMinimum, RateMaximum)
	validator.String("scheduleName", s.ScheduleName).NotEmpty()
}

// IsValid returns true if there is no error and no warning in the validator
func (s *Scheduled) IsValid(validator structure.Validator) bool {
	return !(validator.HasError() || validator.HasWarning())
}

func (s *Scheduled) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(s.Meta())
	}

	s.Basal.Normalize(normalizer)

	if s.InsulinFormulation != nil {
		s.InsulinFormulation.Normalize(normalizer.WithReference("insulinFormulation"))
	}
}

type SuppressedScheduled struct {
	Type         *string `json:"type,omitempty" bson:"type,omitempty"`
	DeliveryType *string `json:"deliveryType,omitempty" bson:"deliveryType,omitempty"`

	Annotations        *metadata.MetadataArray `json:"annotations,omitempty" bson:"annotations,omitempty"`
	InsulinFormulation *insulin.Formulation    `json:"insulinFormulation,omitempty" bson:"insulinFormulation,omitempty"`
	Rate               *float64                `json:"rate,omitempty" bson:"rate,omitempty"`
	ScheduleName       *string                 `json:"scheduleName,omitempty" bson:"scheduleName,omitempty"`
}

func ParseSuppressedScheduled(parser structure.ObjectParser) *SuppressedScheduled {
	if !parser.Exists() {
		return nil
	}
	datum := NewSuppressedScheduled()
	parser.Parse(datum)
	return datum
}

func NewSuppressedScheduled() *SuppressedScheduled {
	return &SuppressedScheduled{
		Type:         pointer.FromString(basal.Type),
		DeliveryType: pointer.FromString(DeliveryType),
	}
}

func (s *SuppressedScheduled) Parse(parser structure.ObjectParser) {
	s.Type = parser.String("type")
	s.DeliveryType = parser.String("deliveryType")

	s.Annotations = metadata.ParseMetadataArray(parser.WithReferenceArrayParser("annotations"))
	s.InsulinFormulation = insulin.ParseFormulation(parser.WithReferenceObjectParser("insulinFormulation"))
	s.Rate = parser.Float64("rate")
	s.ScheduleName = parser.String("scheduleName")
}

func (s *SuppressedScheduled) Validate(validator structure.Validator) {
	validator.String("type", s.Type).Exists().EqualTo(basal.Type)
	validator.String("deliveryType", s.DeliveryType).Exists().EqualTo(DeliveryType)

	if s.Annotations != nil {
		s.Annotations.Validate(validator.WithReference("annotations"))
	}
	if s.InsulinFormulation != nil {
		s.InsulinFormulation.Validate(validator.WithReference("insulinFormulation"))
	}
	validator.Float64("rate", s.Rate).Exists().InRange(RateMinimum, RateMaximum)
	validator.String("scheduleName", s.ScheduleName).NotEmpty()
}

func (s *SuppressedScheduled) Normalize(normalizer data.Normalizer) {
	if s.InsulinFormulation != nil {
		s.InsulinFormulation.Normalize(normalizer.WithReference("insulinFormulation"))
	}
}
