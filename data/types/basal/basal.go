package basal

import (
	"github.com/tidepool-org/platform/data/types"
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

func (b *Basal) IdentityFields(version int) ([]string, error) {
	identityFields := []string{}
	var err error
	if version == types.LegacyIdentityFieldsVersion {

		identityFields, err = types.AppendIdentityFieldVal(identityFields, &b.Type, "type")
		if err != nil {
			return nil, err
		}
		identityFields, err = types.AppendIdentityFieldVal(identityFields, &b.DeliveryType, "delivery type")
		if err != nil {
			return nil, err
		}
		identityFields, err = types.AppendIdentityFieldVal(identityFields, b.DeviceID, "device id")
		if err != nil {
			return nil, err
		}
		identityFields, err = types.AppendLegacyTimeVal(identityFields, b.Time)
		if err != nil {
			return nil, err
		}
		return identityFields, nil
	}

	identityFields, err = b.Base.IdentityFields(version)
	if err != nil {
		return nil, err
	}
	identityFields, err = types.AppendIdentityFieldVal(identityFields, &b.DeliveryType, "delivery type")
	if err != nil {
		return nil, err
	}

	return identityFields, nil
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
