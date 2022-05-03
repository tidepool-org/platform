package test

import (
	dataTypesStatusController "github.com/tidepool-org/platform/data/types/status/controller"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/test"
)

func RandomController() *dataTypesStatusController.Controller {
	datum := randomController()
	datum.Base = *dataTypesTest.RandomBase()
	datum.Type = "controllerStatus"
	return datum
}

func RandomControllerForParser() *dataTypesStatusController.Controller {
	datum := randomController()
	datum.Base = *dataTypesTest.RandomBaseForParser()
	datum.Type = "controllerStatus"
	return datum
}

func randomController() *dataTypesStatusController.Controller {
	datum := dataTypesStatusController.New()
	datum.Battery = RandomBattery()
	return datum
}

func CloneController(datum *dataTypesStatusController.Controller) *dataTypesStatusController.Controller {
	if datum == nil {
		return nil
	}
	clone := dataTypesStatusController.New()
	clone.Base = *dataTypesTest.CloneBase(&datum.Base)
	clone.Battery = CloneBattery(datum.Battery)
	return clone
}

func NewObjectFromController(datum *dataTypesStatusController.Controller, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := dataTypesTest.NewObjectFromBase(&datum.Base, objectFormat)
	if datum.Battery != nil {
		object["battery"] = NewObjectFromBattery(datum.Battery, objectFormat)
	}
	return object
}
