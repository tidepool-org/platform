package cgm

import (
	"reflect"

	validator "gopkg.in/bluesuncorp/validator.v8"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

func init() {
	types.GetPlatformValidator().RegisterValidation(snoozeField.Tag, SnoozeValidator)
	types.GetPlatformValidator().RegisterValidation(rateField.Tag, RateValidator)
	types.GetPlatformValidator().RegisterValidation(levelField.Tag, LevelValidator)
}

type Settings struct {
	Units         *string `json:"units" bson:"units" valid:"mmolmgunits"`
	TransmitterID *string `json:"transmitterId" bson:"transmitterId" valid:"required"`

	High               Alert `json:"highAlerts" bson:"highAlerts"`
	Low                Alert `json:"lowAlerts" bson:"lowAlerts"`
	*OutOfRangeAlert   `json:"outOfRangeAlerts,omitempty" bson:"outOfRangeAlerts,omitempty"`
	ChangeOfRateAlerts map[string]ChangeOfRateAlert `json:"rateOfChangeAlerts" bson:"rateOfChangeAlerts"`

	types.Base `bson:",inline"`
}

type Alert struct {
	Enabled *bool    `json:"enabled" bson:"enabled" valid:"exists"`
	Level   *float64 `json:"level" bson:"level" valid:"cgmsettingslevel"`
	Snooze  *int     `json:"snooze" bson:"snooze" valid:"cgmsettingssnooze"`
}

type OutOfRangeAlert struct {
	Enabled *bool `json:"enabled" bson:"enabled" valid:"exists"`
	Snooze  *int  `json:"snooze" bson:"snooze" valid:"cgmsettingssnooze"`
}

type ChangeOfRateAlert struct {
	Enabled *bool    `json:"enabled" bson:"enabled" valid:"exists"`
	Rate    *float64 `json:"rate" bson:"rate" valid:"cgmsettingsrate"`
}

const Name = "cgmSettings"

var (
	transmitterIDField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "transmitterId"},
		Tag:        "required",
		Message:    "This is a required field",
	}

	enabledField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "enabled"},
		Tag:        "exists",
		Message:    "This is a required field",
	}

	levelField = types.FloatDatumField{
		DatumField:        &types.DatumField{Name: "level"},
		Tag:               "cgmsettingslevel",
		Message:           "Must be >= 3.0 and <= 15.0",
		AllowedFloatRange: &types.AllowedFloatRange{LowerLimit: 3.0, UpperLimit: 1000.0},
	}

	rateField = types.FloatDatumField{
		DatumField:        &types.DatumField{Name: "rate"},
		Tag:               "cgmsettingsrate",
		Message:           "Must be >= -1.0 and <= 1.0",
		AllowedFloatRange: &types.AllowedFloatRange{LowerLimit: -1.0, UpperLimit: 1.0},
	}

	snoozeField = types.IntDatumField{
		DatumField:      &types.DatumField{Name: "snooze"},
		Tag:             "cgmsettingssnooze",
		Message:         "Must be >= 0 and <= 432000000",
		AllowedIntRange: &types.AllowedIntRange{LowerLimit: 0, UpperLimit: 432000000},
	}

	failureReasons = validate.FailureReasons{
		"High.Enabled": validate.ValidationInfo{
			FieldName: "highAlerts/" + enabledField.Name,
			Message:   enabledField.Message,
		},
		"Low.Enabled": validate.ValidationInfo{
			FieldName: "lowAlerts/" + enabledField.Name,
			Message:   enabledField.Message,
		},
		"OutOfRangeAlert.Enabled": validate.ValidationInfo{
			FieldName: "outOfRangeAlerts/" + enabledField.Name,
			Message:   enabledField.Message,
		},
		"Low.Level": validate.ValidationInfo{
			FieldName: "lowAlerts/" + levelField.Name,
			Message:   levelField.Message,
		},
		"High.Level": validate.ValidationInfo{
			FieldName: "highAlerts/" + levelField.Name,
			Message:   levelField.Message,
		},
		"Low.Snooze": validate.ValidationInfo{
			FieldName: "lowAlerts/" + snoozeField.Name,
			Message:   snoozeField.Message,
		},
		"High.Snooze": validate.ValidationInfo{
			FieldName: "highAlerts/" + snoozeField.Name,
			Message:   snoozeField.Message,
		},
		"OutOfRangeAlert.Snooze": validate.ValidationInfo{
			FieldName: "outOfRangeAlerts/" + snoozeField.Name,
			Message:   snoozeField.Message,
		},
		"Rate": validate.ValidationInfo{
			FieldName: rateField.Name,
			Message:   rateField.Message,
		},
		"TransmitterID": validate.ValidationInfo{
			FieldName: transmitterIDField.Name,
			Message:   transmitterIDField.Message,
		},
		"Units": validate.ValidationInfo{
			FieldName: types.MmolOrMgUnitsField.Name,
			Message:   types.MmolOrMgUnitsField.Message,
		},
	}
)

