package test

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/test"
)

func RandomSetID() string {
	return data.NewSetID()
}

func RandomSetIDs() []string {
	return test.RandomStringArrayFromRangeAndGeneratorWithoutDuplicates(1, 3, RandomSetID)
}
