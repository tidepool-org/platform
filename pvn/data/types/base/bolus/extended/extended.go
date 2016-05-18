package extended

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

type Extended struct {
	bolus.Bolus

	Duration *int     `json:"duration" bson:"duration"`
	Extended *float64 `json:"extended" bson:"extended"`
}

func Type() string {
	return bolus.Type()
}

func SubType() string {
	return "square"
}

func New() *Extended {
	extendedType := Type()
	extendedSubType := SubType()

	extended := &Extended{}
	extended.Type = &extendedType
	extended.SubType = &extendedSubType
	return extended
}

func (e *Extended) Parse(parser data.ObjectParser) {
	e.Bolus.Parse(parser)
	e.Duration = parser.ParseInteger("duration")
	e.Extended = parser.ParseFloat("extended")
}

func (e *Extended) Validate(validator data.Validator) {
	e.Bolus.Validate(validator)
	validator.ValidateInteger("duration", e.Duration).Exists().InRange(0, 86400000)
	validator.ValidateFloat("extended", e.Extended).Exists().GreaterThan(0.0).LessThanOrEqualTo(100.0)
}

func (e *Extended) Normalize(normalizer data.Normalizer) {
	e.Bolus.Normalize(normalizer)
}
