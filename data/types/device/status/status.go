package status

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	SubType = "status" // TODO: Rename Type to "device/status"; remove SubType; consider device/resumed + device/suspended

	DurationMinimum = 0
	NameResumed     = "resumed"
	NameSuspended   = "suspended"
)

func Names() []string {
	return []string{
		NameResumed,
		NameSuspended,
	}
}

type Status struct {
	device.Device `bson:",inline"`

	Duration         *int       `json:"duration,omitempty" bson:"duration,omitempty"`
	DurationExpected *int       `json:"expectedDuration,omitempty" bson:"expectedDuration,omitempty"`
	Name             *string    `json:"status,omitempty" bson:"status,omitempty"`
	Reason           *data.Blob `json:"reason,omitempty" bson:"reason,omitempty"`
}

func NewStatusDatum(parser data.ObjectParser) data.Datum {
	if parser.Object() == nil {
		return nil
	}

	if value := parser.ParseString("type"); value == nil {
		parser.AppendError("type", service.ErrorValueNotExists())
		return nil
	} else if *value != device.Type {
		parser.AppendError("type", service.ErrorValueStringNotOneOf(*value, []string{device.Type}))
		return nil
	}

	if value := parser.ParseString("subType"); value == nil {
		parser.AppendError("subType", service.ErrorValueNotExists())
		return nil
	} else if *value != SubType {
		parser.AppendError("subType", service.ErrorValueStringNotOneOf(*value, []string{SubType}))
		return nil
	}

	return New()
}

func ParseStatusDatum(parser data.ObjectParser) *data.Datum {
	datum := NewStatusDatum(parser)
	if datum == nil {
		return nil
	}

	datum.Parse(parser)
	return &datum
}

func New() *Status {
	return &Status{
		Device: device.New(SubType),
	}
}

func (s *Status) Parse(parser data.ObjectParser) error {
	if err := s.Device.Parse(parser); err != nil {
		return err
	}

	s.Duration = parser.ParseInteger("duration")
	s.DurationExpected = parser.ParseInteger("expectedDuration")
	s.Name = parser.ParseString("status")
	s.Reason = data.ParseBlob(parser.NewChildObjectParser("reason"))

	return nil
}

func (s *Status) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(s.Meta())
	}

	s.Device.Validate(validator)

	if s.SubType != "" {
		validator.String("subType", &s.SubType).EqualTo(SubType)
	}

	validator.Int("duration", s.Duration).GreaterThanOrEqualTo(DurationMinimum) // TODO: .Exists() - Suspend events on Animas do not have duration?
	expectedDurationValidator := validator.Int("expectedDuration", s.DurationExpected)
	if s.Duration != nil && *s.Duration >= DurationMinimum {
		expectedDurationValidator.GreaterThanOrEqualTo(*s.Duration)
	} else {
		expectedDurationValidator.GreaterThanOrEqualTo(DurationMinimum)
	}
	validator.String("status", s.Name).Exists().OneOf(Names()...)

	reasonValidator := validator.WithReference("reason")
	if s.Reason != nil {
		s.Reason.Validate(reasonValidator)
	} else {
		reasonValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
}

func (s *Status) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(s.Meta())
	}

	s.Device.Normalize(normalizer)

	if s.Reason != nil {
		s.Reason.Normalize(normalizer.WithReference("reason"))
	}
}
