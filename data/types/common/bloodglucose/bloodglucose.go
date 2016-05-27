package bloodglucose

var (
	Mmoll = "mmol/l"
	MmolL = "mmol/L"
	Mgdl  = "mg/dl"
	MgdL  = "mg/dL"

	MmolLFromValue = 0.0
	MmolLToValue   = 55.0
	MgdLFromValue  = 0.0
	MgdLToValue    = MmolLToValue * MgdlToMmolConversion

	AllowedUnits = []string{Mmoll, MmolL, Mgdl, MgdL}
)

const MgdlToMmolConversion = 18.01559

func AllowedMmolLRange() (float64, float64) {
	return MmolLFromValue, MmolLToValue
}

func AllowedMgdLRange() (float64, float64) {
	return MgdLFromValue, MgdLToValue
}
