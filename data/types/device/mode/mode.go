package mode

import (
	"github.com/tidepool-org/platform/data"
	commontypes "github.com/tidepool-org/platform/data/types/common"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	ConfidentialMode = "confidential"
	ZenMode          = "zen"
	Warmup           = "warmup"
	LoopMode         = "loopMode"
	EnergySaving     = "energySaving"
)

func Modes() []string {
	return []string{
		ConfidentialMode,
		ZenMode,
		Warmup,
		LoopMode,
		EnergySaving,
	}
}

type Mode struct {
	device.Device `bson:",inline"`
	Duration      *commontypes.Duration  `json:"duration,omitempty" bson:"duration,omitempty"`
	InputTime     *commontypes.InputTime `bson:",inline"`
}

func New(subType string) *Mode {
	return &Mode{
		Device:    device.New(subType),
		InputTime: commontypes.NewInputTime(),
	}
}

func (m *Mode) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(m.Meta())
	}
	m.Device.Parse(parser)
	m.Duration = commontypes.ParseDuration(parser.WithReferenceObjectParser("duration"))
	m.InputTime.Parse(parser)
}

func (m *Mode) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(m.Meta())
	}

	m.Device.Validate(validator)
	validator.String("subType", &m.SubType).OneOf(Modes()...)
	validator.String("guid", m.GUID).Exists().NotEmpty()
	if m.Duration != nil {
		m.Duration.Validate(validator.WithReference("duration"))
	} else if m.SubType != LoopMode {
		validator.WithReference("duration").ReportError(structureValidator.ErrorValueNotExists())
	}
	if m.InputTime.InputTime != nil {
		m.InputTime.Validate(validator)
	} else {
		validator.WithReference("inputTime").ReportError(structureValidator.ErrorValueNotExists())
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
	m.InputTime.Normalize(normalizer)
}
