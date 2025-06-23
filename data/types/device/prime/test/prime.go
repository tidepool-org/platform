package test

import (
	dataTypesDevicePrime "github.com/tidepool-org/platform/data/types/device/prime"
	dataTypesDeviceTest "github.com/tidepool-org/platform/data/types/device/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomPrime() *dataTypesDevicePrime.Prime {
	datum := dataTypesDevicePrime.New()
	datum.Device = *dataTypesDeviceTest.RandomDevice()
	datum.SubType = dataTypesDevicePrime.SubType
	datum.Target = pointer.FromString(test.RandomStringFromArray(dataTypesDevicePrime.Targets()))
	switch *datum.Target {
	case dataTypesDevicePrime.TargetCannula:
		datum.Volume = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDevicePrime.VolumeTargetCannulaMinimum, dataTypesDevicePrime.VolumeTargetCannulaMaximum))
	case dataTypesDevicePrime.TargetTubing:
		datum.Volume = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDevicePrime.VolumeTargetTubingMinimum, dataTypesDevicePrime.VolumeTargetTubingMaximum))
	}
	return datum
}

func ClonePrime(datum *dataTypesDevicePrime.Prime) *dataTypesDevicePrime.Prime {
	if datum == nil {
		return nil
	}
	clone := dataTypesDevicePrime.New()
	clone.Device = *dataTypesDeviceTest.CloneDevice(&datum.Device)
	clone.Target = pointer.CloneString(datum.Target)
	clone.Volume = pointer.CloneFloat64(datum.Volume)
	return clone
}
