package bloodglucose

import (
	"reflect"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"
	"github.com/tidepool-org/platform/data/types"

	"github.com/tidepool-org/platform/validate"
)

func init() {
	types.GetPlatformValidator().RegisterValidation(isigField.Tag, IsigValidator)
}

const ContinuousName = "cbg"

type Continuous struct {
	Value      *float64 `json:"value" bson:"value" valid:"bloodglucosevalue"`
	Units      *string  `json:"units" bson:"units" valid:"bloodglucoseunits"`
	Isig       *float64 `json:"isig" bson:"isig" valid:"cbgisig"`
	types.Base `bson:",inline"`
}

var isigField = types.FloatDatumField{
	DatumField:        &types.DatumField{Name: "isig"},
	Tag:               "cbgisig",
	Message:           "Must be greater than 0.0",
	AllowedFloatRange: &types.AllowedFloatRange{LowerLimit: 0.0},
}

func BuildContinuous(datum types.Datum, errs validate.ErrorProcessing) *Continuous {

	continuous := &Continuous{
		Value: datum.ToFloat64(valueField.Name, errs),
		Units: datum.ToString(unitsField.Name, errs),
		Isig:  datum.ToFloat64(isigField.Name, errs),
		Base:  types.BuildBase(datum, errs),
	}

	continuous.Units = normalizeUnitName(continuous.Units)
	continuous.Value = convertMgToMmol(continuous.Value, continuous.Units)

	types.GetPlatformValidator().SetErrorReasons(failureReasons).Struct(continuous, errs)
	return continuous
}

func IsigValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	val, ok := field.Interface().(float64)
	if !ok {
		return false
	}
	return val > isigField.LowerLimit
}
