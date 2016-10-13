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
	MmolL = "mmol/L"
	Mmoll = "mmol/l"

	MgdL = "mg/dL"
	Mgdl = "mg/dl"

	MmolLLowerLimit float64 = 0.0
	MmolLUpperLimit float64 = 55.0

	MgdLLowerLimit float64 = 0.0
	MgdLUpperLimit float64 = 1000.0

	MmolLToMgdLConversionFactor float64 = 18.01559
)

func ConvertValue(value float64, fromUnits string, toUnits string) float64 {
	if (fromUnits == MmolL || fromUnits == Mmoll) && (toUnits == MgdL || toUnits == Mgdl) {
		return value * MmolLToMgdLConversionFactor
	} else if (fromUnits == MgdL || fromUnits == Mgdl) && (toUnits == MmolL || toUnits == Mmoll) {
		return value / MmolLToMgdLConversionFactor
	}
	return value
}
