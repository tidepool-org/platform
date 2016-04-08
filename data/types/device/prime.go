package device

import (
	"reflect"

	"github.com/tidepool-org/platform/data/types"
	"gopkg.in/bluesuncorp/validator.v8"

	"github.com/tidepool-org/platform/validate"
)

func init() {
	types.GetPlatformValidator().RegisterValidation(primeTargetField.Tag, PrimeTargetValidator)
	types.GetPlatformValidator().RegisterValidation(volumeField.Tag, VolumeValidator)
}

type Prime struct {
	PrimeTarget *string  `json:"primeTarget" bson:"primeTarget" valid:"deviceprimetarget"`
	Volume      *float64 `json:"volume,omitempty" bson:"volume,omitempty" valid:"omitempty,devicevolume"`
	Base        `bson:",inline"`
}

var (
	volumeField = types.FloatDatumField{
		DatumField:        &types.DatumField{Name: "volume"},
		Tag:               "devicevolume",
		Message:           "Must be greater than 0.0",
		AllowedFloatRange: &types.AllowedFloatRange{LowerLimit: 0.0},
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
	prime := &Prime{
		PrimeTarget: datum.ToString(primeTargetField.Name, errs),
		Volume:      datum.ToFloat64(volumeField.Name, errs),
		Base:        b,
	}
	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(prime, errs)
	return prime
}

func VolumeValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	volume, ok := field.Interface().(float64)
	if !ok {
		return false
	}
	return volume > volumeField.LowerLimit
}

func PrimeTargetValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	target, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = primeTargetField.Allowed[target]
	return ok
}
