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

	DeliveryContextDevice       = "device"
	DeliveryContextAlgorithm    = "algorithm"
	DeliveryContextRemote       = "remote"
	DeliveryContextUndetermined = "undetermined"
)

type Bolus struct {
	types.Base `bson:",inline"`

	SubType string `json:"subType,omitempty" bson:"subType,omitempty"`

	DeliveryContext    *string              `json:"deliveryContext,omitempty" bson:"deliveryContext,omitempty"`
	InsulinFormulation *insulin.Formulation `json:"insulinFormulation,omitempty" bson:"insulinFormulation,omitempty"`
}

type Meta struct {
	Type    string `json:"type,omitempty"`
	SubType string `json:"subType,omitempty"`
}

func DeliveryContext() []string {
	return []string{DeliveryContextDevice, DeliveryContextAlgorithm, DeliveryContextRemote, DeliveryContextUndetermined}
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

func (b *Bolus) Parse(parser structure.ObjectParser) {
	b.Base.Parse(parser)
	b.DeliveryContext = parser.String("deliveryContext")
	b.InsulinFormulation = insulin.ParseFormulation(parser.WithReferenceObjectParser("insulinFormulation"))
}

func (b *Bolus) Validate(validator structure.Validator) {
	b.Base.Validate(validator)

	if b.Type != "" {
		validator.String("type", &b.Type).EqualTo(Type)
	}

	validator.String("subType", &b.SubType).Exists().NotEmpty()

	if b.InsulinFormulation != nil {
		b.InsulinFormulation.Validate(validator.WithReference("insulinFormulation"))
	}

	if b.DeliveryContext != nil {
		validator.String("deliveryContext", b.DeliveryContext).Exists().OneOf(DeliveryContext()...)
	}
}

func (b *Bolus) Normalize(normalizer data.Normalizer) {
	b.Base.Normalize(normalizer)

	if b.InsulinFormulation != nil {
		b.InsulinFormulation.Normalize(normalizer.WithReference("insulinFormulation"))
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

func (b *Bolus) LegacyIdentityFields() ([]string, error) {
	return types.GetLegacyIdentityFields(&b.Base, types.TypeDeviceIDTimeFormat, &types.LegacyIdentityCustomField{Value: b.SubType, Name: "sub type"})
}
