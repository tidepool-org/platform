package normal

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/structure"
)

const (
	SubType = "normal" // TODO: Rename Type to "bolus/normal"; remove SubType

	BiphasicSubType = "biphasic"
	NormalMaximum   = 100.0
	NormalMinimum   = 0.0
)

func SubTypes() []string {
	return []string{
		SubType,
		BiphasicSubType,
	}
}

type Normal struct {
	bolus.Bolus `bson:",inline"`

	Normal         *float64 `json:"normal,omitempty" bson:"normal,omitempty"`
	NormalExpected *float64 `json:"expectedNormal,omitempty" bson:"expectedNormal,omitempty"`
}

func New() *Normal {
	return &Normal{
		Bolus: bolus.New(SubType),
	}
}

func NewWithSubType(subtype string) Normal {
	return Normal{
		Bolus: bolus.New(subtype),
	}
}

func (n *Normal) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(n.Meta())
	}

	n.Bolus.Parse(parser)

	n.Normal = parser.Float64("normal")
	n.NormalExpected = parser.Float64("expectedNormal")
}

func (n *Normal) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(n.Meta())
	}

	n.Bolus.Validate(validator)

	if n.SubType != "" {
		validator.String("subType", &n.SubType).OneOf(SubTypes()...)
	}

	validator.Float64("normal", n.Normal).Exists().InRange(NormalMinimum, NormalMaximum)
	normalExpectedValidator := validator.Float64("expectedNormal", n.NormalExpected)
	// Temporary workaround for PT-423
	normalExpectedValidator.InRange(NormalMinimum, NormalMaximum)
	// if n.Normal != nil && *n.Normal >= NormalMinimum && *n.Normal <= NormalMaximum {
	// 	if *n.Normal == NormalMinimum {
	// 		normalExpectedValidator.Exists()
	// 	}
	// 	normalExpectedValidator.InRange(*n.Normal, NormalMaximum)
	// } else {
	//  normalExpectedValidator.InRange(NormalMinimum, NormalMaximum)
	// }
}

// IsValid returns true if there is no error in the validator
func (n *Normal) IsValid(validator structure.Validator) bool {
	return !(validator.HasError())
}

func (n *Normal) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(n.Meta())
	}

	n.Bolus.Normalize(normalizer)
}
