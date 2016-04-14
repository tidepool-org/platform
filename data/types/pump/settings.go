package pump

import (
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

type Settings struct {
	*Units               `json:"units,omitempty" bson:"units,omitempty"`
	BasalSchedules       map[string]*BasalSchedule `json:"basalSchedules,omitempty" bson:"basalSchedules,omitempty"`
	CarbohydrateRatios   []*CarbohydrateRatio      `json:"carbRatio,omitempty" bson:"carbRatio,omitempty"`
	InsulinSensitivities []*InsulinSensitivity     `json:"insulinSensitivity,omitempty" bson:"insulinSensitivity,omitempty"`
	BloodGlucoseTargets  []*BloodGlucoseTarget     `json:"bgTarget,omitempty" bson:"bgTarget,omitempty"`

	ActiveSchedule *string `json:"activeSchedule" bson:"activeSchedule" valid:"required"`
	types.Base     `bson:",inline"`
}

type Units struct {
	Carbohydrate *string `json:"carb" bson:"carb" valid:"required"`
	BloodGlucose *string `json:"bg" bson:"bg" valid:"mmolmgunits"`
}

type BloodGlucoseTarget struct {
	Low   *float64 `json:"low" bson:"low" valid:"bloodglucosevalue"`
	High  *float64 `json:"high" bson:"high" valid:"bloodglucosevalue"`
	Start *int     `json:"start" bson:"start" valid:"required"`
}

type CarbohydrateRatio struct {
	Amount *float64 `json:"amount" bson:"amount" valid:"required"`
	Start  *int     `json:"start" bson:"start" valid:"required"`
}

type InsulinSensitivity struct {
	Amount *float64 `json:"amount" bson:"amount" valid:"required"`
	Start  *int     `json:"start" bson:"start" valid:"required"`
}

type BasalSchedule struct {
	Rate  *float64 `json:"rate" bson:"rate" valid:"required"`
	Start *int     `json:"start" bson:"start" valid:"required"`
}

const Name = "pumpSettings"

var (
	activeScheduleField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "activeSchedule"},
		Tag:        "required",
		Message:    "This is a required field",
	}

	carbohydrateUnitsField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "carb"},
		Tag:        "carbunits",
		Message:    "This is a required field",
	}

	bloodGlucoseUnitsField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "bg"},
		Tag:        types.MmolOrMgUnitsField.Tag,
		Message:    types.MmolOrMgUnitsField.Message,
	}

	amountField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "amount"},
		Tag:        "required",
		Message:    "This is a required field",
	}

	rateField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "rate"},
		Tag:        "required",
		Message:    "This is a required field",
	}

	startField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "start"},
		Tag:        "required",
		Message:    "This is a required field",
	}

	failureReasons = validate.FailureReasons{
		"ActiveSchedule": validate.VaidationInfo{FieldName: activeScheduleField.Name, Message: activeScheduleField.Message},

		"Carbohydrate": validate.VaidationInfo{FieldName: carbohydrateUnitsField.Name, Message: carbohydrateUnitsField.Message},
		"BloodGlucose": validate.VaidationInfo{FieldName: bloodGlucoseUnitsField.Name, Message: bloodGlucoseUnitsField.Message},

		"High": validate.VaidationInfo{FieldName: "high", Message: types.BloodGlucoseValueField.Message},
		"Low":  validate.VaidationInfo{FieldName: "low", Message: types.BloodGlucoseValueField.Message},

		"Rate":   validate.VaidationInfo{FieldName: rateField.Name, Message: rateField.Message},
		"Start":  validate.VaidationInfo{FieldName: startField.Name, Message: startField.Message},
		"Amount": validate.VaidationInfo{FieldName: amountField.Name, Message: amountField.Message},
	}
)

func buildUnits(unitsDatum types.Datum, errs validate.ErrorProcessing) *Units {
	return &Units{
		Carbohydrate: unitsDatum.ToString(carbohydrateUnitsField.Name, errs),
		BloodGlucose: unitsDatum.ToString(bloodGlucoseUnitsField.Name, errs),
	}
}

func buildBloodGlucoseTargets(targetsDatum types.Datum, errs validate.ErrorProcessing) []*BloodGlucoseTarget {
	return []*BloodGlucoseTarget{}
}

func buildInsulinSensitivities(sensitivitiesDatum types.Datum, errs validate.ErrorProcessing) []*InsulinSensitivity {
	return []*InsulinSensitivity{}
}

func buildCarbohydrateRatios(sensitivitiesDatum types.Datum, errs validate.ErrorProcessing) []*CarbohydrateRatio {
	return []*CarbohydrateRatio{}
}

func Build(datum types.Datum, errs validate.ErrorProcessing) *Settings {

	var units *Units
	unitsDatum, ok := datum["units"].(map[string]interface{})
	if ok {
		units = buildUnits(unitsDatum, errs)
	}

	var targets []*BloodGlucoseTarget
	targetsDatum, ok := datum["bgTarget"].(map[string]interface{})
	if ok {
		targets = buildBloodGlucoseTargets(targetsDatum, errs)
	}

	var insulinSensitivities []*InsulinSensitivity
	sensitivitiesDatum, ok := datum["insulinSensitivity"].(map[string]interface{})
	if ok {
		insulinSensitivities = buildInsulinSensitivities(sensitivitiesDatum, errs)
	}

	var carbohydrateRatios []*CarbohydrateRatio
	carbRatioDatum, ok := datum["carbRatio"].(map[string]interface{})
	if ok {
		carbohydrateRatios = buildCarbohydrateRatios(carbRatioDatum, errs)
	}

	settings := &Settings{
		Units:                units,
		BloodGlucoseTargets:  targets,
		InsulinSensitivities: insulinSensitivities,
		CarbohydrateRatios:   carbohydrateRatios,
		ActiveSchedule:       datum.ToString(activeScheduleField.Name, errs),
		Base:                 types.BuildBase(datum, errs),
	}

	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(settings, errs)

	return settings
}
