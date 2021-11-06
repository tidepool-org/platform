package test

import (
	dataTypesBasal "github.com/tidepool-org/platform/data/types/basal"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/test"
)

func RandomBasal() *dataTypesBasal.Basal {
	datum := randomBasal()
	datum.Base = *dataTypesTest.RandomBase()
	datum.Type = "basal"
	return datum
}

func RandomBasalForParser() *dataTypesBasal.Basal {
	datum := randomBasal()
	datum.Base = *dataTypesTest.RandomBaseForParser()
	datum.Type = "basal"
	return datum
}

func randomBasal() *dataTypesBasal.Basal {
	datum := &dataTypesBasal.Basal{}
	datum.DeliveryType = dataTypesTest.NewType()
	return datum
}

func CloneBasal(datum *dataTypesBasal.Basal) *dataTypesBasal.Basal {
	if datum == nil {
		return nil
	}
	clone := &dataTypesBasal.Basal{}
	clone.Base = *dataTypesTest.CloneBase(&datum.Base)
	clone.DeliveryType = datum.DeliveryType
	return clone
}

func NewObjectFromBasal(datum *dataTypesBasal.Basal, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := dataTypesTest.NewObjectFromBase(&datum.Base, objectFormat)
	object["deliveryType"] = test.NewObjectFromString(datum.DeliveryType, objectFormat)
	return object
}

func RandomScheduleName() string {
	return test.RandomStringFromRange(1, dataTypesBasal.ScheduleNameLengthMaximum)
}
