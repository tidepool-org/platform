package bolus

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/errors"
)

type Bolus struct {
	types.Base `bson:",inline"`

	SubType string `json:"subType,omitempty" bson:"subType,omitempty"`
}

type Meta struct {
	Type    string `json:"type,omitempty"`
	SubType string `json:"subType,omitempty"`
}

func Type() string {
	return "bolus"
}

func (b *Bolus) Init() {
	b.Base.Init()
	b.Type = Type()

	b.SubType = ""
}

func (b *Bolus) Meta() interface{} {
	return &Meta{
		Type:    b.Type,
		SubType: b.SubType,
	}
}

func (b *Bolus) Parse(parser data.ObjectParser) error {
	parser.SetMeta(b.Meta())

	return b.Base.Parse(parser)
}

func (b *Bolus) Validate(validator data.Validator) error {
	validator.SetMeta(b.Meta())

	if err := b.Base.Validate(validator); err != nil {
		return err
	}

	validator.ValidateString("type", &b.Type).EqualTo(Type())

	validator.ValidateString("subType", &b.SubType).NotEmpty()

	return nil
}

func (b *Bolus) Normalize(normalizer data.Normalizer) error {
	normalizer.SetMeta(b.Meta())

	return b.Base.Normalize(normalizer)
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
