package test

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/id"
	testInternet "github.com/tidepool-org/platform/test/internet"
)

func NewDeduplicatorDescriptor() *data.DeduplicatorDescriptor {
	datum := data.NewDeduplicatorDescriptor()
	datum.Name = testInternet.NewReverseDomain()
	datum.Version = testInternet.NewSemanticVersion()
	datum.Hash = id.New()
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
