package basal

import (
	"reflect"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"
	"github.com/tidepool-org/platform/data/types"

	"github.com/tidepool-org/platform/validate"
)

func init() {
	types.GetPlatformValidator().RegisterValidation(percentField.Tag, PercentValidator)
}

type Temporary struct {
	Rate       *float64    `json:"rate,omitempty" bson:"rate,omitempty" valid:"omitempty,basalrate"`
	Percent    *float64    `json:"percent,omitempty" bson:"percent,omitempty" valid:"omitempty,basalpercent"`
	Suppressed *Suppressed `json:"suppressed,omitempty" bson:"suppressed,omitempty" valid:"omitempty,required"`
	Base       `bson:",inline"`
}

var (
	percentField = types.FloatDatumField{
		DatumField:        &types.DatumField{Name: "percent"},
		Tag:               "basalpercent",
		Message:           "Must be greater than 0.0",
		AllowedFloatRange: &types.AllowedFloatRange{LowerLimit: 0.0, UpperLimit: 1.0},
	}
)

func (b Base) makeTemporary(datum types.Datum, errs validate.ErrorProcessing) *Temporary {

	var suppressed *Suppressed
	suppressedDatum, ok := datum["suppressed"].(map[string]interface{})
	if ok {
		suppressed = makeSuppressed(suppressedDatum, errs)
	}

	temporary := &Temporary{
		Rate:       datum.ToFloat64(rateField.Name, errs),
		Percent:    datum.ToFloat64(percentField.Name, errs),
		Suppressed: suppressed,
		Base:       b,
	}
	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(temporary, errs)
	return temporary
}

func PercentValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	percent, ok := field.Interface().(float64)
	if !ok {
		return false
	}
	return percent > percentField.LowerLimit && percent < percentField.UpperLimit
}
