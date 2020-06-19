package pen

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/structure"
)

const (
	SubType = "pen"

	NormalMaximum = 100.0
	NormalMinimum = 0.0
)

type Pen struct {
	bolus.Bolus `bson:",inline"`

	Normal *float64 `json:"normal,omitempty" bson:"normal,omitempty"`
}

func New() *Pen {
	return &Pen{
		Bolus: bolus.New(SubType),
	}
}

func (n *Pen) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(n.Meta())
	}

	n.Bolus.Parse(parser)

	n.Normal = parser.Float64("normal")
}

func (n *Pen) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(n.Meta())
	}

	n.Bolus.Validate(validator)

	if n.SubType != "" {
		validator.String("subType", &n.SubType).EqualTo(SubType)
	}

	validator.Float64("normal", n.Normal).Exists().InRange(NormalMinimum, NormalMaximum)
}

// IsValid returns true if there is no error in the validator
func (n *Pen) IsValid(validator structure.Validator) bool {
	return !(validator.HasError())
}

func (n *Pen) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(n.Meta())
	}

	n.Bolus.Normalize(normalizer)
}
