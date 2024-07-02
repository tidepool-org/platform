package basal

import (
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

// TODO: Can we use suppressed by reference only (i.e. by id)?

const (
	Type = "basal"

	ScheduleNameLengthMaximum = 1000
)

type Basal struct {
	types.Base `bson:",inline"`

	DeliveryType string `json:"deliveryType,omitempty" bson:"deliveryType,omitempty"`
}

type Meta struct {
	Type         string `json:"type,omitempty"`
	DeliveryType string `json:"deliveryType,omitempty"`
}

func New(deliveryType string) Basal {
	return Basal{
		Base:         types.New(Type),
		DeliveryType: deliveryType,
	}
}

func (b *Basal) Meta() interface{} {
	return &Meta{
		Type:         b.Type,
		DeliveryType: b.DeliveryType,
	}
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

func (b *Basal) LegacyIdentityFields() ([]string, error) {
	return types.GetLegacyIdentityFields(&b.Base, types.TypeDeviceIDTimeFormat, &types.LegacyIdentityCustomField{Value: b.DeliveryType, Name: "delivery type"})
}

func ParseDeliveryType(parser structure.ObjectParser) *string {
	if !parser.Exists() {
		return nil
	}

	typ := parser.String("type")
	if typ == nil {
		parser.WithReferenceErrorReporter("type").ReportError(structureValidator.ErrorValueNotExists())
		return nil
	}
	if *typ != Type {
		parser.WithReferenceErrorReporter("type").ReportError(structureValidator.ErrorValueStringNotOneOf(*typ, []string{Type}))
		return nil
	}

	dlvryTyp := parser.String("deliveryType")
	if dlvryTyp == nil {
		parser.WithReferenceErrorReporter("deliveryType").ReportError(structureValidator.ErrorValueNotExists())
		return nil
	}

	return dlvryTyp
}
