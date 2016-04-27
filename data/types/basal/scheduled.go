package basal

import (
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

type Scheduled struct {
	ScheduleName *string  `json:"scheduleName,omitempty" bson:"scheduleName,omitempty"`
	Duration     *int     `json:"duration" bson:"duration" valid:"basalduration"`
	Rate         *float64 `json:"rate" bson:"rate" valid:"required,basalrate"`

	Base `bson:",inline"`
}

var scheduleNameField = types.DatumField{Name: "scheduleName"}

func (b Base) makeScheduled(datum types.Datum, errs validate.ErrorProcessing) *Scheduled {
	scheduled := &Scheduled{
		ScheduleName: datum.ToString(scheduleNameField.Name, errs),
		Rate:         datum.ToFloat64(rateField.Name, errs),
		Duration:     datum.ToInt(durationField.Name, errs),
		Base:         b,
	}
	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(scheduled, errs)
	return scheduled
}
