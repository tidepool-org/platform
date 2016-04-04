package data

import "github.com/tidepool-org/platform/validate"

type (
	TypesDatumField struct {
		*DatumField
		Tag     validate.ValidationTag
		Message string
		AllowedTypes
	}

	FloatDatumField struct {
		*DatumField
		Tag     validate.ValidationTag
		Message string
		*AllowedFloatRange
	}

	IntDatumField struct {
		*DatumField
		Tag     validate.ValidationTag
		Message string
		*AllowedIntRange
	}

	DatumField struct {
		Name string
	}

	AllowedFloatRange struct {
		UpperLimit float64
		LowerLimit float64
	}

	AllowedIntRange struct {
		UpperLimit int
		LowerLimit int
	}

	AllowedTypes map[string]bool
)
