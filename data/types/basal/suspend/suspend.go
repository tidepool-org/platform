package suspend

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/basal"
	dataTypesBasalAutomated "github.com/tidepool-org/platform/data/types/basal/automated"
	dataTypesBasalScheduled "github.com/tidepool-org/platform/data/types/basal/scheduled"
	dataTypesBasalTemporary "github.com/tidepool-org/platform/data/types/basal/temporary"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	DeliveryType = "suspend" // TODO: Rename Type to "basal/suspended"; remove DeliveryType

	DurationMaximum = 604800000
	DurationMinimum = 0
)

type Suppressed interface {
	Parse(parser structure.ObjectParser)
	Validate(validator structure.Validator)
	Normalize(normalizer data.Normalizer)
}

type Suspend struct {
	basal.Basal `bson:",inline"`

	Duration         *int       `json:"duration,omitempty" bson:"duration,omitempty"`
	DurationExpected *int       `json:"expectedDuration,omitempty" bson:"expectedDuration,omitempty"`
	Suppressed       Suppressed `json:"suppressed,omitempty" bson:"suppressed,omitempty"`
}

func New() *Suspend {
	return &Suspend{
		Basal: basal.New(DeliveryType),
	}
}

func (s *Suspend) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(s.Meta())
	}

	s.Basal.Parse(parser)

	s.Duration = parser.Int("duration")
	s.DurationExpected = parser.Int("expectedDuration")
	s.Suppressed = parseSuppressed(parser.WithReferenceObjectParser("suppressed"))
}

func (s *Suspend) Validate(validator structure.Validator) {
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
	validateSuppressed(validator.WithReference("suppressed"), s.Suppressed)
}

// IsValid returns true if there is no error in the validator
func (s *Suspend) IsValid(validator structure.Validator) bool {
	return !(validator.HasError())
}

func (s *Suspend) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(s.Meta())
	}

	s.Basal.Normalize(normalizer)

	if s.Suppressed != nil {
		s.Suppressed.Normalize(normalizer.WithReference("suppressed"))
	}
}

var suppressedDeliveryTypes = []string{
	dataTypesBasalAutomated.DeliveryType,
	dataTypesBasalScheduled.DeliveryType,
	dataTypesBasalTemporary.DeliveryType,
}

func parseSuppressed(parser structure.ObjectParser) Suppressed {
	if deliveryType := basal.ParseDeliveryType(parser); deliveryType != nil {
		switch *deliveryType {
		case dataTypesBasalAutomated.DeliveryType:
			return dataTypesBasalAutomated.ParseSuppressedAutomated(parser)
		case dataTypesBasalScheduled.DeliveryType:
			return dataTypesBasalScheduled.ParseSuppressedScheduled(parser)
		case dataTypesBasalTemporary.DeliveryType:
			return dataTypesBasalTemporary.ParseSuppressedTemporary(parser)
		default:
			parser.WithReferenceErrorReporter("type").ReportError(structureValidator.ErrorValueStringNotOneOf(*deliveryType, suppressedDeliveryTypes))
		}
	}
	return nil
}

func validateSuppressed(validator structure.Validator, suppressed Suppressed) {
	if suppressed != nil {
		switch suppressed := suppressed.(type) {
		case *dataTypesBasalAutomated.SuppressedAutomated:
			suppressed.Validate(validator)
		case *dataTypesBasalScheduled.SuppressedScheduled:
			suppressed.Validate(validator)
		case *dataTypesBasalTemporary.SuppressedTemporary:
			suppressed.Validate(validator)
		default:
			validator.ReportError(structureValidator.ErrorValueExists()) // TODO: Better error?
		}
	}
}
