package normal

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/structure"
)

const (
	SubType = "normal" // TODO: Rename Type to "bolus/normal"; remove SubType

	NormalMaximum = 100.0
	NormalMinimum = 0.0
)

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

func (n *Normal) Parse(parser data.ObjectParser) error {
	if err := n.Bolus.Parse(parser); err != nil {
		return err
	}

	n.Normal = parser.ParseFloat("normal")
	n.NormalExpected = parser.ParseFloat("expectedNormal")

	return nil
}

func (n *Normal) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(n.Meta())
	}

	n.Bolus.Validate(validator)

	if n.SubType != "" {
		validator.String("subType", &n.SubType).EqualTo(SubType)
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

func (n *Normal) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(n.Meta())
	}

	n.Bolus.Normalize(normalizer)
}
