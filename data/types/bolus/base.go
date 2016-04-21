package bolus

import (
	"reflect"

	validator "gopkg.in/bluesuncorp/validator.v8"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

func init() {
	types.GetPlatformValidator().RegisterValidation(extendedField.Tag, ExtendedValidator)
	types.GetPlatformValidator().RegisterValidation(durationField.Tag, DurationValidator)
	types.GetPlatformValidator().RegisterValidation(normalField.Tag, NormalValidator)
}

type Base struct {
	SubType    *string `json:"subType" bson:"subType" valid:"bolussubtype"`
	types.Base `bson:",inline"`
}

const Name = "Bolus"

var (
	BolusSubTypeField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "subType"},
		Tag:        "bolussubtype",
		Message:    "Must be one of normal, square, dual/square",
		Allowed:    types.Allowed{"normal": true, "square": true, "dual/square": true},
	}

	extendedField = types.FloatDatumField{
		DatumField:        &types.DatumField{Name: "extended"},
		Tag:               "bolusextended",
		Message:           "Must be greater than 0.0",
		AllowedFloatRange: &types.AllowedFloatRange{LowerLimit: 0.0},
	}

	durationField = types.IntDatumField{
		DatumField:      &types.DatumField{Name: "duration"},
		Tag:             "bolusduration",
		Message:         "Must be greater than 0",
		AllowedIntRange: &types.AllowedIntRange{LowerLimit: 0},
	}

	normalField = types.FloatDatumField{
		DatumField:        &types.DatumField{Name: "normal"},
		Tag:               "bolusnormal",
		Message:           "Must be greater than 0.0",
		AllowedFloatRange: &types.AllowedFloatRange{LowerLimit: 0.0},
	}

	failureReasons = validate.FailureReasons{
		"SubType":  validate.VaidationInfo{FieldName: types.BolusSubTypeField.Name, Message: types.BolusSubTypeField.Message},
		"Normal":   validate.VaidationInfo{FieldName: normalField.Name, Message: normalField.Message},
		"Extended": validate.VaidationInfo{FieldName: extendedField.Name, Message: extendedField.Message},
		"Duration": validate.VaidationInfo{FieldName: durationField.Name, Message: durationField.Message},
	}
)

func Build(datum types.Datum, errs validate.ErrorProcessing) interface{} {

	base := &Base{
		SubType: datum.ToString(types.BolusSubTypeField.Name, errs),
		Base:    types.BuildBase(datum, errs),
	}

	if base.SubType != nil {

		switch *base.SubType {
		case "normal":
			return base.makeNormal(datum, errs)
		case "square":
			return base.makeSquare(datum, errs)
		case "dual/square":
			return base.makeDualSquare(datum, errs)
		}
	}

	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(base, errs)
	return base
}

func NormalValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	normal, ok := field.Interface().(float64)
	if !ok {
		return false
	}
	return normal > normalField.LowerLimit
}

func ExtendedValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	extended, ok := field.Interface().(float64)
	if !ok {
		return false
	}
	return extended > extendedField.LowerLimit
}

func DurationValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	duration, ok := field.Interface().(int)
	if !ok {
		return false
	}
	return duration > durationField.LowerLimit
}
