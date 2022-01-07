package automated

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/metadata"
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

	DurationExpected   *int                 `json:"expectedDuration,omitempty" bson:"expectedDuration,omitempty"`
	InsulinFormulation *insulin.Formulation `json:"insulinFormulation,omitempty" bson:"insulinFormulation,omitempty"`
	ScheduleName       *string              `json:"scheduleName,omitempty" bson:"scheduleName,omitempty"`
}

func New() *Automated {
	return &Automated{
		Basal: basal.New(DeliveryType),
	}
}

func (a *Automated) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(a.Meta())
	}

	a.Basal.Parse(parser)

	a.Duration = parser.Int("duration")
	a.DurationExpected = parser.Int("expectedDuration")
	a.InsulinFormulation = insulin.ParseFormulation(parser.WithReferenceObjectParser("insulinFormulation"))
	a.Rate = parser.Float64("rate")
	a.ScheduleName = parser.String("scheduleName")
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
	if a.InsulinFormulation != nil {
		a.InsulinFormulation.Validate(validator.WithReference("insulinFormulation"))
	}
	validator.Float64("rate", a.Rate).Exists().InRange(RateMinimum, RateMaximum)
	validator.String("scheduleName", a.ScheduleName).NotEmpty()
}

// IsValid returns true if there is no error in the validator
func (a *Automated) IsValid(validator structure.Validator) bool {
	return !(validator.HasError())
}

func (a *Automated) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(a.Meta())
	}

	a.Basal.Normalize(normalizer)

	if a.InsulinFormulation != nil {
		a.InsulinFormulation.Normalize(normalizer.WithReference("insulinFormulation"))
	}
}

type SuppressedAutomated struct {
	Type         *string `json:"type,omitempty" bson:"type,omitempty"`
	DeliveryType *string `json:"deliveryType,omitempty" bson:"deliveryType,omitempty"`

	Annotations        *metadata.MetadataArray `json:"annotations,omitempty" bson:"annotations,omitempty"`
	InsulinFormulation *insulin.Formulation    `json:"insulinFormulation,omitempty" bson:"insulinFormulation,omitempty"`
	Rate               *float64                `json:"rate,omitempty" bson:"rate,omitempty"`
	ScheduleName       *string                 `json:"scheduleName,omitempty" bson:"scheduleName,omitempty"`
}

func ParseSuppressedAutomated(parser structure.ObjectParser) *SuppressedAutomated {
	if !parser.Exists() {
		return nil
	}
	datum := NewSuppressedAutomated()
	parser.Parse(datum)
	return datum
}

func NewSuppressedAutomated() *SuppressedAutomated {
	return &SuppressedAutomated{
		Type:         pointer.FromString(basal.Type),
		DeliveryType: pointer.FromString(DeliveryType),
	}
}

func (s *SuppressedAutomated) Parse(parser structure.ObjectParser) {
	s.Type = parser.String("type")
	s.DeliveryType = parser.String("deliveryType")

	s.Annotations = metadata.ParseMetadataArray(parser.WithReferenceArrayParser("annotations"))
	s.InsulinFormulation = insulin.ParseFormulation(parser.WithReferenceObjectParser("insulinFormulation"))
	s.Rate = parser.Float64("rate")
	s.ScheduleName = parser.String("scheduleName")
}

func (s *SuppressedAutomated) Validate(validator structure.Validator) {
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

func (s *SuppressedAutomated) Normalize(normalizer data.Normalizer) {
	if s.InsulinFormulation != nil {
		s.InsulinFormulation.Normalize(normalizer.WithReference("insulinFormulation"))
	}
}
