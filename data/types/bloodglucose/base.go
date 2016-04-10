package bloodglucose

import (
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

var (
	mmol = "mmol/L"
	mg   = "mg/dL"

	failureReasons = validate.ErrorReasons{
		types.MmolOrMgUnitsField.Tag: types.MmolOrMgUnitsField.Message,
		types.BloodValueField.Tag:    types.BloodValueField.Message,
		isigField.Tag:                isigField.Message,
	}
)

func normalizeUnitName(unitsName *string) *string {
	switch *unitsName {
	case mmol, "mmol/l":
		return &mmol
	case mg, "mg/dl":
		return &mg
	}
	return unitsName
}

func convertMgToMmol(mgValue *float64, units *string) *float64 {

	switch *normalizeUnitName(units) {
	case mg:
		converted := *mgValue / 18.01559
		return &converted
	default:
		return mgValue
	}
}
