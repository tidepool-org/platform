package normal

import (
	"github.com/tidepool-org/platform/data"
	dataTypesBolus "github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/structure"
)

const (
	SubType = "normal" // TODO: Rename Type to "bolus/normal"; remove SubType

	NormalMaximum = 100.0
	NormalMinimum = 0.0
)

type NormalFields struct {
	Normal         *float64 `json:"normal,omitempty" bson:"normal,omitempty"`
	NormalExpected *float64 `json:"expectedNormal,omitempty" bson:"expectedNormal,omitempty"`
}

func (n *NormalFields) Parse(parser structure.ObjectParser) {
	n.Normal = parser.Float64("normal")
	n.NormalExpected = parser.Float64("expectedNormal")
}

func (n *NormalFields) Validate(validator structure.Validator) {
	normalValidator := validator.Float64("normal", n.Normal).Exists()
	if n.NormalExpected != nil && structure.InRange(*n.NormalExpected, NormalMinimum, NormalMaximum) {
		normalValidator.InRange(NormalMinimum, *n.NormalExpected)
	} else {
		normalValidator.InRange(NormalMinimum, NormalMaximum)
	}
	validator.Float64("expectedNormal", n.NormalExpected).InRange(NormalMinimum, NormalMaximum)
}

type Normal struct {
	dataTypesBolus.Bolus `bson:",inline"`

	NormalFields `bson:",inline"`
}

func New() *Normal {
	return &Normal{
		Bolus: dataTypesBolus.New(SubType),
	}
}

func (n *Normal) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(n.Meta())
	}

	n.Bolus.Parse(parser)

	n.NormalFields.Parse(parser)
}

func (n *Normal) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(n.Meta())
	}

	n.Bolus.Validate(validator)

	if n.SubType != "" {
		validator.String("subType", &n.SubType).EqualTo(SubType)
	}

	n.NormalFields.Validate(validator)
}

func (n *Normal) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(n.Meta())
	}

	n.Bolus.Normalize(normalizer)
}
