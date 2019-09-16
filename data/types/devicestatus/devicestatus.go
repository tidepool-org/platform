package devicestatus

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/devicestatus/status"
	"github.com/tidepool-org/platform/structure"
)

const (
	Type     = "deviceStatus"
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

	DeviceType *string                 `json:"deviceType,omitempty" bson:"deviceType,omitempty"`
	Version    *string                 `json:"version,omitempty" bson:"version,omitempty"`
	Status     *status.TypeStatusArray `json:"status,omitempty" bson:"status,omitempty"`
}

func New() *DeviceStatus {
	return &DeviceStatus{
		Base: types.New(Type),
	}
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
	if !parser.HasMeta() {
		parser = parser.WithMeta(a.Meta())
	}

	a.Base.Parse(parser)

	a.DeviceType = parser.String("deviceType")
	a.Version = parser.String("version")
	a.Status = status.ParseStatusArray(parser.WithReferenceArrayParser("status"))
}

func (a *DeviceStatus) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(a.Meta())
	}

	a.Base.Validate(validator)

	if a.Type != "" {
		validator.String("type", &a.Type).EqualTo(Type)
	}

	validator.String("deviceType", a.DeviceType).Exists().OneOf(DeviceTypes()...)
}

func (a *DeviceStatus) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(a.Meta())
	}

	a.Base.Normalize(normalizer)
}
