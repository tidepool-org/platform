package mode

import (
	"github.com/tidepool-org/platform/data"
	commontypes "github.com/tidepool-org/platform/data/types/common"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/structure"
)

const (
	ConfidentialMode = "confidential"
	ZenMode          = "zen"
)

func Modes() []string {
	return []string{
		ConfidentialMode,
		ZenMode,
	}
}

type Mode struct {
	device.Device `bson:",inline"`
	EventID       *string               `json:"eventId,omitempty" bson:"eventId,omitempty"`
	Duration      *commontypes.Duration `json:"duration,omitempty" bson:"duration,omitempty"`
}

func New(subType string) *Mode {
	return &Mode{
		Device: device.New(subType),
	}
}

func NewWithEvent(subType string, deviceEvent string) *Mode {
	return &Mode{
		Device: device.NewWithEvent(subType, deviceEvent),
	}
}

func (m *Mode) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(m.Meta())
	}

	m.Device.Parse(parser)
	m.EventID = parser.String("eventId")
	m.Duration = commontypes.ParseDuration(parser.WithReferenceObjectParser("duration"))
}

func (m *Mode) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(m.Meta())
	}

	m.Device.Validate(validator)

	if m.SubType != "" {
		validator.String("subType", &m.SubType).OneOf(Modes()...)
	}
	validator.String("eventId", m.EventID).Exists().NotEmpty()
	if m.Duration != nil {
		m.Duration.Validate(validator.WithReference("duration"))
	}
}

// IsValid returns true if there is no error in the validator
func (m *Mode) IsValid(validator structure.Validator) bool {
	return !(validator.HasError())
}

func (m *Mode) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(m.Meta())
	}
	m.Device.Normalize(normalizer)
	if m.Duration != nil {
		m.Duration.Normalize(normalizer.WithReference("duration"))
	}
}
