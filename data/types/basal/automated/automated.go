package automated

import (
	"github.com/tidepool-org/platform/data"
	dataTypesBasal "github.com/tidepool-org/platform/data/types/basal"
	dataTypesBasalScheduled "github.com/tidepool-org/platform/data/types/basal/scheduled"
	dataTypesInsulin "github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	DeliveryType = "automated" // TODO: Rename Type to "basal/automated"; remove DeliveryType

	DurationMaximum = 604800000
	DurationMinimum = 0
	RateMaximum     = 100.0
	RateMinimum     = 0.0
)

type Suppressed interface {
	Parse(parser structure.ObjectParser)
	Validate(validator structure.Validator)
	Normalize(normalizer data.Normalizer)
}

type Automated struct {
	dataTypesBasal.Basal `bson:",inline"`

	Duration           *int                          `json:"duration,omitempty" bson:"duration,omitempty"`
	DurationExpected   *int                          `json:"expectedDuration,omitempty" bson:"expectedDuration,omitempty"`
	InsulinFormulation *dataTypesInsulin.Formulation `json:"insulinFormulation,omitempty" bson:"insulinFormulation,omitempty"`
	Rate               *float64                      `json:"rate,omitempty" bson:"rate,omitempty"`
	ScheduleName       *string                       `json:"scheduleName,omitempty" bson:"scheduleName,omitempty"`
	Suppressed         Suppressed                    `json:"suppressed,omitempty" bson:"suppressed,omitempty"`
}

func New() *Automated {
	return &Automated{
		Basal: dataTypesBasal.New(DeliveryType),
	}
}

func (a *Automated) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(a.Meta())
	}

	a.Basal.Parse(parser)

	a.Duration = parser.Int("duration")
	a.DurationExpected = parser.Int("expectedDuration")
	a.InsulinFormulation = dataTypesInsulin.ParseFormulation(parser.WithReferenceObjectParser("insulinFormulation"))
	a.Rate = parser.Float64("rate")
	a.ScheduleName = parser.String("scheduleName")
	a.Suppressed = parseSuppressed(parser.WithReferenceObjectParser("suppressed"))
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
	validator.String("scheduleName", a.ScheduleName).NotEmpty().LengthLessThanOrEqualTo(dataTypesBasal.ScheduleNameLengthMaximum)
	validateSuppressed(validator.WithReference("suppressed"), a.Suppressed)
}

func (a *Automated) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(a.Meta())
	}

	a.Basal.Normalize(normalizer)

	if a.InsulinFormulation != nil {
		a.InsulinFormulation.Normalize(normalizer.WithReference("insulinFormulation"))
	}
	if a.Suppressed != nil {
		a.Suppressed.Normalize(normalizer.WithReference("suppressed"))
	}
}

type SuppressedAutomated struct {
	Type         *string `json:"type,omitempty" bson:"type,omitempty"`
	DeliveryType *string `json:"deliveryType,omitempty" bson:"deliveryType,omitempty"`

	Annotations        *metadata.MetadataArray       `json:"annotations,omitempty" bson:"annotations,omitempty"`
	InsulinFormulation *dataTypesInsulin.Formulation `json:"insulinFormulation,omitempty" bson:"insulinFormulation,omitempty"`
	Rate               *float64                      `json:"rate,omitempty" bson:"rate,omitempty"`
	ScheduleName       *string                       `json:"scheduleName,omitempty" bson:"scheduleName,omitempty"`
	Suppressed         Suppressed                    `json:"suppressed,omitempty" bson:"suppressed,omitempty"`
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
		Type:         pointer.FromString(dataTypesBasal.Type),
		DeliveryType: pointer.FromString(DeliveryType),
	}
}

func (s *SuppressedAutomated) Parse(parser structure.ObjectParser) {
	s.Type = parser.String("type")
	s.DeliveryType = parser.String("deliveryType")

	s.Annotations = metadata.ParseMetadataArray(parser.WithReferenceArrayParser("annotations"))
	s.InsulinFormulation = dataTypesInsulin.ParseFormulation(parser.WithReferenceObjectParser("insulinFormulation"))
	s.Rate = parser.Float64("rate")
	s.ScheduleName = parser.String("scheduleName")
	s.Suppressed = parseSuppressed(parser.WithReferenceObjectParser("suppressed"))
}

func (s *SuppressedAutomated) Validate(validator structure.Validator) {
	validator.String("type", s.Type).Exists().EqualTo(dataTypesBasal.Type)
	validator.String("deliveryType", s.DeliveryType).Exists().EqualTo(DeliveryType)

	if s.Annotations != nil {
		s.Annotations.Validate(validator.WithReference("annotations"))
	}
	if s.InsulinFormulation != nil {
		s.InsulinFormulation.Validate(validator.WithReference("insulinFormulation"))
	}
	validator.Float64("rate", s.Rate).Exists().InRange(RateMinimum, RateMaximum)
	validator.String("scheduleName", s.ScheduleName).NotEmpty().LengthLessThanOrEqualTo(dataTypesBasal.ScheduleNameLengthMaximum)
	validateSuppressed(validator.WithReference("suppressed"), s.Suppressed)
}

func (s *SuppressedAutomated) Normalize(normalizer data.Normalizer) {
	if s.InsulinFormulation != nil {
		s.InsulinFormulation.Normalize(normalizer.WithReference("insulinFormulation"))
	}
	if s.Suppressed != nil {
		s.Suppressed.Normalize(normalizer.WithReference("suppressed"))
	}
}

var suppressedDeliveryTypes = []string{
	dataTypesBasalScheduled.DeliveryType,
}

func parseSuppressed(parser structure.ObjectParser) Suppressed {
	if deliveryType := dataTypesBasal.ParseDeliveryType(parser); deliveryType != nil {
		switch *deliveryType {
		case dataTypesBasalScheduled.DeliveryType:
			return dataTypesBasalScheduled.ParseSuppressedScheduled(parser)
		default:
			parser.WithReferenceErrorReporter("deliveryType").ReportError(structureValidator.ErrorValueStringNotOneOf(*deliveryType, suppressedDeliveryTypes))
		}
	}
	return nil
}

func validateSuppressed(validator structure.Validator, suppressed Suppressed) {
	if suppressed != nil {
		switch suppressed := suppressed.(type) {
		case *dataTypesBasalScheduled.SuppressedScheduled:
			suppressed.Validate(validator)
		default:
			validator.ReportError(structureValidator.ErrorValueNotValid())
		}
	}
}
