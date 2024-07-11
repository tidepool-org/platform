package schema

import (
	"time"

	"github.com/tidepool-org/platform/data/types/calculator"
	"github.com/tidepool-org/platform/errors"
)

type (
	WizardBucket struct {
		Id                string    `bson:"_id,omitempty"`
		CreationTimestamp time.Time `bson:"creationTimestamp,omitempty"`
		UserId            string    `bson:"userId,omitempty" `
		Day               time.Time `bson:"day,omitempty"` // ie: 2021-09-28
		Samples           []Wizard  `bson:"samples"`
	}

	Wizard struct {
		Sample         `bson:",inline"`
		Uuid           string       `bson:"uuid,omitempty"`
		DeviceId       string       `bson:"deviceId,omitempty"`
		Guid           string       `bson:"guid,omitempty"`
		BolusId        string       `bson:"bolus,omitempty"`
		BolusIds       []string     `bson:"bolusIds,omitempty"`
		CarbInput      float64      `bson:"carbInput,omitempty"`
		InputMeal      *InputMeal   `bson:"inputMeal"`
		Recommended    *Recommended `bson:"recommended,omitempty"`
		Units          string       `bson:"units,omitempty"`
		InputTimestamp string       `bson:"inputTimestamp,omitempty"`
	}

	Recommended struct {
		Carbs      *float64 `bson:"carb,omitempty"`
		Correction *float64 `bson:"correction,omitempty"`
		Net        *float64 `bson:"net,omitempty"`
	}

	InputMeal struct {
		Fat    string `bson:"fat,omitempty"`
		Source string `bson:"source,omitempty"`
	}
)

func (wb WizardBucket) GetId() string {
	return wb.Id
}

func (w Wizard) GetTimestamp() time.Time {
	return w.Timestamp
}

func (w *Wizard) MapForWizard(event *calculator.Calculator) error {
	var err error
	if event.GUID != nil {
		w.Guid = *event.GUID
	}
	if event.DeviceID != nil {
		w.DeviceId = *event.DeviceID
	}

	if event.BolusID != nil {
		w.BolusId = *event.BolusID
	}

	if event.CarbohydrateInput != nil {
		w.CarbInput = *event.CarbohydrateInput
	}

	if event.Recommended != nil {
		w.Recommended = &Recommended{
			Carbs:      event.Recommended.Carbohydrate,
			Correction: event.Recommended.Correction,
			Net:        event.Recommended.Net,
		}
	}

	if event.Units != nil {
		w.Units = *event.Units
	}

	if event.InputTime != nil && event.InputTime.InputTime != nil {
		w.InputTimestamp = *event.InputTime.InputTime
	}

	if event.InputMeal != nil {
		i := &InputMeal{}
		if event.InputMeal.Fat != nil {
			i.Fat = *event.InputMeal.Fat
		}
		if event.InputMeal.Source != nil {
			i.Source = *event.InputMeal.Source
		}
		w.InputMeal = i
	}
	// time infos mapping
	w.Timezone = *event.TimeZoneName
	w.TimezoneOffset = *event.TimeZoneOffset
	// what is this mess ???
	strTime := *event.Time
	w.Timestamp, err = time.Parse(time.RFC3339Nano, strTime)
	if err != nil {
		return errors.Wrap(err, ErrEventTime)
	}

	return nil
}
