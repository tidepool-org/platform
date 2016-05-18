package temporary

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/types/base/basal"
)

type Temporary struct {
	basal.Basal

	Duration *int     `json:"duration" bson:"duration"`
	Rate     *float64 `json:"rate" bson:"rate"`
	Percent  *float64 `json:"percent" bson:"percent"`
}

func Type() string {
	return basal.Type()
}

func DeliveryType() string {
	return "temporary"
}

func New() *Temporary {
	temporaryType := Type()
	temporarySubType := DeliveryType()

	temporary := &Temporary{}
	temporary.Type = &temporaryType
	temporary.DeliveryType = &temporarySubType
	return temporary
}

func (t *Temporary) Parse(parser data.ObjectParser) {
	t.Basal.Parse(parser)
	t.Duration = parser.ParseInteger("duration")
	t.Rate = parser.ParseFloat("rate")
	t.Percent = parser.ParseFloat("percent")
}

func (t *Temporary) Validate(validator data.Validator) {
	t.Basal.Validate(validator)
	validator.ValidateInteger("duration", t.Duration).Exists().InRange(0, 86400000)
	validator.ValidateFloat("rate", t.Rate).Exists().InRange(0.0, 20.0)
	validator.ValidateFloat("percent", t.Percent).Exists().InRange(0.0, 10.0)
}

func (t *Temporary) Normalize(normalizer data.Normalizer) {
	t.Basal.Normalize(normalizer)
}
