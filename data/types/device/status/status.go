package status

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/device"
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

func NewStatusDatum(parser structure.ObjectParser) data.Datum {
	if !parser.Exists() {
		return nil
	}

	if value := parser.String("type"); value == nil {
		parser.WithReferenceErrorReporter("type").ReportError(structureValidator.ErrorValueNotExists())
		return nil
	} else if *value != device.Type {
		parser.WithReferenceErrorReporter("type").ReportError(structureValidator.ErrorValueStringNotOneOf(*value, []string{device.Type}))
		return nil
	}

	if value := parser.String("subType"); value == nil {
		parser.WithReferenceErrorReporter("subType").ReportError(structureValidator.ErrorValueNotExists())
		return nil
	} else if *value != SubType {
		parser.WithReferenceErrorReporter("subType").ReportError(structureValidator.ErrorValueStringNotOneOf(*value, []string{SubType}))
		return nil
	}

	return New()
}

func ParseStatusDatum(parser structure.ObjectParser) *data.Datum {
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

func (s *Status) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(s.Meta())
	}

	s.Device.Parse(parser)

	s.Duration = parser.Int("duration")
	s.DurationExpected = parser.Int("expectedDuration")
	s.Name = parser.String("status")
	s.Reason = data.ParseBlob(parser.WithReferenceObjectParser("reason"))
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
