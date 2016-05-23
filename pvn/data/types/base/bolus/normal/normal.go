package normal

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
	"github.com/tidepool-org/platform/pvn/data/types/base/bolus"
)

type Normal struct {
	bolus.Bolus `bson:",inline"`
	Normal      *float64 `json:"normal" bson:"normal"`
}

func Type() string {
	return bolus.Type()
}

func SubType() string {
	return "normal"
}

func New() *Normal {
	normalType := Type()
	normalSubType := SubType()

	normal := &Normal{}
	normal.Type = &normalType
	normal.SubType = &normalSubType
	return normal
}

func (n *Normal) Parse(parser data.ObjectParser) {
	n.Bolus.Parse(parser)
	n.Normal = parser.ParseFloat("normal")
}

func (n *Normal) Validate(validator data.Validator) {
	n.Bolus.Validate(validator)
	validator.ValidateFloat("normal", n.Normal).Exists().GreaterThan(0.0).LessThanOrEqualTo(100.0)
}

func (n *Normal) Normalize(normalizer data.Normalizer) {
	n.Bolus.Normalize(normalizer)
}
