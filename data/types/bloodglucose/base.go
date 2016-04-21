package bloodglucose

import (
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

var (
	mmol = "mmol/L"
	mg   = "mg/dL"

	failureReasons = validate.FailureReasons{
		"Units": validate.ValidationInfo{FieldName: types.MmolOrMgUnitsField.Name, Message: types.MmolOrMgUnitsField.Message},
		"Value": validate.ValidationInfo{FieldName: types.BloodGlucoseValueField.Name, Message: types.BloodGlucoseValueField.Message},
		"Isig":  validate.ValidationInfo{FieldName: isigField.Name, Message: isigField.Message},
	}
)

func NormalizeUnitName(unitsName *string) *string {
	if unitsName == nil {
		return unitsName
	}

	switch *unitsName {
	case mmol, "mmol/l":
		return &mmol
	case mg, "mg/dl":
		return &mg
	default:
		return unitsName
	}
}

func ConvertMgToMmol(mgValue *float64, units *string) *float64 {

	if mgValue == nil || units == nil {
		return mgValue
	}

	switch *NormalizeUnitName(units) {
	case mg:
		converted := *mgValue / 18.01559
		return &converted
	default:
		return mgValue
	}
}
