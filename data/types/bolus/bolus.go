package bolus

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
)

const (
	Type = "bolus"
)

type Bolus struct {
	types.Base `bson:",inline"`

	SubType string `json:"subType,omitempty" bson:"subType,omitempty"`

	InsulinType *insulin.InsulinType `json:"insulinType,omitempty" bson:"insulinType,omitempty"`
}

type Meta struct {
	Type    string `json:"type,omitempty"`
	SubType string `json:"subType,omitempty"`
}

func New(subType string) Bolus {
	return Bolus{
		Base:    types.New(Type),
		SubType: subType,
	}
}

func (b *Bolus) Meta() interface{} {
	return &Meta{
		Type:    b.Type,
		SubType: b.SubType,
	}
}

func (b *Bolus) Parse(parser data.ObjectParser) error {
	parser.SetMeta(b.Meta())

	if err := b.Base.Parse(parser); err != nil {
		return nil
	}

	b.InsulinType = insulin.ParseInsulinType(parser.NewChildObjectParser("insulinType"))

	return nil
}

func (b *Bolus) Validate(validator structure.Validator) {
	b.Base.Validate(validator)

	if b.Type != "" {
		validator.String("type", &b.Type).EqualTo(Type)
	}

	validator.String("subType", &b.SubType).Exists().NotEmpty()

	if b.InsulinType != nil {
		b.InsulinType.Validate(validator.WithReference("insulinType"))
	}
}

func (b *Bolus) Normalize(normalizer data.Normalizer) {
	b.Base.Normalize(normalizer)

	if b.InsulinType != nil {
		b.InsulinType.Normalize(normalizer.WithReference("insulinType"))
	}
}

func (b *Bolus) IdentityFields() ([]string, error) {
	identityFields, err := b.Base.IdentityFields()
	if err != nil {
		return nil, err
	}

	if b.SubType == "" {
		return nil, errors.New("sub type is empty")
	}

	return append(identityFields, b.SubType), nil
}
