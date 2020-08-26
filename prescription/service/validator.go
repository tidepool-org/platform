package service

import (
	"context"

	devicesApi "github.com/tidepool-org/devices/api"

	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/devices"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/prescription"
	"github.com/tidepool-org/platform/structure"
)

type DeviceSettingsValidator interface {
	Validate(ctx context.Context, settings *prescription.InitialSettings, validator structure.Validator) error
}

type deviceSettingsValidator struct {
	devicesClient devicesApi.DevicesClient
}

func NewDeviceSettingsValidator(client devicesApi.DevicesClient) DeviceSettingsValidator {
	return &deviceSettingsValidator{
		devicesClient: client,
	}
}

func (d *deviceSettingsValidator) Validate(ctx context.Context, settings *prescription.InitialSettings, validator structure.Validator) error {
	if settings == nil {
		return nil
	}

	if settings.CgmID != "" {
		// Make sure the referenced CGM exists
		_, err := d.devicesClient.GetCgmById(ctx, &devicesApi.GetCgmByIdRequest{Id: settings.CgmID})
		if err != nil {
			return err
		}
	}
	// Only verify the settings if a pump has been selected.
	if settings.PumpID == "" {
		return nil
	}

	response, err := d.devicesClient.GetPumpById(ctx, &devicesApi.GetPumpByIdRequest{Id: settings.PumpID})
	if err != nil {
		return err
	}

	guardRails := response.GetPump().GetGuardRails()
	if !canValidatePrescriptionSettings(settings, guardRails) {
		return errors.New("cannot validate device specific prescription settings increments and limits")
	}

	if settings.BasalRateSchedule != nil {
		devices.ValidateBasalRateSchedule(*settings.BasalRateSchedule, guardRails.GetBasalRates(), validator.WithReference("basalRateSchedule"))
	}
	if settings.BloodGlucoseSuspendThreshold != nil {
		devices.ValidateBloodGlucoseSuspendThreshold(settings.BloodGlucoseSuspendThreshold, guardRails.GetSuspendThreshold(), validator.WithReference("bloodGlucoseSuspendThreshold"))
	}
	if settings.BloodGlucoseTargetSchedule != nil {
		devices.ValidateBloodGlucoseTargetSchedule(*settings.BloodGlucoseTargetSchedule, guardRails.GetCorrectionRange(), validator.WithReference("bloodGlucoseTargetSchedule"))
	}
	if settings.CarbohydrateRatioSchedule != nil {
		devices.ValidateCarbohydrateRatioSchedule(*settings.CarbohydrateRatioSchedule, guardRails.GetCarbohydrateRatio(), validator.WithReference("carbohydrateRatio"))
	}
	if settings.InsulinSensitivitySchedule != nil {
		devices.ValidateInsulinSensitivitySchedule(*settings.InsulinSensitivitySchedule, guardRails.GetInsulinSensitivity(), validator.WithReference("insulinSensitivitySchedule"))
	}
	if settings.BasalRateMaximum != nil {
		devices.ValidateBasalRateMaximum(*settings.BasalRateMaximum, guardRails.GetBasalRateMaximum(), validator.WithReference("basalRateMaximum"))
	}
	if settings.BolusAmountMaximum != nil {
		devices.ValidateBolusAmountMaximum(*settings.BolusAmountMaximum, guardRails.GetBolusAmountMaximum(), validator.WithReference("bolusAmountMaximum"))
	}

	return nil
}

func canValidatePrescriptionSettings(settings *prescription.InitialSettings, guardRails *devicesApi.GuardRails) bool {
	if settings == nil || guardRails == nil {
		return false
	}
	bgUnitsInSettings := settings.BloodGlucoseUnits
	if bgUnitsInSettings != glucose.Mgdl && bgUnitsInSettings != glucose.MgdL {
		return false
	}
	if guardRails.GetCorrectionRange().Units != devicesApi.BloodGlucoseUnits_MilligramsPerDeciliter {
		return false
	}
	if guardRails.GetSuspendThreshold().Units != devicesApi.BloodGlucoseUnits_MilligramsPerDeciliter {
		return false
	}
	if guardRails.GetInsulinSensitivity().Units != devicesApi.BloodGlucoseUnits_MilligramsPerDeciliter {
		return false
	}
	return true
}
