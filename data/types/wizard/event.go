package upload

import (
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

type Event struct {
	*Recommended        `json:"recommended,omitempty" bson:"recommended,omitempty" valid:"-"`
	CarbohydrateInput   *float64 `json:"carbInput,omitempty" bson:"carbInput,omitempty" valid:"omitempty,required"`
	InsulinOnBoard      *float64 `json:"insulinOnBoard,omitempty" bson:"insulinOnBoard,omitempty" valid:"omitempty,insulinvalue"`
	InsulinSensitivity  *float64 `json:"insulinSensitivity,omitempty" bson:"insulinSensitivity,omitempty" valid:"omitempty,insulinsensitivity"`
	*BloodGlucoseTarget `json:"bgTarget,omitempty" bson:"bgTarget,omitempty" valid:"-"`
	BloodGlucoseInput   *float64 `json:"bgInput,omitempty" bson:"bgInput,omitempty" valid:"omitempty,bloodvalue"`
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

const Name = "wizard"

var (
	carbohydrateInputField  = types.DatumField{Name: "carbInput"}
	insulinOnBoardField     = types.DatumField{Name: "insulinOnBoard"}
	insulinSensitivityField = types.DatumField{Name: "insulinSensitivity"}
	bloodGlucoseInputField  = types.DatumField{Name: "bgInput"}

	failureReasons = validate.FailureReasons{
		"Time":    validate.VaidationInfo{FieldName: types.TimeStringField.Name, Message: types.TimeStringField.Message},
		"High":    validate.VaidationInfo{FieldName: "high", Message: types.BloodGlucoseValueField.Message},
		"Low":     validate.VaidationInfo{FieldName: "low", Message: types.BloodGlucoseValueField.Message},
		"Units":   validate.VaidationInfo{FieldName: types.MmolOrMgUnitsField.Name, Message: types.MmolOrMgUnitsField.Message},
		"SubType": validate.VaidationInfo{FieldName: types.BolusSubTypeField.Name, Message: types.BolusSubTypeField.Message},
	}
)

func buildRecommended(datum types.Datum, errs validate.ErrorProcessing) *Recommended {
	recommendedDatum, ok := datum["recommended"].(types.Datum)
	if ok {
		return &Recommended{
			Carbohydrate: recommendedDatum.ToFloat64("carb", errs),
			Correction:   recommendedDatum.ToFloat64("correction", errs),
			Net:          recommendedDatum.ToFloat64("net", errs),
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
		CarbohydrateInput:  datum.ToFloat64(carbohydrateInputField.Name, errs),
		InsulinOnBoard:     datum.ToFloat64(insulinOnBoardField.Name, errs),
		InsulinSensitivity: datum.ToFloat64(insulinSensitivityField.Name, errs),
		BloodGlucoseInput:  datum.ToFloat64(bloodGlucoseInputField.Name, errs),
		Units:              datum.ToString(types.MmolOrMgUnitsField.Name, errs),
		Base:               types.BuildBase(datum, errs),
	}

	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(event, errs)

	return event
}
