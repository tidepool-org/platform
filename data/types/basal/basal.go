package basal

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/structure"
)

// TODO: Can we use suppressed by reference only (i.e. by id)?

const (
	Type = "basal"
)

type Basal struct {
	types.Base `bson:",inline"`

	DeliveryType string `json:"deliveryType,omitempty" bson:"deliveryType,omitempty"`
}

type Meta struct {
	Type         string `json:"type,omitempty"`
	DeliveryType string `json:"deliveryType,omitempty"`
}

func (b *Basal) Init() {
	b.Base.Init()
	b.Type = Type

	b.DeliveryType = ""
}

func (b *Basal) Meta() interface{} {
	return &Meta{
		Type:         b.Type,
		DeliveryType: b.DeliveryType,
	}
}

func (b *Basal) Parse(parser data.ObjectParser) error {
	parser.SetMeta(b.Meta())

	return b.Base.Parse(parser)
}

func (b *Basal) Validate(validator structure.Validator) {
	b.Base.Validate(validator)

	if b.Type != "" {
		validator.String("type", &b.Type).EqualTo(Type)
	}

	validator.String("deliveryType", &b.DeliveryType).Exists().NotEmpty()
}

func (b *Basal) IdentityFields() ([]string, error) {
	identityFields, err := b.Base.IdentityFields()
	if err != nil {
		return nil, err
	}

	if b.DeliveryType == "" {
		return nil, errors.New("delivery type is empty")
	}

	return append(identityFields, b.DeliveryType), nil
}

func ParseDeliveryType(parser data.ObjectParser) *string {
	if parser.Object() == nil {
		return nil
	}

	typ := parser.ParseString("type")
	if typ == nil {
		parser.AppendError("type", service.ErrorValueNotExists())
		return nil
	}
	if *typ != Type {
		parser.AppendError("type", service.ErrorValueStringNotOneOf(*typ, []string{Type}))
		return nil
	}

	dlvryTyp := parser.ParseString("deliveryType")
	if dlvryTyp == nil {
		parser.AppendError("deliveryType", service.ErrorValueNotExists())
		return nil
	}

	return dlvryTyp
}
