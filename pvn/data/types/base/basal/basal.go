package basal

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import "github.com/tidepool-org/platform/pvn/data"
import "github.com/tidepool-org/platform/pvn/data/types/base"

type Basal struct {
	base.Base
	DeliveryType *string `json:"deliveryType" bson:"deliveryType"`
}

func Type() string {
	return "basal"
}

func New() *Basal {
	basalType := Type()

	basal := &Basal{}
	basal.Type = &basalType
	return basal
}

func (b *Basal) Parse(parser data.ObjectParser) {
	b.Base.Parse(parser)
}

func (b *Basal) Validate(validator data.Validator) {

	b.Base.Validate(validator)
	validator.ValidateString("type", b.Type).Exists().EqualTo(Type())
	validator.ValidateString("deliveryType", b.DeliveryType).Exists()
}

func (b *Basal) Normalize(normalizer data.Normalizer) {
	b.Base.Normalize(normalizer)
}
