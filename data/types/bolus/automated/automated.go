package automated

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/structure"
)

const (
	SubType = "automated" // TODO: Rename Type to "bolus/automated"; remove SubType

	NormalMaximum = 100.0
	NormalMinimum = 0.0
)

type Automated struct {
	bolus.Bolus `bson:",inline"`

	Normal         *float64 `json:"normal,omitempty" bson:"normal,omitempty"`
	NormalExpected *float64 `json:"expectedNormal,omitempty" bson:"expectedNormal,omitempty"`
}

func New() *Automated {
	return &Automated{
		Bolus: bolus.New(SubType),
	}
}

func (a *Automated) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(a.Meta())
	}

	a.Bolus.Parse(parser)

	a.Normal = parser.Float64("normal")
	a.NormalExpected = parser.Float64("expectedNormal")
}

func (a *Automated) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(a.Meta())
	}

	a.Bolus.Validate(validator)

	if a.SubType != "" {
		validator.String("subType", &a.SubType).EqualTo(SubType)
	}

	validator.Float64("normal", a.Normal).Exists().InRange(NormalMinimum, NormalMaximum)
	normalExpectedValidator := validator.Float64("expectedNormal", a.NormalExpected)
	if a.Normal != nil && *a.Normal >= NormalMinimum && *a.Normal <= NormalMaximum {
		if *a.Normal == NormalMinimum {
			normalExpectedValidator.Exists()
		}
		normalExpectedValidator.InRange(*a.Normal, NormalMaximum)
	} else {
		normalExpectedValidator.InRange(NormalMinimum, NormalMaximum)
	}
}

func (a *Automated) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(a.Meta())
	}

	a.Bolus.Normalize(normalizer)
}
