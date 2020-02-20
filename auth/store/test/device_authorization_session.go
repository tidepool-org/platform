package test

import (
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/test"
)

type DeviceAuthorizationSession struct {
	*test.Closer
	*authTest.DeviceAuthorizationAccessor
}

func NewDeviceAuthorizationSession() *DeviceAuthorizationSession {
	return &DeviceAuthorizationSession{
		Closer:                      test.NewCloser(),
		DeviceAuthorizationAccessor: authTest.NewDeviceAuthorizationAccessor(),
	}
}

func (p *DeviceAuthorizationSession) Expectations() {
	p.Closer.AssertOutputsEmpty()
	p.DeviceAuthorizationAccessor.Expectations()
}
