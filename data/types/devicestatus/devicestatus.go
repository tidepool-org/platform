package devicestatus

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/devicestatus/status"
	"github.com/tidepool-org/platform/structure"
)

const (
	Aid      = "aid"
	Cgm      = "cgm"
	Pump     = "pump"
	SmartPen = "smartpen"
)

func DeviceTypes() []string {
	return []string{Aid, Cgm, Pump, SmartPen}
}

type DeviceStatus struct {
	types.Base `bson:",inline"`

	DeviceType *string      `json:"deviceType,omitempty" bson:"deviceType,omitempty"`
	Version    *string      `json:"version,omitempty" bson:"version,omitempty"`
	Status     status.Array `json:"status,omitempty" bson:"status,omitempty"`
}

func ParseDeviceStatus(parser structure.ObjectParser) *DeviceStatus {
	if !parser.Exists() {
		return nil
	}
	datum := NewDeviceStatus()
	parser.Parse(datum)
	return datum
}

func NewDeviceStatus() *DeviceStatus {
	return &DeviceStatus{}
}

func (a *DeviceStatus) Parse(parser structure.ObjectParser) {
	a.DeviceType = parser.String("deviceType")
	a.Version = parser.String("version")
}

func (a *DeviceStatus) Validate(validator structure.Validator) {
	validator.String("deviceType", a.DeviceType).Exists().OneOf(DeviceTypes()...)
}

func (a *DeviceStatus) Normalize(normalizer data.Normalizer) {}
