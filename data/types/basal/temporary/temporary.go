package temporary

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/basal"
	dataTypesBasalScheduled "github.com/tidepool-org/platform/data/types/basal/scheduled"
	"github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	DeliveryType = "temp" // TODO: Rename Type to "basal/temporary"; remove DeliveryType

	DurationMaximum = 604800000
	DurationMinimum = 0
	PercentMaximum  = 10.0
	PercentMinimum  = 0.0
	RateMaximum     = 100.0
	RateMinimum     = 0.0
)

type Suppressed interface {
	Parse(parser data.ObjectParser) error
	Validate(validator structure.Validator)
	Normalize(normalizer data.Normalizer)
}

type Temporary struct {
	basal.Basal `bson:",inline"`

	Duration         *int                 `json:"duration,omitempty" bson:"duration,omitempty"`
	DurationExpected *int                 `json:"expectedDuration,omitempty" bson:"expectedDuration,omitempty"`
	InsulinType      *insulin.InsulinType `json:"insulinType,omitempty" bson:"insulinType,omitempty"`
	Percent          *float64             `json:"percent,omitempty" bson:"percent,omitempty"`
	Rate             *float64             `json:"rate,omitempty" bson:"rate,omitempty"`
	Suppressed       Suppressed           `json:"suppressed,omitempty" bson:"suppressed,omitempty"`
}

func New() *Temporary {
	return &Temporary{
		Basal: basal.New(DeliveryType),
	}
}

func (t *Temporary) Parse(parser data.ObjectParser) error {
	if err := t.Basal.Parse(parser); err != nil {
		return err
	}

	t.Duration = parser.ParseInteger("duration")
	t.DurationExpected = parser.ParseInteger("expectedDuration")
	t.InsulinType = insulin.ParseInsulinType(parser.NewChildObjectParser("insulinType"))
	t.Percent = parser.ParseFloat("percent")
	t.Rate = parser.ParseFloat("rate")
	t.Suppressed = parseSuppressed(parser.NewChildObjectParser("suppressed"))

	return nil
}

func (t *Temporary) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(t.Meta())
	}

	t.Basal.Validate(validator)

	if t.DeliveryType != "" {
		validator.String("deliveryType", &t.DeliveryType).EqualTo(DeliveryType)
	}

	validator.Int("duration", t.Duration).Exists().InRange(DurationMinimum, DurationMaximum)
	expectedDurationValidator := validator.Int("expectedDuration", t.DurationExpected)
	if t.Duration != nil && *t.Duration >= DurationMinimum && *t.Duration <= DurationMaximum {
		expectedDurationValidator.InRange(*t.Duration, DurationMaximum)
	} else {
		expectedDurationValidator.InRange(DurationMinimum, DurationMaximum)
	}
	if t.InsulinType != nil {
		t.InsulinType.Validate(validator.WithReference("insulinType"))
	}
	validator.Float64("percent", t.Percent).InRange(PercentMinimum, PercentMaximum)
	validator.Float64("rate", t.Rate).Exists().InRange(RateMinimum, RateMaximum)
	validateSuppressed(validator.WithReference("suppressed"), t.Suppressed)
}

func (t *Temporary) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(t.Meta())
	}

	t.Basal.Normalize(normalizer)

	if t.InsulinType != nil {
		t.InsulinType.Normalize(normalizer.WithReference("insulinType"))
	}
	if t.Suppressed != nil {
		t.Suppressed.Normalize(normalizer.WithReference("suppressed"))
	}
}

type SuppressedTemporary struct {
	Type         *string `json:"type,omitempty" bson:"type,omitempty"`
	DeliveryType *string `json:"deliveryType,omitempty" bson:"deliveryType,omitempty"`

	Annotations *data.BlobArray      `json:"annotations,omitempty" bson:"annotations,omitempty"`
	InsulinType *insulin.InsulinType `json:"insulinType,omitempty" bson:"insulinType,omitempty"`
	Percent     *float64             `json:"percent,omitempty" bson:"percent,omitempty"`
	Rate        *float64             `json:"rate,omitempty" bson:"rate,omitempty"`
	Suppressed  Suppressed           `json:"suppressed,omitempty" bson:"suppressed,omitempty"`
}

func ParseSuppressedTemporary(parser data.ObjectParser) *SuppressedTemporary {
	if parser.Object() == nil {
		return nil
	}
	suppressed := NewSuppressedTemporary()
	suppressed.Parse(parser)
	parser.ProcessNotParsed()
	return suppressed
}

func NewSuppressedTemporary() *SuppressedTemporary {
	return &SuppressedTemporary{
		Type:         pointer.String(basal.Type),
		DeliveryType: pointer.String(DeliveryType),
	}
}

func (s *SuppressedTemporary) Parse(parser data.ObjectParser) error {
	s.Type = parser.ParseString("type")
	s.DeliveryType = parser.ParseString("deliveryType")

	s.Annotations = data.ParseBlobArray(parser.NewChildArrayParser("annotations"))
	s.InsulinType = insulin.ParseInsulinType(parser.NewChildObjectParser("insulinType"))
	s.Percent = parser.ParseFloat("percent")
	s.Rate = parser.ParseFloat("rate")
	s.Suppressed = parseSuppressed(parser.NewChildObjectParser("suppressed"))

	return nil
}

func (s *SuppressedTemporary) Validate(validator structure.Validator) {
	validator.String("type", s.Type).Exists().EqualTo(basal.Type)
	validator.String("deliveryType", s.DeliveryType).Exists().EqualTo(DeliveryType)

	if s.Annotations != nil {
		s.Annotations.Validate(validator.WithReference("annotations"))
	}
	if s.InsulinType != nil {
		s.InsulinType.Validate(validator.WithReference("insulinType"))
	}
	validator.Float64("percent", s.Percent).InRange(PercentMinimum, PercentMaximum)
	validator.Float64("rate", s.Rate).Exists().InRange(RateMinimum, RateMaximum)
	validateSuppressed(validator.WithReference("suppressed"), s.Suppressed)
}

func (s *SuppressedTemporary) Normalize(normalizer data.Normalizer) {
	if s.Annotations != nil {
		s.Annotations.Normalize(normalizer.WithReference("annotations"))
	}
	if s.InsulinType != nil {
		s.InsulinType.Normalize(normalizer.WithReference("insulinType"))
	}
	if s.Suppressed != nil {
		s.Suppressed.Normalize(normalizer.WithReference("suppressed"))
	}
}

var suppressedDeliveryTypes = []string{
	dataTypesBasalScheduled.DeliveryType,
}

func parseSuppressed(parser data.ObjectParser) Suppressed {
	if deliveryType := basal.ParseDeliveryType(parser); deliveryType != nil {
		switch *deliveryType {
		case dataTypesBasalScheduled.DeliveryType:
			return dataTypesBasalScheduled.ParseSuppressedScheduled(parser)
		default:
			parser.AppendError("type", service.ErrorValueStringNotOneOf(*deliveryType, suppressedDeliveryTypes))
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
			validator.ReportError(structureValidator.ErrorValueExists()) // TODO: Better error?
		}
	}
}
