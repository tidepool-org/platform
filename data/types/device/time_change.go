package device

import (
	"reflect"

	validator "gopkg.in/bluesuncorp/validator.v8"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

func init() {
	//types.GetPlatformValidator().RegisterValidation(timeChangeReasonsField.Tag, TimeChangeReasonsValidator)
	types.GetPlatformValidator().RegisterValidation(timeChangeAgentField.Tag, TimeChangeAgentValidator)
}

type TimeChange struct {
	Change `json:"change" bson:"change"`
	Base   `bson:",inline"`
}

type Change struct {
	From     *string   `json:"from" bson:"from" valid:"timestr"`
	To       *string   `json:"to" bson:"to" valid:"timestr"`
	Agent    *string   `json:"agent" bson:"agent" valid:"changeagent"`
	Timezone *string   `json:"timezone,omitempty" bson:"timezone,omitempty"`
	Reasons  *[]string `json:"reasons" bson:"reasons" valid:"omitempty,changereasons"`
}

var (
	timeChangeReasonsField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "reasons"},
		Tag:        "timechangereasons",
		Message:    "Must be any of from_daylight_savings, to_daylight_savings, travel, correction, other",
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
		Tag:        "changeagent",
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
	change := Change{
		From:     datum.ToString(timeChangeFromField.Name, errs),
		To:       datum.ToString(timeChangeToField.Name, errs),
		Agent:    datum.ToString(timeChangeAgentField.Name, errs),
		Timezone: datum.ToString(timeChangeTimezoneField.Name, errs),
	}

	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(change, errs)

	return change
}

func (b Base) makeTimeChange(datum types.Datum, errs validate.ErrorProcessing) *TimeChange {

	change := Change{}
	changeDatum, ok := datum["change"].(map[string]interface{})
	if ok {
		change = makeChange(changeDatum, errs)
	}

	return &TimeChange{
		Change: change,
		Base:   b,
	}
}

func TimeChangeAgentValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {

	agent, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = timeChangeAgentField.Allowed[agent]
	return ok
}

/*func TimeChangeReasonsValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {

	log.Println("## TimeChangeReasonsValidator  ##")
	return false
	//log.Println("## TimeChangeReasonsValidator  ##")
	reasons, ok := field.Interface().([]string)
	if !ok {
		//log.Println("## TimeChangeReasonsValidator  ##", reasons)
		return false
	}

	for i := range reasons {
		_, ok = timeChangeReasonsField.Allowed[reasons[i]]
		if !ok {
			return false
		}
	}
	return ok
}*/
