package data

import (
	"reflect"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"
	"github.com/tidepool-org/platform/validate"
)

func init() {
	getPlatformValidator().RegisterValidation(bolusInsulinField.Tag, BolusInsulinValidator)
	getPlatformValidator().RegisterValidation(bolusValueField.Tag, BolusValueValidator)
	getPlatformValidator().RegisterValidation(bolusExtendedField.Tag, BolusExtendedValidator)
	getPlatformValidator().RegisterValidation(bolusNormalField.Tag, BolusNormalValidator)
	getPlatformValidator().RegisterValidation(bolusDurationField.Tag, BolusDurationValidator)
	getPlatformValidator().RegisterValidation(bolusSubTypeField.Tag, BolusSubTypeValidator)
}

type Bolus struct {
	SubType  *string  `json:"subType" bson:"subType" valid:"bolussubtype"`
	Insulin  *string  `json:"insulin,omitempty" bson:"insulin,omitempty" valid:"omitempty,bolusinsulin"`
	Normal   *float64 `json:"normal,omitempty" bson:"normal,omitempty" valid:"omitempty,bolusnormal"`
	Extended *float64 `json:"extended,omitempty" bson:"extended,omitempty" valid:"omitempty,bolusextended"`
	Duration *int     `json:"duration,omitempty" bson:"duration,omitempty" valid:"omitempty,bolusduration"`
	Value    *int     `json:"value,omitempty" bson:"value,omitempty" valid:"omitempty,bolusvalue"`
	Base     `bson:",inline"`
}

const BolusName = "Bolus"

var (
	bolusSubTypeField = TypesDatumField{
		DatumField:   &DatumField{Name: "subType"},
		Tag:          "bolussubtype",
		Message:      "Must be one of injected, normal, square, dual/square",
		AllowedTypes: AllowedTypes{"injected": true, "normal": true, "square": true, "dual/square": true},
	}

	bolusInsulinField = TypesDatumField{
		DatumField:   &DatumField{Name: "insulin"},
		Tag:          "bolusinsulin",
		Message:      "Must be one of novolog, humalog",
		AllowedTypes: AllowedTypes{"novolog": true, "humalog": true},
	}

	bolusNormalField = FloatDatumField{
		DatumField:        &DatumField{Name: "normal"},
		Tag:               "bolusnormal",
		Message:           "Must be greater than 0.0",
		AllowedFloatRange: &AllowedFloatRange{LowerLimit: 0.0},
	}

	bolusExtendedField = FloatDatumField{
		DatumField:        &DatumField{Name: "extended"},
		Tag:               "bolusextended",
		Message:           "Must be greater than 0.0",
		AllowedFloatRange: &AllowedFloatRange{LowerLimit: 0.0},
	}

	bolusDurationField = IntDatumField{
		DatumField:      &DatumField{Name: "duration"},
		Tag:             "bolusduration",
		Message:         "Must be greater than 0",
		AllowedIntRange: &AllowedIntRange{LowerLimit: 0},
	}

	bolusValueField = IntDatumField{
		DatumField:      &DatumField{Name: "value"},
		Tag:             "bolusvalue",
		Message:         "Must be greater than 0",
		AllowedIntRange: &AllowedIntRange{LowerLimit: 0},
	}

	bolusFailureReasons = validate.ErrorReasons{
		bolusValueField.Tag:    bolusValueField.Message,
		bolusNormalField.Tag:   bolusNormalField.Message,
		bolusExtendedField.Tag: bolusExtendedField.Message,
		bolusDurationField.Tag: bolusDurationField.Message,
		bolusInsulinField.Tag:  bolusInsulinField.Message,
		bolusSubTypeField.Tag:  bolusSubTypeField.Message,
	}
)

func BuildBolus(datum Datum, errs validate.ErrorProcessing) *Bolus {

	bolus := &Bolus{
		SubType:  datum.ToString(bolusSubTypeField.Name, errs),
		Insulin:  datum.ToString(bolusInsulinField.Name, errs),
		Value:    datum.ToInt(bolusValueField.Name, errs),
		Duration: datum.ToInt(bolusDurationField.Name, errs),
		Normal:   datum.ToFloat64(bolusNormalField.Name, errs),
		Extended: datum.ToFloat64(bolusExtendedField.Name, errs),
		Base:     BuildBase(datum, errs),
	}

	getPlatformValidator().SetErrorReasons(bolusFailureReasons).Struct(bolus, errs)

	return bolus
}

func BolusNormalValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	normal, ok := field.Interface().(float64)
	if !ok {
		return false
	}
	return normal > bolusNormalField.LowerLimit
}

func BolusExtendedValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	extended, ok := field.Interface().(float64)
	if !ok {
		return false
	}
	return extended > bolusExtendedField.LowerLimit
}

func BolusDurationValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	duration, ok := field.Interface().(int)
	if !ok {
		return false
	}
	return duration > bolusDurationField.LowerLimit
}

func BolusSubTypeValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	subType, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = bolusSubTypeField.AllowedTypes[subType]
	return ok
}

func BolusInsulinValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	insulin, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = bolusInsulinField.AllowedTypes[insulin]
	return ok
}

func BolusValueValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	value, ok := field.Interface().(int)
	if !ok {
		return false
	}
	return value > bolusValueField.LowerLimit
}
