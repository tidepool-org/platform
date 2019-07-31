package test

import (
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/pointer"
)

func NewClient() *dataTypesUpload.Client {
	datum := dataTypesUpload.NewClient()
	datum.Name = pointer.FromString(netTest.RandomReverseDomain())
	datum.Version = pointer.FromString(netTest.RandomSemanticVersion())
	datum.Private = metadataTest.RandomMetadata()
	return datum
}

func CloneClient(datum *dataTypesUpload.Client) *dataTypesUpload.Client {
	if datum == nil {
		return nil
	}
	clone := dataTypesUpload.NewClient()
	clone.Name = pointer.CloneString(datum.Name)
	clone.Version = pointer.CloneString(datum.Version)
	clone.Private = metadataTest.CloneMetadata(datum.Private)
	return clone
}
