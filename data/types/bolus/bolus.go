package bolus

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
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

func (b *Bolus) Validate(validator structure.Validator) {
	b.Base.Validate(validator)

	if b.Type != "" {
		validator.String("type", &b.Type).EqualTo(Type())
	}

	validator.String("subType", &b.SubType).Exists().NotEmpty()
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
