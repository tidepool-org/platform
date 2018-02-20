package suspend

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/data/types/basal/scheduled"
	"github.com/tidepool-org/platform/data/types/basal/temporary"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
)

const (
	DurationMaximum = 604800000
	DurationMinimum = 0
)

func SuppressedDeliveryTypes() []string {
	return []string{
		scheduled.DeliveryType(),
		temporary.DeliveryType(),
	}
}

type Suspend struct {
	basal.Basal `bson:",inline"`

	Duration         *int              `json:"duration,omitempty" bson:"duration,omitempty"`
	DurationExpected *int              `json:"expectedDuration,omitempty" bson:"expectedDuration,omitempty"`
	Suppressed       *basal.Suppressed `json:"suppressed,omitempty" bson:"suppressed,omitempty"`
}

func DeliveryType() string {
	return "suspend" // TODO: Rename Type to "basal/suspended"; remove DeliveryType
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
	s.DurationExpected = nil
	s.Suppressed = nil
}

func (s *Suspend) Parse(parser data.ObjectParser) error {
	if err := s.Basal.Parse(parser); err != nil {
		return err
	}

	s.Duration = parser.ParseInteger("duration")
	s.DurationExpected = parser.ParseInteger("expectedDuration")
	s.Suppressed = basal.ParseSuppressed(parser.NewChildObjectParser("suppressed"))

	return nil
}

func (s *Suspend) Validate(validator structure.Validator) {
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
	if s.Suppressed != nil {
		s.Suppressed.Validate(validator.WithReference("suppressed"), pointer.StringArray(SuppressedDeliveryTypes()))
	}
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
