package test

import (
	"math/rand"

	dataTypesSettingsController "github.com/tidepool-org/platform/data/types/settings/controller"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomDevice() *dataTypesSettingsController.Device {
	datum := dataTypesSettingsController.NewDevice()
	datum.FirmwareVersion = pointer.FromString(test.RandomStringFromRange(1, dataTypesSettingsController.FirmwareVersionLengthMaximum))
	datum.HardwareVersion = pointer.FromString(test.RandomStringFromRange(1, dataTypesSettingsController.HardwareVersionLengthMaximum))
	datum.Manufacturers = pointer.FromStringArray(RandomManufacturersFromRange(1, dataTypesSettingsController.ManufacturersLengthMaximum))
	datum.Model = pointer.FromString(test.RandomStringFromRange(1, dataTypesSettingsController.ModelLengthMaximum))
	datum.Name = pointer.FromString(test.RandomStringFromRange(1, dataTypesSettingsController.NameLengthMaximum))
	datum.SerialNumber = pointer.FromString(test.RandomStringFromRange(1, dataTypesSettingsController.SerialNumberLengthMaximum))
	datum.SoftwareVersion = pointer.FromString(test.RandomStringFromRange(1, dataTypesSettingsController.SoftwareVersionLengthMaximum))
	return datum
}

func RandomManufacturersFromRange(minimumLength int, maximumLength int) []string {
	result := make([]string, minimumLength+rand.Intn(maximumLength-minimumLength+1))
	for index := range result {
		result[index] = RandomManufacturerFromRange(1, dataTypesSettingsController.ManufacturerLengthMaximum)
	}
	return result
}

func RandomManufacturerFromRange(minimumLength int, maximumLength int) string {
	return test.RandomStringFromRange(minimumLength, maximumLength)
}

func CloneDevice(datum *dataTypesSettingsController.Device) *dataTypesSettingsController.Device {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsController.NewDevice()
	clone.FirmwareVersion = pointer.CloneString(datum.FirmwareVersion)
	clone.HardwareVersion = pointer.CloneString(datum.HardwareVersion)
	clone.Manufacturers = pointer.CloneStringArray(datum.Manufacturers)
	clone.Model = pointer.CloneString(datum.Model)
	clone.Name = pointer.CloneString(datum.Name)
	clone.SerialNumber = pointer.CloneString(datum.SerialNumber)
	clone.SoftwareVersion = pointer.CloneString(datum.SoftwareVersion)
	return clone
}

func NewObjectFromDevice(datum *dataTypesSettingsController.Device, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.FirmwareVersion != nil {
		object["firmwareVersion"] = test.NewObjectFromString(*datum.FirmwareVersion, objectFormat)
	}
	if datum.HardwareVersion != nil {
		object["hardwareVersion"] = test.NewObjectFromString(*datum.HardwareVersion, objectFormat)
	}
	if datum.Manufacturers != nil {
		object["manufacturers"] = test.NewObjectFromStringArray(*datum.Manufacturers, objectFormat)
	}
	if datum.Model != nil {
		object["model"] = test.NewObjectFromString(*datum.Model, objectFormat)
	}
	if datum.Name != nil {
		object["name"] = test.NewObjectFromString(*datum.Name, objectFormat)
	}
	if datum.SerialNumber != nil {
		object["serialNumber"] = test.NewObjectFromString(*datum.SerialNumber, objectFormat)
	}
	if datum.SoftwareVersion != nil {
		object["softwareVersion"] = test.NewObjectFromString(*datum.SoftwareVersion, objectFormat)
	}
	return object
}
