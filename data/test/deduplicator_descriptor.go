package test

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/id"
	testInternet "github.com/tidepool-org/platform/test/internet"
)

func NewDeduplicatorDescriptor() *data.DeduplicatorDescriptor {
	return &data.DeduplicatorDescriptor{
		Name:    testInternet.NewReverseDomain(),
		Version: testInternet.NewSemanticVersion(),
		Hash:    id.New(),
	}
}
