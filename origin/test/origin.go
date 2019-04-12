package test

import (
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	"github.com/tidepool-org/platform/origin"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomOrigin() *origin.Origin {
	datum := origin.NewOrigin()
	datum.ID = pointer.FromString(RandomID())
	datum.Name = pointer.FromString(RandomName())
	datum.Payload = metadataTest.RandomMetadata()
	datum.Time = pointer.FromString(RandomTime())
	datum.Type = pointer.FromString(RandomType())
	datum.Version = pointer.FromString(RandomVersion())
	return datum
}

func CloneOrigin(datum *origin.Origin) *origin.Origin {
	if datum == nil {
		return nil
	}
	clone := origin.NewOrigin()
	clone.ID = pointer.CloneString(datum.ID)
	clone.Name = pointer.CloneString(datum.Name)
	clone.Payload = metadataTest.CloneMetadata(datum.Payload)
	clone.Time = pointer.CloneString(datum.Time)
	clone.Type = pointer.CloneString(datum.Type)
	clone.Version = pointer.CloneString(datum.Version)
	return clone
}

func NewObjectFromOrigin(datum *origin.Origin, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.ID != nil {
		object["id"] = test.NewObjectFromString(*datum.ID, objectFormat)
	}
	if datum.Name != nil {
		object["name"] = test.NewObjectFromString(*datum.Name, objectFormat)
	}
	if datum.Payload != nil {
		object["payload"] = metadataTest.NewObjectFromMetadata(datum.Payload, objectFormat)
	}
	if datum.Time != nil {
		object["time"] = test.NewObjectFromString(*datum.Time, objectFormat)
	}
	if datum.Type != nil {
		object["type"] = test.NewObjectFromString(*datum.Type, objectFormat)
	}
	if datum.Version != nil {
		object["version"] = test.NewObjectFromString(*datum.Version, objectFormat)
	}
	return object
}

func RandomID() string {
	return test.RandomStringFromRange(1, origin.IDLengthMaximum)
}

func RandomName() string {
	return test.RandomStringFromRange(1, origin.NameLengthMaximum)
}

func RandomTime() string {
	return test.RandomTime().Format(origin.TimeFormat)
}

func RandomType() string {
	return test.RandomStringFromArray(origin.Types())
}

func RandomVersion() string {
	return test.RandomStringFromRange(1, origin.VersionLengthMaximum)
}
