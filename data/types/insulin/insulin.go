package insulin

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
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

func (i *Insulin) Parse(parser data.ObjectParser) error {
	parser.SetMeta(i.Meta())

	if err := i.Base.Parse(parser); err != nil {
		return err
	}

	i.Dose = ParseDose(parser.NewChildObjectParser("dose"))
	i.Formulation = ParseFormulation(parser.NewChildObjectParser("formulation"))
	i.Site = parser.ParseString("site")

	return nil
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
	} else {
		validator.WithReference("dose").ReportError(structureValidator.ErrorValueNotExists())
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
