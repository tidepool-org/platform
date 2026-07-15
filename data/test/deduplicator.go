package test

import (
	"github.com/tidepool-org/platform/data"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomHash() string {
	return test.RandomStringFromRangeAndCharset(32, 32, test.CharsetHexadecimalLowercase)
}

func RandomDeduplicatorDescriptor(options ...test.Option) *data.DeduplicatorDescriptor {
	datum := data.NewDeduplicatorDescriptor()
	datum.Name = test.RandomOptional(netTest.RandomReverseDomain, options...)
	datum.Version = test.RandomOptional(netTest.RandomSemanticVersion, options...)
	datum.Hash = test.RandomOptional(RandomHash, options...)
	return datum
}

func CloneDeduplicatorDescriptor(datum *data.DeduplicatorDescriptor) *data.DeduplicatorDescriptor {
	if datum == nil {
		return nil
	}
	clone := data.NewDeduplicatorDescriptor()
	clone.Name = pointer.CloneString(datum.Name)
	clone.Version = pointer.CloneString(datum.Version)
	clone.Hash = pointer.CloneString(datum.Hash)
	return clone
}

func NewObjectFromDeduplicatorDescriptor(datum *data.DeduplicatorDescriptor, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Name != nil {
		object["name"] = test.NewObjectFromString(*datum.Name, objectFormat)
	}
	if datum.Version != nil {
		object["version"] = test.NewObjectFromString(*datum.Version, objectFormat)
	}
	if datum.Hash != nil {
		object["hash"] = test.NewObjectFromString(*datum.Hash, objectFormat)
	}
	return object
}
