package test

import (
	"github.com/tidepool-org/platform/data"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/test"
)

func NewDeduplicatorDescriptor() *data.DeduplicatorDescriptor {
	datum := data.NewDeduplicatorDescriptor()
	datum.Name = netTest.RandomReverseDomain()
	datum.Version = netTest.RandomSemanticVersion()
	datum.Hash = test.NewString(32, test.CharsetHexidecimalLowercase)
	return datum
}

func CloneDeduplicatorDescriptor(datum *data.DeduplicatorDescriptor) *data.DeduplicatorDescriptor {
	if datum == nil {
		return nil
	}
	clone := data.NewDeduplicatorDescriptor()
	clone.Name = datum.Name
	clone.Version = datum.Version
	clone.Hash = datum.Hash
	return clone
}
