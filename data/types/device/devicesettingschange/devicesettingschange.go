package devicesettingschange

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/structure"
)

const (
	SubType = "deviceSettingsChange" // TODO: Rename Type to "device/reservoirChange"; remove SubType
)

type DeviceSettingsChange struct {
	device.Device `bson:",inline"`

	BasalSchedule            *SettingsChange `bson:",inline"`
	BgTarget                 *SettingsChange `bson:",inline"`
	CarbRatio                *SettingsChange `bson:",inline"`
	InsulinSensitivityFactor *SettingsChange `bson:",inline"`
}

func New() *DeviceSettingsChange {
	return &DeviceSettingsChange{
		Device: device.New(SubType),
	}
}

func (r *DeviceSettingsChange) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(r.Meta())
	}

	r.Device.Parse(parser)
	r.BasalSchedule.Parse(parser)
	r.BgTarget.Parse(parser)
	r.CarbRatio.Parse(parser)
	r.InsulinSensitivityFactor.Parse(parser)
}

func (r *DeviceSettingsChange) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(r.Meta())
	}

	r.Device.Validate(validator)
	r.BasalSchedule.Validate(validator)
	r.BgTarget.Validate(validator)
	r.CarbRatio.Validate(validator)
	r.InsulinSensitivityFactor.Validate(validator)

	if r.SubType != "" {
		validator.String("subType", &r.SubType).EqualTo(SubType)
	}

}

func (r *DeviceSettingsChange) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(r.Meta())
	}

	r.Device.Normalize(normalizer)
	r.BasalSchedule.Normalize(normalizer)
	r.BgTarget.Normalize(normalizer)
	r.CarbRatio.Normalize(normalizer)
	r.InsulinSensitivityFactor.Normalize(normalizer)

	if normalizer.Origin() == structure.OriginExternal {
	}
}
