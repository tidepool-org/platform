package common

var (
	Mmoll = "mmol/l"
	MmolL = "mmol/L"
	Mgdl  = "mg/dl"
	MgdL  = "mg/dL"

	MmolLFromValue = 0.0
	MmolLToValue   = 55.0
	MgdLFromValue  = 0.0
	MgdLToValue    = MmolLToValue * MgdlToMmolConversion
)

const MgdlToMmolConversion = 18.01559
