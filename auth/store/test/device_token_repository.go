package test

import (
	"context"

	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/devicetokens"
)

type DeviceTokenRepository struct {
	*authTest.DeviceTokenAccessor
}

func NewDeviceTokenRepository() *DeviceTokenRepository {
	return &DeviceTokenRepository{
		DeviceTokenAccessor: authTest.NewDeviceTokenAccessor(),
	}
}

func (r *DeviceTokenRepository) Expectations() {
	r.DeviceTokenAccessor.Expectations()
}

func (r *DeviceTokenRepository) Upsert(ctx context.Context, doc *devicetokens.Document) error {
	return nil
}

func (r *DeviceTokenRepository) EnsureIndexes() error {
	return nil
}
