package test

import (
	"time"

	dataTypesAlert "github.com/tidepool-org/platform/data/types/alert"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomAlert() *dataTypesAlert.Alert {
	datum := randomAlert()
	datum.Base = *dataTypesTest.RandomBase()
	datum.Type = "alert"
	return datum
}

func RandomAlertForParser() *dataTypesAlert.Alert {
	datum := randomAlert()
	datum.Base = *dataTypesTest.RandomBaseForParser()
	datum.Type = "alert"
	return datum
}

func randomAlert() *dataTypesAlert.Alert {
	datum := dataTypesAlert.New()
	datum.Name = pointer.FromString(test.RandomStringFromRange(1, dataTypesAlert.NameLengthMaximum))
	datum.Priority = pointer.FromString(test.RandomStringFromArray(dataTypesAlert.Priorities()))
	datum.Trigger = pointer.FromString(test.RandomStringFromArray(dataTypesAlert.Triggers()))
	if *datum.Trigger == dataTypesAlert.TriggerDelayed || *datum.Trigger == dataTypesAlert.TriggerRepeating {
		datum.TriggerDelay = pointer.FromInt(test.RandomIntFromRange(dataTypesAlert.TriggerDelayMinimum, dataTypesAlert.TriggerDelayMaximum))
	}
	datum.Sound = pointer.FromString(test.RandomStringFromArray(dataTypesAlert.Sounds()))
	if *datum.Sound == dataTypesAlert.SoundName {
		datum.SoundName = pointer.FromString(test.RandomStringFromRange(1, dataTypesAlert.SoundNameLengthMaximum))
	}
	datum.IssuedTime = pointer.FromTime(test.RandomTimeBeforeNow())
	if test.RandomBool() {
		datum.AcknowledgedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.IssuedTime, time.Now()))
	} else {
		datum.RetractedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.IssuedTime, time.Now()))
	}
	return datum
}

func CloneAlert(datum *dataTypesAlert.Alert) *dataTypesAlert.Alert {
	if datum == nil {
		return nil
	}
	clone := dataTypesAlert.New()
	clone.Base = *dataTypesTest.CloneBase(&datum.Base)
	clone.Name = pointer.CloneString(datum.Name)
	clone.Priority = pointer.CloneString(datum.Priority)
	clone.Trigger = pointer.CloneString(datum.Trigger)
	clone.TriggerDelay = pointer.CloneInt(datum.TriggerDelay)
	clone.Sound = pointer.CloneString(datum.Sound)
	clone.SoundName = pointer.CloneString(datum.SoundName)
	clone.IssuedTime = pointer.CloneTime(datum.IssuedTime)
	clone.AcknowledgedTime = pointer.CloneTime(datum.AcknowledgedTime)
	clone.RetractedTime = pointer.CloneTime(datum.RetractedTime)
	return clone
}

func NewObjectFromAlert(datum *dataTypesAlert.Alert, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := dataTypesTest.NewObjectFromBase(&datum.Base, objectFormat)
	if datum.Name != nil {
		object["name"] = test.NewObjectFromString(*datum.Name, objectFormat)
	}
	if datum.Priority != nil {
		object["priority"] = test.NewObjectFromString(*datum.Priority, objectFormat)
	}
	if datum.Trigger != nil {
		object["trigger"] = test.NewObjectFromString(*datum.Trigger, objectFormat)
	}
	if datum.TriggerDelay != nil {
		object["triggerDelay"] = test.NewObjectFromInt(*datum.TriggerDelay, objectFormat)
	}
	if datum.Sound != nil {
		object["sound"] = test.NewObjectFromString(*datum.Sound, objectFormat)
	}
	if datum.SoundName != nil {
		object["soundName"] = test.NewObjectFromString(*datum.SoundName, objectFormat)
	}
	if datum.IssuedTime != nil {
		object["issuedTime"] = test.NewObjectFromTime(*datum.IssuedTime, objectFormat)
	}
	if datum.AcknowledgedTime != nil {
		object["acknowledgedTime"] = test.NewObjectFromTime(*datum.AcknowledgedTime, objectFormat)
	}
	if datum.RetractedTime != nil {
		object["retractedTime"] = test.NewObjectFromTime(*datum.RetractedTime, objectFormat)
	}
	return object
}
