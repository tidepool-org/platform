package test

import (
	dataTypesSettingsController "github.com/tidepool-org/platform/data/types/settings/controller"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomNotifications() *dataTypesSettingsController.Notifications {
	datum := dataTypesSettingsController.NewNotifications()
	datum.Authorization = pointer.FromString(RandomAuthorization())
	datum.Alert = pointer.FromBool(test.RandomBool())
	datum.CriticalAlert = pointer.FromBool(test.RandomBool())
	datum.Badge = pointer.FromBool(test.RandomBool())
	datum.Sound = pointer.FromBool(test.RandomBool())
	datum.Announcement = pointer.FromBool(test.RandomBool())
	datum.NotificationCenter = pointer.FromBool(test.RandomBool())
	datum.LockScreen = pointer.FromBool(test.RandomBool())
	datum.AlertStyle = pointer.FromString(RandomAlertStyle())
	return datum
}

func RandomAlertStyle() string {
	return test.RandomStringFromArray(dataTypesSettingsController.AlertStyles())
}

func RandomAuthorization() string {
	return test.RandomStringFromArray(dataTypesSettingsController.Authorizations())
}

func CloneNotifications(datum *dataTypesSettingsController.Notifications) *dataTypesSettingsController.Notifications {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsController.NewNotifications()
	clone.Authorization = pointer.CloneString(datum.Authorization)
	clone.Alert = pointer.CloneBool(datum.Alert)
	clone.CriticalAlert = pointer.CloneBool(datum.CriticalAlert)
	clone.Badge = pointer.CloneBool(datum.Badge)
	clone.Sound = pointer.CloneBool(datum.Sound)
	clone.Announcement = pointer.CloneBool(datum.Announcement)
	clone.NotificationCenter = pointer.CloneBool(datum.NotificationCenter)
	clone.LockScreen = pointer.CloneBool(datum.LockScreen)
	clone.AlertStyle = pointer.CloneString(datum.AlertStyle)
	return clone
}

func NewObjectFromNotifications(datum *dataTypesSettingsController.Notifications, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Authorization != nil {
		object["authorization"] = test.NewObjectFromString(*datum.Authorization, objectFormat)
	}
	if datum.Alert != nil {
		object["alert"] = test.NewObjectFromBool(*datum.Alert, objectFormat)
	}
	if datum.CriticalAlert != nil {
		object["criticalAlert"] = test.NewObjectFromBool(*datum.CriticalAlert, objectFormat)
	}
	if datum.Badge != nil {
		object["badge"] = test.NewObjectFromBool(*datum.Badge, objectFormat)
	}
	if datum.Sound != nil {
		object["sound"] = test.NewObjectFromBool(*datum.Sound, objectFormat)
	}
	if datum.Announcement != nil {
		object["announcement"] = test.NewObjectFromBool(*datum.Announcement, objectFormat)
	}
	if datum.NotificationCenter != nil {
		object["notificationCenter"] = test.NewObjectFromBool(*datum.NotificationCenter, objectFormat)
	}
	if datum.LockScreen != nil {
		object["lockScreen"] = test.NewObjectFromBool(*datum.LockScreen, objectFormat)
	}
	if datum.AlertStyle != nil {
		object["alertStyle"] = test.NewObjectFromString(*datum.AlertStyle, objectFormat)
	}
	return object
}
