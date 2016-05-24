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
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/base/device"
)

type TimeChange struct {
	device.Device `bson:",inline"`

	*Change `json:"change,omitempty" bson:"change,omitempty"`
}

func SubType() string {
	return "timeChange"
}

func New() (*TimeChange, error) {
	timeChangeDevice, err := device.New(SubType())
	if err != nil {
		return nil, err
	}

	return &TimeChange{
		Device: *timeChangeDevice,
	}, nil
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