func buildAlert(alertDatum types.Datum, errs validate.ErrorProcessing) Alert {
	return Alert{
		Enabled: alertDatum.ToBool(enabledField.Name, errs),
		Level:   alertDatum.ToFloat64(levelField.Name, errs),
		Snooze:  alertDatum.ToInt(snoozeField.Name, errs),
	}
}

func buildOutOfRangeAlert(changeOfRateDatum types.Datum, errs validate.ErrorProcessing) *OutOfRangeAlert {
	return &OutOfRangeAlert{
		Enabled: changeOfRateDatum.ToBool(enabledField.Name, errs),
		Snooze:  changeOfRateDatum.ToInt(snoozeField.Name, errs),
	}
}

func buildChangeOfRateAlerts(changeOfRateAlertsDatum []map[string]interface{}, errs validate.ErrorProcessing) map[string]ChangeOfRateAlert {

	var changes map[string]ChangeOfRateAlert

	for _, val := range changeOfRateAlertsDatum {

		datum := types.Datum(val)

		change := ChangeOfRateAlert{
			Enabled: datum.ToBool(enabledField.Name, errs),
			Rate:    datum.ToFloat64(rateField.Name, errs),
		}

		changes["todo"] = change
	}

	return changes
}

func Build(datum types.Datum, errs validate.ErrorProcessing) *Settings {

	var high Alert
	highDatum, ok := datum["highAlerts"].(map[string]interface{})
	if ok {
		high = buildAlert(highDatum, errs)
	}

	var low Alert
	lowDatum, ok := datum["lowAlerts"].(map[string]interface{})
	if ok {
		low = buildAlert(lowDatum, errs)
	}

	var outOfRangeAlert *OutOfRangeAlert
	outOfRangeDatum, ok := datum["outOfRangeAlerts"].(map[string]interface{})
	if ok {
		outOfRangeAlert = buildOutOfRangeAlert(outOfRangeDatum, errs)
	} else {
		outOfRangeAlert = nil
	}

	var changeOfRateAlerts map[string]ChangeOfRateAlert
	changeOfRateAlertsDatum, ok := datum["rateOfChangeAlerts"].([]map[string]interface{})
	if ok {
		changeOfRateAlerts = buildChangeOfRateAlerts(changeOfRateAlertsDatum, errs)
	}

	settings := &Settings{
		Units:              datum.ToString(types.MmolOrMgUnitsField.Name, errs),
		TransmitterID:      datum.ToString(transmitterIDField.Name, errs),
		High:               high,
		Low:                low,
		ChangeOfRateAlerts: changeOfRateAlerts,
		OutOfRangeAlert:    outOfRangeAlert,
		Base:               types.BuildBase(datum, errs),
	}

	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(settings, errs)

	return settings
}

func SnoozeValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	snooze, ok := field.Interface().(int)
	if !ok {
		return false
	}
	return snooze >= snoozeField.LowerLimit && snooze <= snoozeField.UpperLimit
}

func RateValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	rate, ok := field.Interface().(float64)
	if !ok {
		return false
	}
	return rate >= rateField.LowerLimit && rate <= rateField.UpperLimit
}

func LevelValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	level, ok := field.Interface().(float64)
	if !ok {
		return false
	}
	return level >= levelField.LowerLimit && level <= levelField.UpperLimit
}
