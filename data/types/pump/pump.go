package pump

import (
	"reflect"

	validator "gopkg.in/bluesuncorp/validator.v8"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

func init() {
	types.GetPlatformValidator().RegisterValidation(startField.Tag, StartValidator)
}

type Settings struct {
	*Units               `json:"units,omitempty" bson:"units,omitempty"`
	BasalSchedules       map[string][]*BasalSchedule `json:"basalSchedules,omitempty" bson:"basalSchedules,omitempty"`
	CarbohydrateRatios   `json:"carbRatio,omitempty" bson:"carbRatio,omitempty"`
	InsulinSensitivities []*InsulinSensitivity `json:"insulinSensitivity,omitempty" bson:"insulinSensitivity,omitempty"`
	BloodGlucoseTargets  []*BloodGlucoseTarget `json:"bgTarget,omitempty" bson:"bgTarget,omitempty"`

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
	Start *int     `json:"start" bson:"start" valid:"startrange"`
}

type CarbohydrateRatio struct {
	Amount *float64 `json:"amount" bson:"amount" valid:"required"`
	Start  *int     `json:"start" bson:"start" valid:"startrange"`
}

type CarbohydrateRatios []CarbohydrateRatio

type InsulinSensitivity struct {
	Amount *float64 `json:"amount" bson:"amount" valid:"required"`
	Start  *int     `json:"start" bson:"start" valid:"startrange"`
}

type BasalSchedule struct {
	Rate  *float64 `json:"rate" bson:"rate" valid:"required"`
	Start *int     `json:"start" bson:"start" valid:"startrange"`
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

	startField = types.IntDatumField{
		DatumField:      &types.DatumField{Name: "start"},
		Tag:             "startrange",
		Message:         "Needs to be in the range of >= 0 and < 86400000",
		AllowedIntRange: &types.AllowedIntRange{LowerLimit: 0, UpperLimit: 86400000},
	}

	failureReasons = validate.FailureReasons{
		"ActiveSchedule": validate.ValidationInfo{
			FieldName: activeScheduleField.Name,
			Message:   activeScheduleField.Message,
		},
		"Units.Carbohydrate": validate.ValidationInfo{
			FieldName: "units/" + carbohydrateUnitsField.Name,
			Message:   carbohydrateUnitsField.Message,
		},
		"Units.BloodGlucose": validate.ValidationInfo{
			FieldName: "units/" + bloodGlucoseUnitsField.Name,
			Message:   bloodGlucoseUnitsField.Message,
		},
		"BloodGlucoseTarget.Start": validate.ValidationInfo{
			FieldName: "bgTarget/" + startField.Name,
			Message:   startField.Message,
		},
		"BloodGlucoseTarget.High": validate.ValidationInfo{
			FieldName: "bgTarget/high",
			Message:   types.BloodGlucoseValueField.Message,
		},
		"BloodGlucoseTarget.Low": validate.ValidationInfo{
			FieldName: "bgTarget/low",
			Message:   types.BloodGlucoseValueField.Message,
		},
		"CarbohydrateRatio.Start": validate.ValidationInfo{
			FieldName: "carbRatio/" + startField.Name,
			Message:   startField.Message,
		},
		"CarbohydrateRatio.Amount": validate.ValidationInfo{
			FieldName: "carbRatio/" + amountField.Name,
			Message:   amountField.Message,
		},
		"InsulinSensitivity.Start": validate.ValidationInfo{
			FieldName: "insulinSensitivity/" + startField.Name,
			Message:   startField.Message,
		},
		"InsulinSensitivity.Amount": validate.ValidationInfo{
			FieldName: "insulinSensitivity/" + amountField.Name,
			Message:   amountField.Message,
		},
		"BasalSchedule.Start": validate.ValidationInfo{
			FieldName: "basalSchedules/" + startField.Name,
			Message:   startField.Message,
		},
		"BasalSchedule.Rate": validate.ValidationInfo{
			FieldName: "basalSchedules/" + rateField.Name,
			Message:   rateField.Message,
		},
	}
)

func buildUnits(unitsDatum types.Datum, errs validate.ErrorProcessing) *Units {
	return &Units{
		Carbohydrate: unitsDatum.ToString(carbohydrateUnitsField.Name, errs),
		BloodGlucose: unitsDatum.ToString(bloodGlucoseUnitsField.Name, errs),
	}
}

func buildBloodGlucoseTargets(targetsDatum []map[string]interface{}, errs validate.ErrorProcessing) []*BloodGlucoseTarget {

	var targets []*BloodGlucoseTarget

	for _, val := range targetsDatum {

		datum := types.Datum(val)

		target := &BloodGlucoseTarget{
			Low:   datum.ToFloat64("low", errs),
			High:  datum.ToFloat64("high", errs),
			Start: datum.ToInt(startField.Name, errs),
		}

		//types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(target, errs)

		targets = append(targets, target)

	}

	return targets
}

func buildInsulinSensitivities(sensitivitiesDatum []map[string]interface{}, errs validate.ErrorProcessing) []*InsulinSensitivity {
	var sensitivities []*InsulinSensitivity

	for _, val := range sensitivitiesDatum {

		datum := types.Datum(val)

		sensitivity := &InsulinSensitivity{
			Amount: datum.ToFloat64(amountField.Name, errs),
			Start:  datum.ToInt(startField.Name, errs),
		}

		//types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(sensitivity, errs)

		sensitivities = append(sensitivities, sensitivity)

	}

	return sensitivities
}

func buildCarbohydrateRatios(carbohydrateRatiosDatum []map[string]interface{}, errs validate.ErrorProcessing) CarbohydrateRatios {

	ratios := make(CarbohydrateRatios, 0)

	for _, val := range carbohydrateRatiosDatum {

		datum := types.Datum(val)

		ratio := CarbohydrateRatio{
			Amount: datum.ToFloat64(amountField.Name, errs),
			Start:  datum.ToInt(startField.Name, errs),
		}

		//log.Println("## Amount failure ", failureReasons["CarbohydrateRatio.Amount"])
		//log.Println("## Start failure ", failureReasons["CarbohydrateRatio.Start"])
		//types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(ratio, errs)
		//log.Println("## errs ", errs)
		//log.Printf("## built %#v ", ratio)

		ratios = append(ratios, ratio)
	}
	return ratios
}

func buildBasalSchedules(schedulesDatum map[string][]map[string]interface{}, errs validate.ErrorProcessing) map[string][]*BasalSchedule {

	namedSchedules := make(map[string][]*BasalSchedule, 0)

	for key, vals := range schedulesDatum {

		var schedules []*BasalSchedule

		for i := range vals {
			datum := types.Datum(vals[i])

			schedule := &BasalSchedule{
				Rate:  datum.ToFloat64(rateField.Name, errs),
				Start: datum.ToInt(startField.Name, errs),
			}

			//types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(schedule, errs)

			schedules = append(schedules, schedule)

		}

		namedSchedules[key] = schedules
	}

	return namedSchedules

}

func Build(datum types.Datum, errs validate.ErrorProcessing) *Settings {

	var units *Units
	unitsDatum, ok := datum["units"].(map[string]interface{})
	if ok {
		units = buildUnits(unitsDatum, errs)
	}

	var targets []*BloodGlucoseTarget
	targetsDatum, ok := datum["bgTarget"].([]map[string]interface{})
	if ok {
		targets = buildBloodGlucoseTargets(targetsDatum, errs)
	}

	var insulinSensitivities []*InsulinSensitivity
	sensitivitiesDatum, ok := datum["insulinSensitivity"].([]map[string]interface{})
	if ok {
		insulinSensitivities = buildInsulinSensitivities(sensitivitiesDatum, errs)
	}

	var carbohydrateRatios CarbohydrateRatios
	carbRatioDatum, ok := datum["carbRatio"].([]map[string]interface{})
	if ok {
		carbohydrateRatios = buildCarbohydrateRatios(carbRatioDatum, errs)
	}

	var basalSchedules map[string][]*BasalSchedule
	basalSchedulesDatum, ok := datum["basalSchedules"].(map[string][]map[string]interface{})
	if ok {
		basalSchedules = buildBasalSchedules(basalSchedulesDatum, errs)
	}

	settings := &Settings{
		Units:                units,
		BloodGlucoseTargets:  targets,
		InsulinSensitivities: insulinSensitivities,
		CarbohydrateRatios:   carbohydrateRatios,
		BasalSchedules:       basalSchedules,
		ActiveSchedule:       datum.ToString(activeScheduleField.Name, errs),
		Base:                 types.BuildBase(datum, errs),
	}

	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(settings, errs)

	return settings
}

func StartValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	start, ok := field.Interface().(int)
	if !ok {
		return false
	}
	return start >= startField.LowerLimit && start < startField.UpperLimit
}
