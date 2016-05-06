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
	types.GetPlatformValidator().RegisterValidation(SubTypeField.Tag, SubTypeValidator)
}

type Base struct {
	SubType    *string `json:"subType" bson:"subType" valid:"bolussubtype"`
	types.Base `bson:",inline"`
}

const Name = "bolus"

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
		Message:           "Must be greater than 0 and less than or equal to 100.0",
		AllowedFloatRange: &types.AllowedFloatRange{LowerLimit: 0.0, UpperLimit: 100.0},
	}

	durationField = types.IntDatumField{
		DatumField:      &types.DatumField{Name: "duration"},
		Tag:             "bolusduration",
		Message:         "Must be greater than 0 and less than 86400000",
		AllowedIntRange: &types.AllowedIntRange{LowerLimit: 0, UpperLimit: 86400000},
	}

	normalField = types.FloatDatumField{
		DatumField:        &types.DatumField{Name: "normal"},
		Tag:               "bolusnormal",
		Message:           "Must be greater than or equal to 0 and less than or equal to 100", // TODO_DATA: Tandem can have 0 normal
		AllowedFloatRange: &types.AllowedFloatRange{LowerLimit: 0.0, UpperLimit: 100.0},
	}

	SubTypeField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "subType"},
		Tag:        "bolussubtype",
		Message:    "Must be one of normal, square, dual/square",
		Allowed:    types.Allowed{"normal": true, "square": true, "dual/square": true},
	}

	failureReasons = validate.FailureReasons{
		"SubType": validate.ValidationInfo{
			FieldName: SubTypeField.Name,
			Message:   SubTypeField.Message,
		},
		"Normal": validate.ValidationInfo{
			FieldName: normalField.Name,
			Message:   normalField.Message,
		},
		"Extended": validate.ValidationInfo{
			FieldName: extendedField.Name,
			Message:   extendedField.Message,
		},
		"Duration": validate.ValidationInfo{
			FieldName: durationField.Name,
			Message:   durationField.Message,
		},
	}
)

func Build(datum types.Datum, errs validate.ErrorProcessing) interface{} {

	base := &Base{
		SubType: datum.ToString(SubTypeField.Name, errs),
		Base:    types.BuildBase(datum, errs),
	}

	if base.SubType != nil {

		//TODO: we have a naming mismatch on the `SubType` until these names are
		// migrated to reflect the name of the struct
		//  i.e. `square` => `extended`
		//  i.e. `dual/square` => `combo`
		switch *base.SubType {
		case "normal":
			return base.makeNormal(datum, errs)
		case "square":
			return base.makeExtended(datum, errs)
		case "dual/square":
			return base.makeCombo(datum, errs)
		}
	}

	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(base, errs)
	return base
}

func SubTypeValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	subType, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = SubTypeField.Allowed[subType]
	return ok
}

func NormalValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	normal, ok := field.Interface().(float64)
	if !ok {
		return false
	}
	return normal >= normalField.LowerLimit && normal <= normalField.UpperLimit // TODO_DATA: Tandem can have 0 normal
}

func ExtendedValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	extended, ok := field.Interface().(float64)
	if !ok {
		return false
	}
	return extended > extendedField.LowerLimit && extended <= extendedField.UpperLimit
}

func DurationValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	duration, ok := field.Interface().(int)
	if !ok {
		return false
	}
	return duration >= durationField.LowerLimit && duration <= durationField.UpperLimit
}
