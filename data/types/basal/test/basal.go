package test

import (
	"github.com/tidepool-org/platform/data/types/basal"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/test"
)

func NewBasal() *basal.Basal {
	datum := &basal.Basal{}
	datum.Base = *dataTypesTest.NewBase()
	datum.Type = "basal"
	datum.DeliveryType = dataTypesTest.NewType()
	return datum
}

func CloneBasal(datum *basal.Basal) *basal.Basal {
	if datum == nil {
		return nil
	}
	clone := &basal.Basal{}
	clone.Base = *dataTypesTest.CloneBase(&datum.Base)
	clone.DeliveryType = datum.DeliveryType
	return clone
}

func NewScheduleName() string {
	return test.RandomStringFromRange(1, 32)
}
