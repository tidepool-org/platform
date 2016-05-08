package base

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

type Base struct {
	Type *string `json:"type,omitempty"`
}

func (b *Base) Parse(parser data.ObjectParser) {
	b.Type = parser.ParseString("type")
}

func (b *Base) Validate(validator data.Validator) {
	validator.ValidateString("type", b.Type).Exists()
}

func (b *Base) Normalize(normalizer data.Normalizer) {
}
