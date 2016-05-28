package bloodglucose

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

const (
	Mmoll = "mmol/l"
	MmolL = "mmol/L"
	Mgdl  = "mg/dl"
	MgdL  = "mg/dL"

	MmolLToMgdLConversionFactor float64 = 18.01559

	MmolLLowerLimit float64 = 0.0
	MmolLUpperLimit float64 = 55.0
	MgdLLowerLimit  float64 = 0.0
	MgdLUpperLimit  float64 = 1000.0
)

func ConvertValue(value float64, fromUnits string, toUnits string) float64 {
	if (fromUnits == Mmoll || fromUnits == MmolL) && (toUnits == Mgdl || toUnits == MgdL) {
		return value * MmolLToMgdLConversionFactor
	} else if (fromUnits == Mgdl || fromUnits == MgdL) && (toUnits == Mmoll || toUnits == MmolL) {
		return value / MmolLToMgdLConversionFactor
	}
	return value
}
