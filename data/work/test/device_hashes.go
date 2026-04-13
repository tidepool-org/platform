package test

import (
	"maps"

	dataWork "github.com/tidepool-org/platform/data/work"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomDeviceID() string {
	return test.RandomStringFromRange(1, dataWork.DeviceIDLengthMaximum)
}

func RandomDeviceHash() string {
	return test.RandomStringFromRange(1, dataWork.DeviceHashLengthMaximum)
}

func RandomDeviceHashesMap() map[string]string {
	datum := map[string]string{}
	for range test.RandomIntFromRange(1, 3) {
		datum[RandomDeviceID()] = RandomDeviceHash()
	}
	return datum
}

func CloneDeviceHashesMap(datum map[string]string) map[string]string {
	return maps.Clone(datum)
}

func NewObjectFromDeviceHashesMap(datum map[string]string, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	for deviceID, deviceHash := range datum {
		object[deviceID] = test.NewObjectFromString(deviceHash, objectFormat)
	}
	return object
}

func RandomDeviceHashes() *dataWork.DeviceHashes {
	return pointer.From(dataWork.DeviceHashes(RandomDeviceHashesMap()))
}

func CloneDeviceHashes(datum *dataWork.DeviceHashes) *dataWork.DeviceHashes {
	if datum == nil {
		return nil
	}
	return pointer.From(dataWork.DeviceHashes(CloneDeviceHashesMap(*datum)))
}

func NewObjectFromDeviceHashes(datum *dataWork.DeviceHashes, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	return NewObjectFromDeviceHashesMap(*datum, objectFormat)
}

func RandomDeviceHashesMetadata(options ...test.Option) *dataWork.DeviceHashesMetadata {
	return &dataWork.DeviceHashesMetadata{
		DeviceHashes: test.RandomOptionalPointer(RandomDeviceHashes, options...),
	}
}

func CloneDeviceHashesMetadata(datum *dataWork.DeviceHashesMetadata) *dataWork.DeviceHashesMetadata {
	if datum == nil {
		return nil
	}
	return &dataWork.DeviceHashesMetadata{
		DeviceHashes: CloneDeviceHashes(datum.DeviceHashes),
	}
}

func NewObjectFromDeviceHashesMetadata(datum *dataWork.DeviceHashesMetadata, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.DeviceHashes != nil {
		object[dataWork.MetadataKeyDeviceHashes] = NewObjectFromDeviceHashes(datum.DeviceHashes, objectFormat)
	}
	return object
}
