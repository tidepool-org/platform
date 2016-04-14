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
	*Recommended        `json:"recommended,omitempty" bson:"recommended,omitempty"`
	*BloodGlucoseTarget `json:"bgTarget,omitempty" bson:"bgTarget,omitempty"`
	*Bolus              `json:"bolus,omitempty" bson:"bolus,omitempty"`

	CarbohydrateInput  *int     `json:"carbInput,omitempty" bson:"carbInput,omitempty" valid:"omitempty,required"`
	InsulinOnBoard     *float64 `json:"insulinOnBoard,omitempty" bson:"insulinOnBoard,omitempty" valid:"omitempty,insulinvalue"`
	InsulinSensitivity *int     `json:"insulinSensitivity,omitempty" bson:"insulinSensitivity,omitempty" valid:"omitempty,insulinsensitivity"`
	BloodGlucoseInput  *float64 `json:"bgInput,omitempty" bson:"bgInput,omitempty" valid:"omitempty,bloodglucosevalue"`
	Units              *string  `json:"units" bson:"units" valid:"mmolmgunits"`
	types.Base         `bson:",inline"`
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
	High *float64 `json:"high" bson:"high" valid:"bloodglucosevalue"`
	Low  *float64 `json:"low" bson:"low" valid:"bloodglucosevalue"`
}

//Name is currently `wizard` for backwards compatatbilty but will be migrated to `calculator`
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
		"InsulinOnBoard":    validate.VaidationInfo{FieldName: types.BolusSubTypeField.Name, Message: types.BolusSubTypeField.Message},

		"SubType":  validate.VaidationInfo{FieldName: types.BolusSubTypeField.Name, Message: types.BolusSubTypeField.Message},
		"DeviceID": validate.VaidationInfo{FieldName: types.BaseDeviceIDField.Name, Message: types.BaseDeviceIDField.Message},

		"High": validate.VaidationInfo{FieldName: "high", Message: types.BloodGlucoseValueField.Message},
		"Low":  validate.VaidationInfo{FieldName: "low", Message: types.BloodGlucoseValueField.Message},

		"Net":          validate.VaidationInfo{FieldName: netField.Name, Message: netField.Message},
		"Correction":   validate.VaidationInfo{FieldName: correctionField.Name, Message: correctionField.Message},
		"Carbohydrate": validate.VaidationInfo{FieldName: carbField.Name, Message: carbField.Message},
	}
)

func buildRecommended(recommendedDatum types.Datum, errs validate.ErrorProcessing) *Recommended {
	return &Recommended{
		Carbohydrate: recommendedDatum.ToFloat64(carbField.Name, errs),
		Correction:   recommendedDatum.ToFloat64(correctionField.Name, errs),
		Net:          recommendedDatum.ToFloat64(netField.Name, errs),
	}
}

func buildBolus(bolusDatum types.Datum, errs validate.ErrorProcessing) *Bolus {
	bolusType := "bolus"
	return &Bolus{
		Type:     &bolusType,
		SubType:  bolusDatum.ToString(types.BolusSubTypeField.Name, errs),
		Time:     bolusDatum.ToString(types.TimeStringField.Name, errs),
		DeviceID: bolusDatum.ToString("deviceId", errs),
	}
}

func buildBloodGlucoseTarget(bgTargetDatum types.Datum, errs validate.ErrorProcessing) *BloodGlucoseTarget {
	return &BloodGlucoseTarget{
		High: bgTargetDatum.ToFloat64("high", errs),
		Low:  bgTargetDatum.ToFloat64("low", errs),
	}
}

func Build(datum types.Datum, errs validate.ErrorProcessing) *Event {

	var bolus *Bolus
	bolusDatum, ok := datum["bolus"].(map[string]interface{})
	if ok {
		bolus = buildBolus(bolusDatum, errs)
	}

	var bloodGlucoseTarget *BloodGlucoseTarget
	bloodGlucoseTargetDatum, ok := datum["bgTarget"].(map[string]interface{})
	if ok {
		bloodGlucoseTarget = buildBloodGlucoseTarget(bloodGlucoseTargetDatum, errs)
	}

	var recommended *Recommended
	recommendedDatum, ok := datum["recommended"].(map[string]interface{})
	if ok {
		recommended = buildRecommended(recommendedDatum, errs)
	}

	event := &Event{
		Recommended:        recommended,
		Bolus:              bolus,
		BloodGlucoseTarget: bloodGlucoseTarget,
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
