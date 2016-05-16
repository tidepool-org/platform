package bolus

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

type Bolus struct {
	base.Base
	SubType *string `json:"subType" bson:"subType"`
}

func Type() string {
	return "bolus"
}

func New() *Bolus {
	bolusType := Type()

	bolus := &Bolus{}
	bolus.Type = &bolusType
	return bolus
}

func (b *Bolus) Parse(parser data.ObjectParser) {
	b.Base.Parse(parser)
}

func (b *Bolus) Validate(validator data.Validator) {

	b.Base.Validate(validator)
	validator.ValidateString("type", b.Type).Exists().EqualTo(Type())
	validator.ValidateString("subType", b.SubType).Exists()
}

func (b *Bolus) Normalize(normalizer data.Normalizer) {
	b.Base.Normalize(normalizer)
}
