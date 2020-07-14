package test

import (
	"context"

	"github.com/tidepool-org/platform/prescription"
	"github.com/tidepool-org/platform/prescription/service"
	"github.com/tidepool-org/platform/structure"
)

type NoopSettingsValidator struct{}

func NewNoopSettingsValidator() service.DeviceSettingsValidator {
	return &NoopSettingsValidator{}
}

func (n NoopSettingsValidator) Validate(ctx context.Context, settings *prescription.InitialSettings, validator structure.Validator) error {
	return nil
}
