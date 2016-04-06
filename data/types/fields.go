package types

import "github.com/tidepool-org/platform/validate"

type (
	DatumFieldInformation struct {
		*DatumField
		Tag     validate.ValidationTag
		Message string
		Allowed
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

	Allowed map[string]bool
)
