package test

import (
	dataTest "github.com/tidepool-org/platform/data/test"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewClient() *dataTypesUpload.Client {
	datum := dataTypesUpload.NewClient()
	datum.Name = pointer.FromString(netTest.RandomReverseDomain())
	datum.Version = pointer.FromString(netTest.RandomSemanticVersion())
	datum.Private = dataTest.NewBlob()
	return datum
}

func CloneClient(datum *dataTypesUpload.Client) *dataTypesUpload.Client {
	if datum == nil {
		return nil
	}
	clone := dataTypesUpload.NewClient()
	clone.Name = test.CloneString(datum.Name)
	clone.Version = test.CloneString(datum.Version)
	clone.Private = dataTest.CloneBlob(datum.Private)
	return clone
}
