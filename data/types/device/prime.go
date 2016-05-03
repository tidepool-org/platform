package device

import (
	"reflect"

	validator "gopkg.in/bluesuncorp/validator.v8"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

func init() {
	types.GetPlatformValidator().RegisterValidation(primeTargetField.Tag, PrimeTargetValidator)
}

type Prime struct {
	PrimeTarget *string  `json:"primeTarget" bson:"primeTarget" valid:"deviceprimetarget"`
	Volume      *float64 `json:"volume,omitempty" bson:"volume,omitempty" valid:"omitempty,devicevolume"`
	Base        `bson:",inline"`
}

var (
	cannulaVolumeField = types.FloatDatumField{
		DatumField:        &types.DatumField{Name: "volume"},
		Tag:               "devicevolume",
		Message:           "Must be >= 0.0 and <= 3.0",
		AllowedFloatRange: &types.AllowedFloatRange{LowerLimit: 0.0, UpperLimit: 3.0},
	}

	tubingVolumeField = types.FloatDatumField{
		DatumField:        &types.DatumField{Name: "volume"},
		Tag:               "devicevolume",
		Message:           "Must be >= 0.0 and <= 100.0",
		AllowedFloatRange: &types.AllowedFloatRange{LowerLimit: 0.0, UpperLimit: 100.0},
	}

	primeTargetField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "primeTarget"},
		Tag:        "deviceprimetarget",
		Message:    "Must be one of cannula, tubing",
		Allowed: types.Allowed{
			"cannula": true,
			"tubing":  true,
		},
	}
)

func (b Base) makePrime(datum types.Datum, errs validate.ErrorProcessing) *Prime {

	var prime *Prime

	primeTarget := datum.ToString(primeTargetField.Name, errs)

	if primeTarget == nil {
		prime = &Prime{
			PrimeTarget: primeTarget,
			Base:        b,
		}
	} else if *primeTarget == "cannula" {
		prime = &Prime{
			PrimeTarget: primeTarget,
			Volume:      datum.ToFloat64(cannulaVolumeField.Name, errs),
			Base:        b,
		}
		types.GetPlatformValidator().RegisterValidation(cannulaVolumeField.Tag, CannulaVolumeValidator)
		failureReasons["Volume"] = validate.ValidationInfo{FieldName: cannulaVolumeField.Name, Message: cannulaVolumeField.Message}
	} else {
		prime = &Prime{
			PrimeTarget: primeTarget,
			Volume:      datum.ToFloat64(tubingVolumeField.Name, errs),
			Base:        b,
		}
		types.GetPlatformValidator().RegisterValidation(tubingVolumeField.Tag, TubingVolumeValidator)
		failureReasons["Volume"] = validate.ValidationInfo{FieldName: tubingVolumeField.Name, Message: tubingVolumeField.Message}
	}

	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(prime, errs)
	return prime
}

func CannulaVolumeValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	volume, ok := field.Interface().(float64)
	if !ok {
		return false
	}

	return volume >= cannulaVolumeField.LowerLimit && volume <= cannulaVolumeField.UpperLimit
}

func TubingVolumeValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	volume, ok := field.Interface().(float64)
	if !ok {
		return false
	}

	return volume >= tubingVolumeField.LowerLimit && volume <= tubingVolumeField.UpperLimit
}

func PrimeTargetValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	target, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = primeTargetField.Allowed[target]
	return ok
}
