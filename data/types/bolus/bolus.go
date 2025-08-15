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

	DeliveryContextAlgorithm    = "algorithm"
	DeliveryContextDevice       = "device"
	DeliveryContextOneButton    = "oneButton"
	DeliveryContextRemote       = "remote"
	DeliveryContextUndetermined = "undetermined"
	DeliveryContextWatch        = "watch"
)

func DeliveryContexts() []string {
	return []string{
		DeliveryContextAlgorithm,
		DeliveryContextDevice,
		DeliveryContextOneButton,
		DeliveryContextRemote,
		DeliveryContextUndetermined,
		DeliveryContextWatch,
	}
}

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

	validator.String("deliveryContext", b.DeliveryContext).OneOf(DeliveryContexts()...)
	if b.InsulinFormulation != nil {
		b.InsulinFormulation.Validate(validator.WithReference("insulinFormulation"))
	}
}

func (b *Bolus) Normalize(normalizer data.Normalizer) {
	b.Base.Normalize(normalizer)

	if b.InsulinFormulation != nil {
		b.InsulinFormulation.Normalize(normalizer.WithReference("insulinFormulation"))
	}
}

func (b *Bolus) IdentityFields(version string) ([]string, error) {
	identityFields, err := b.Base.IdentityFields(version)
	if err != nil {
		return nil, err
	}

	switch version {
	case types.IdentityFieldsVersionDeviceID, types.IdentityFieldsVersionDataSetID:
		if b.SubType == "" {
			return nil, errors.New("sub type is empty")
		}
		return append(identityFields, b.SubType), nil
	default:
		return nil, errors.New("version is invalid")
	}
}
