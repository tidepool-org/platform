package test

import (
	"context"
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/page"
)

type DeviceAuthorizationAccessor struct {
}

func NewDeviceAuthorizationAccessor() *DeviceAuthorizationAccessor {
	return &DeviceAuthorizationAccessor{}
}

func (d *DeviceAuthorizationAccessor) GetUserDeviceAuthorization(ctx context.Context, userID string, id string) (*auth.DeviceAuthorization, error) {
	panic("implement me")
}

func (d *DeviceAuthorizationAccessor) ListUserDeviceAuthorizations(ctx context.Context, userID string, pagination *page.Pagination) (auth.DeviceAuthorizations, error) {
	panic("implement me")
}

func (d *DeviceAuthorizationAccessor) GetDeviceAuthorizationByToken(ctx context.Context, token string) (*auth.DeviceAuthorization, error) {
	panic("implement me")
}

func (d *DeviceAuthorizationAccessor) CreateUserDeviceAuthorization(ctx context.Context, userID string, create *auth.DeviceAuthorizationCreate) (*auth.DeviceAuthorization, error) {
	panic("implement me")
}

func (d *DeviceAuthorizationAccessor) UpdateDeviceAuthorization(ctx context.Context, id string, update *auth.DeviceAuthorizationUpdate) (*auth.DeviceAuthorization, error) {
	panic("implement me")
}

func (p *DeviceAuthorizationAccessor) Expectations() {
}