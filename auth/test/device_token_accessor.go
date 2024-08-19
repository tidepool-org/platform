package test

type DeviceTokenAccessor struct{}

func NewDeviceTokenAccessor() *DeviceTokenAccessor {
	return &DeviceTokenAccessor{}
}

func (a *DeviceTokenAccessor) Expectations() {}
