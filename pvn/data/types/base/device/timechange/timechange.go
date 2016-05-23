package timechange

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
	"github.com/tidepool-org/platform/pvn/data/types/base/device"
)

type TimeChange struct {
	device.Device `bson:",inline"`

	*Change `json:"change" bson:"change"`
}

func Type() string {
	return device.Type()
}

func SubType() string {
	return "timeChange"
}

func New() *TimeChange {
	timechangeType := Type()
	timechangeSubType := SubType()

	timechange := &TimeChange{}
	timechange.Type = &timechangeType
	timechange.SubType = &timechangeSubType
	return timechange
}

func (t *TimeChange) Parse(parser data.ObjectParser) {
	t.Device.Parse(parser)
	t.Change = ParseChange(parser.NewChildObjectParser("change"))
}

func (t *TimeChange) Validate(validator data.Validator) {
	t.Device.Validate(validator)

	if t.Change != nil {
		t.Change.Validate(validator.NewChildValidator("change"))
	}
}

func (t *TimeChange) Normalize(normalizer data.Normalizer) {
	t.Device.Normalize(normalizer)
}
