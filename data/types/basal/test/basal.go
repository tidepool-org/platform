package test

import (
	"github.com/tidepool-org/platform/data/types/basal"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/test"
)

func NewBasal() *basal.Basal {
	datum := &basal.Basal{}
	datum.Base = *testDataTypes.NewBase()
	datum.Type = "basal"
	datum.DeliveryType = testDataTypes.NewType()
	return datum
}

func CloneBasal(datum *basal.Basal) *basal.Basal {
	if datum == nil {
		return nil
	}
	clone := &basal.Basal{}
	clone.Base = *testDataTypes.CloneBase(&datum.Base)
	clone.DeliveryType = datum.DeliveryType
	return clone
}

func NewScheduleName() string {
	return test.NewText(1, 32)
}
