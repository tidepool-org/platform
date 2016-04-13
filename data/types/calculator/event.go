package calculator

import (
	"reflect"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

func init() {
	types.GetPlatformValidator().RegisterValidation(insulinSensitivityField.Tag, InsulinSensitivityValidator)
	types.GetPlatformValidator().RegisterValidation(insulinOnBoardField.Tag, InsulinOnBoardValidator)
}

type Event struct {
	*Recommended        `json:"recommended,omitempty" bson:"recommended,omitempty" valid:"-"`
	CarbohydrateInput   *int     `json:"carbInput,omitempty" bson:"carbInput,omitempty" valid:"omitempty,required"`
	InsulinOnBoard      *float64 `json:"insulinOnBoard,omitempty" bson:"insulinOnBoard,omitempty" valid:"omitempty,insulinvalue"`
	InsulinSensitivity  *int     `json:"insulinSensitivity,omitempty" bson:"insulinSensitivity,omitempty" valid:"omitempty,insulinsensitivity"`
	*BloodGlucoseTarget `json:"bgTarget,omitempty" bson:"bgTarget,omitempty" valid:"-"`
	BloodGlucoseInput   *float64 `json:"bgInput,omitempty" bson:"bgInput,omitempty" valid:"omitempty,bloodglucosevalue"`
	*Bolus              `json:"bolus,omitempty" bson:"bolus,omitempty" valid:"-"`
	Units               *string `json:"units" bson:"units" valid:"mmolmgunits"`
	types.Base          `bson:",inline"`
}

type Recommended struct {
	Carbohydrate *float64 `json:"carb" bson:"carb" valid:"required"`
	Correction   *float64 `json:"correction" bson:"correction" valid:"required"`
	Net          *float64 `json:"net" bson:"net" valid:"required"`
}

type Bolus struct {
	Type     *string `json:"type" bson:"type" valid:"-"`
	SubType  *string `json:"subType" bson:"subType" valid:"bolussubtype"`
	Time     *string `json:"time" bson:"time" valid:"timestr"`
	DeviceID *string `json:"deviceId" bson:"deviceId" valid:"required"`
}

type BloodGlucoseTarget struct {
	High *float64 `json:"high" bson:"high" valid:"bloodvalue"`
	Low  *float64 `json:"low" bson:"low" valid:"bloodvalue"`
}

//NOTE: for backwards compatatbilty the type name is `wizard` but will be migrated to `calculator`
const Name = "wizard"

var (
	carbohydrateInputField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "carbInput"},
		Tag:        "required",
		Message:    "This is a required field",
	}

	bloodGlucoseInputField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "bgInput"},
		Tag:        types.BloodGlucoseValueField.Tag,
		Message:    types.BloodGlucoseValueField.Message,
	}

	insulinSensitivityField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "insulinSensitivity"},
		Tag:        "insulinsensitivity",
		Message:    "This is a required field",
	}

	insulinOnBoardField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "insulinOnBoard"},
		Tag:        "insulinvalue",
		Message:    "This is a required field",
	}

	carbField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "carb"},
		Tag:        "required",
		Message:    "This is a required field",
	}

	netField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "net"},
		Tag:        "required",
		Message:    "This is a required field",
	}

	correctionField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "correction"},
		Tag:        "required",
		Message:    "This is a required field",
	}

	failureReasons = validate.FailureReasons{
		"Time":              validate.VaidationInfo{FieldName: types.TimeStringField.Name, Message: types.TimeStringField.Message},
		"BloodGlucoseInput": validate.VaidationInfo{FieldName: bloodGlucoseInputField.Name, Message: bloodGlucoseInputField.Message},
		"CarbohydrateInput": validate.VaidationInfo{FieldName: carbohydrateInputField.Name, Message: carbohydrateInputField.Message},
		"Units":             validate.VaidationInfo{FieldName: types.MmolOrMgUnitsField.Name, Message: types.MmolOrMgUnitsField.Message},
		"SubType":           validate.VaidationInfo{FieldName: types.BolusSubTypeField.Name, Message: types.BolusSubTypeField.Message},
		"InsulinOnBoard":    validate.VaidationInfo{FieldName: types.BolusSubTypeField.Name, Message: types.BolusSubTypeField.Message},

		"High": validate.VaidationInfo{FieldName: "high", Message: types.BloodGlucoseValueField.Message},
		"Low":  validate.VaidationInfo{FieldName: "low", Message: types.BloodGlucoseValueField.Message},

		"Net":          validate.VaidationInfo{FieldName: netField.Name, Message: netField.Message},
		"Correction":   validate.VaidationInfo{FieldName: correctionField.Name, Message: correctionField.Message},
		"Carbohydrate": validate.VaidationInfo{FieldName: carbField.Name, Message: carbField.Message},
	}
)

func buildRecommended(datum types.Datum, errs validate.ErrorProcessing) *Recommended {
	recommendedDatum, ok := datum["recommended"].(types.Datum)
	if ok {
		return &Recommended{
			Carbohydrate: recommendedDatum.ToFloat64(carbField.Name, errs),
			Correction:   recommendedDatum.ToFloat64(correctionField.Name, errs),
			Net:          recommendedDatum.ToFloat64(netField.Name, errs),
		}
	}
	return nil
}

func buildBolus(datum types.Datum, errs validate.ErrorProcessing) *Bolus {
	bolusDatum, ok := datum["bolus"].(types.Datum)
	if ok {
		bolus := "bolus"
		return &Bolus{
			Type:     &bolus,
			SubType:  bolusDatum.ToString(types.BolusSubTypeField.Name, errs),
			Time:     bolusDatum.ToString(types.TimeStringField.Name, errs),
			DeviceID: bolusDatum.ToString("deviceId", errs),
		}
	}
	return nil
}

func buildBloodGlucoseTarget(datum types.Datum, errs validate.ErrorProcessing) *BloodGlucoseTarget {
	bgTargetDatum, ok := datum["bgTarget"].(types.Datum)
	if ok {
		return &BloodGlucoseTarget{
			High: bgTargetDatum.ToFloat64("high", errs),
			Low:  bgTargetDatum.ToFloat64("low", errs),
		}
	}
	return nil
}

func Build(datum types.Datum, errs validate.ErrorProcessing) *Event {

	event := &Event{
		Recommended:        buildRecommended(datum, errs),
		Bolus:              buildBolus(datum, errs),
		BloodGlucoseTarget: buildBloodGlucoseTarget(datum, errs),
		CarbohydrateInput:  datum.ToInt(carbohydrateInputField.Name, errs),
		InsulinOnBoard:     datum.ToFloat64(insulinOnBoardField.Name, errs),
		InsulinSensitivity: datum.ToInt(insulinSensitivityField.Name, errs),
		BloodGlucoseInput:  datum.ToFloat64(bloodGlucoseInputField.Name, errs),
		Units:              datum.ToString(types.MmolOrMgUnitsField.Name, errs),
		Base:               types.BuildBase(datum, errs),
	}

	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(event, errs)

	return event
}

func InsulinSensitivityValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	_, ok := field.Interface().(float64)
	if !ok {
		return false
	}
	//TODO: correct validation here
	return true
}

func InsulinOnBoardValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	iob, ok := field.Interface().(float64)
	if !ok {
		return false
	}
	return iob >= 0.0
}
