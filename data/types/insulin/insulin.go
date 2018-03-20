package insulin

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	Type = "insulin"

	ActingTypeIntermediate = "intermediate"
	ActingTypeLong         = "long"
	ActingTypeRapid        = "rapid"
	ActingTypeShort        = "short"
	BrandLengthMaximum     = 100
	ConcentrationMaximum   = 10000
	ConcentrationMinimum   = 1
	NameLengthMaximum      = 100
	SiteLengthMaximum      = 100
)

func ActingTypes() []string {
	return []string{
		ActingTypeIntermediate,
		ActingTypeLong,
		ActingTypeRapid,
		ActingTypeShort,
	}
}

type Insulin struct {
	types.Base `bson:",inline"`

	ActingType    *string `json:"actingType,omitempty" bson:"actingType,omitempty"`
	Brand         *string `json:"brand,omitempty" bson:"brand,omitempty"`
	Concentration *int    `json:"concentration,omitempty" bson:"concentration,omitempty"`
	Dose          *Dose   `json:"dose,omitempty" bson:"dose,omitempty"`
	Name          *string `json:"name,omitempty" bson:"name,omitempty"`
	Site          *string `json:"site,omitempty" bson:"site,omitempty"`
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

	i.ActingType = parser.ParseString("actingType")
	i.Brand = parser.ParseString("brand")
	i.Concentration = parser.ParseInteger("concentration")
	i.Dose = ParseDose(parser.NewChildObjectParser("dose"))
	i.Name = parser.ParseString("name")
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

	validator.String("actingType", i.ActingType).OneOf(ActingTypes()...)
	validator.String("brand", i.Brand).NotEmpty().LengthLessThanOrEqualTo(BrandLengthMaximum)
	validator.Int("concentration", i.Concentration).InRange(ConcentrationMinimum, ConcentrationMaximum)
	if i.Dose != nil {
		i.Dose.Validate(validator.WithReference("dose"))
	} else {
		validator.WithReference("dose").ReportError(structureValidator.ErrorValueNotExists())
	}
	validator.String("name", i.Name).NotEmpty().LengthLessThanOrEqualTo(NameLengthMaximum)
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
}
