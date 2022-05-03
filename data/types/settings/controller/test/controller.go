package test

import (
	dataTypesSettingsController "github.com/tidepool-org/platform/data/types/settings/controller"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/test"
)

func RandomController() *dataTypesSettingsController.Controller {
	datum := randomController()
	datum.Base = *dataTypesTest.RandomBase()
	datum.Type = "controllerSettings"
	return datum
}

func RandomControllerForParser() *dataTypesSettingsController.Controller {
	datum := randomController()
	datum.Base = *dataTypesTest.RandomBaseForParser()
	datum.Type = "controllerSettings"
	return datum
}

func randomController() *dataTypesSettingsController.Controller {
	datum := dataTypesSettingsController.New()
	datum.Device = RandomDevice()
	datum.Notifications = RandomNotifications()
	return datum
}

func CloneController(datum *dataTypesSettingsController.Controller) *dataTypesSettingsController.Controller {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsController.New()
	clone.Base = *dataTypesTest.CloneBase(&datum.Base)
	clone.Device = CloneDevice(datum.Device)
	clone.Notifications = CloneNotifications(datum.Notifications)
	return clone
}

func NewObjectFromController(datum *dataTypesSettingsController.Controller, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := dataTypesTest.NewObjectFromBase(&datum.Base, objectFormat)
	if datum.Device != nil {
		object["device"] = NewObjectFromDevice(datum.Device, objectFormat)
	}
	if datum.Notifications != nil {
		object["notifications"] = NewObjectFromNotifications(datum.Notifications, objectFormat)
	}
	return object
}
