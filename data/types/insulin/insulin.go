package insulin

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/structure"
)

const (
	Type = "insulin"

	SiteLengthMaximum = 100
)

type Insulin struct {
	types.Base `bson:",inline"`

	Dose        *Dose        `json:"dose,omitempty" bson:"dose,omitempty"`
	Formulation *Formulation `json:"formulation,omitempty" bson:"formulation,omitempty"`
	Site        *string      `json:"site,omitempty" bson:"site,omitempty"`
}

func New() *Insulin {
	return &Insulin{
		Base: types.New(Type),
	}
}

func (i *Insulin) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(i.Meta())
	}

	i.Base.Parse(parser)

	i.Dose = ParseDose(parser.WithReferenceObjectParser("dose"))
	i.Formulation = ParseFormulation(parser.WithReferenceObjectParser("formulation"))
	i.Site = parser.String("site")
}

func (i *Insulin) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(i.Meta())
	}

	i.Base.Validate(validator)

	if i.Type != "" {
		validator.String("type", &i.Type).EqualTo(Type)
	}

	if i.Dose != nil {
		i.Dose.Validate(validator.WithReference("dose"))
	}
	if i.Formulation != nil {
		i.Formulation.Validate(validator.WithReference("formulation"))
	}
	validator.String("site", i.Site).NotEmpty().LengthLessThanOrEqualTo(SiteLengthMaximum)
}

func (i *Insulin) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(i.Meta())
	}

	i.Base.Normalize(normalizer)

	if i.Dose != nil {
		i.Dose.Normalize(normalizer.WithReference("dose"))
	}
	if i.Formulation != nil {
		i.Formulation.Normalize(normalizer.WithReference("formulation"))
	}
}

func (i *Insulin) LegacyIdentityFields() ([]string, error) {
	return types.GetLegacyIDFields(
		types.LegacyIDField{Name: "type", Value: &i.Type},
		types.LegacyIDField{Name: "device id", Value: i.DeviceID},
		types.GetLegacyTimeField(i.Time),
	)
}
