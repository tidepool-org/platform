package basal

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/base"
)

type Basal struct {
	base.Base `bson:",inline"`

	DeliveryType string `json:"deliveryType,omitempty" bson:"deliveryType,omitempty"`
}

type Meta struct {
	Type         string `json:"type,omitempty"`
	DeliveryType string `json:"deliveryType,omitempty"`
}

func Type() string {
	return "basal"
}

func (b *Basal) Init() {
	b.Base.Init()
	b.Type = Type()

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

func (b *Basal) Validate(validator data.Validator) error {
	validator.SetMeta(b.Meta())

	if err := b.Base.Validate(validator); err != nil {
		return err
	}

	validator.ValidateString("type", &b.Type).EqualTo(Type())

	validator.ValidateString("deliveryType", &b.DeliveryType).NotEmpty()

	return nil
}

func (b *Basal) Normalize(normalizer data.Normalizer) error {
	normalizer.SetMeta(b.Meta())

	return b.Base.Normalize(normalizer)
}

func (b *Basal) IdentityFields() ([]string, error) {
	identityFields, err := b.Base.IdentityFields()
	if err != nil {
		return nil, err
	}

	if b.DeliveryType == "" {
		return nil, app.Error("basal", "delivery type is empty")
	}

	return append(identityFields, b.DeliveryType), nil
}
