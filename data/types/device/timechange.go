package device

import (
	"reflect"

	"github.com/tidepool-org/platform/data/types"
	"gopkg.in/bluesuncorp/validator.v8"

	"github.com/tidepool-org/platform/validate"
)

func init() {
	types.GetPlatformValidator().RegisterValidation(timeChangeReasonsField.Tag, TimeChangeReasonsValidator)
	types.GetPlatformValidator().RegisterValidation(timeChangeAgentField.Tag, TimeChangeAgentValidator)
}

type TimeChange struct {
	Change `json:"change" bson:"change"`
	Base   `bson:",inline"`
}

type Change struct {
	From     *string   `json:"from" bson:"from" valid:"timestr"`
	To       *string   `json:"to" bson:"to" valid:"timestr"`
	Agent    *string   `json:"from" bson:"from" valid:"devicechangeagent"`
	Timezone *string   `json:"timezone,omitempty" bson:"timezone,omitempty" valid:"-"`
	Reasons  *[]string `json:"reasons,omitempty" bson:"reasons,omitempty" valid:"omitempty,devicechangereasons"`
}

var (
	timeChangeReasonsField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "reasons"},
		Tag:        "devicechangereasons",
		Message:    "Must be one of from_daylight_savings, to_daylight_savings, travel, correction, other",
		Allowed: types.Allowed{
			"from_daylight_savings": true,
			"to_daylight_savings":   true,
			"travel":                true,
			"correction":            true,
			"other":                 true,
		},
	}

	timeChangeAgentField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "agent"},
		Tag:        "devicechangeagent",
		Message:    "Must be one of manual, automatic",
		Allowed: types.Allowed{
			"manual":    true,
			"automatic": true,
		},
	}

	timeChangeFromField     = types.DatumField{Name: "from"}
	timeChangeToField       = types.DatumField{Name: "to"}
	timeChangeTimezoneField = types.DatumField{Name: "timezone"}
)

func makeChange(datum types.Datum, errs validate.ErrorProcessing) Change {
	return Change{
		From:     datum.ToString(timeChangeFromField.Name, errs),
		To:       datum.ToString(timeChangeToField.Name, errs),
		Agent:    datum.ToString(timeChangeAgentField.Name, errs),
		Timezone: datum.ToString(timeChangeTimezoneField.Name, errs),
	}
}

func (b Base) makeTimeChange(datum types.Datum, errs validate.ErrorProcessing) *TimeChange {

	change := Change{}
	changeDatum, ok := datum["change"].(map[string]interface{})
	if ok {
		change = makeChange(changeDatum, errs)
	}

	timeChange := &TimeChange{
		Change: change,
		Base:   b,
	}
	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(timeChange, errs)
	return timeChange
}

func TimeChangeAgentValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	agent, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = timeChangeAgentField.Allowed[agent]
	return ok
}

func TimeChangeReasonsValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	reasons, ok := field.Interface().([]string)
	if !ok {
		return false
	}
	for i := range reasons {
		_, ok = timeChangeReasonsField.Allowed[reasons[i]]
		if !ok {
			break
		}
	}
	return ok
}
