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

	Change *Change `json:"change,omitempty" bson:"change,omitempty"`
}

func SubType() string {
	return "timeChange"
}

func NewDatum() data.Datum {
	return New()
}

func New() *TimeChange {
	return &TimeChange{}
}

func Init() *TimeChange {
	timeChange := New()
	timeChange.Init()
	return timeChange
}

func (t *TimeChange) Init() {
	t.Device.Init()
	t.SubType = SubType()

	t.Change = nil
}

func (t *TimeChange) Parse(parser data.ObjectParser) error {
	if err := t.Device.Parse(parser); err != nil {
		return err
	}

	t.Change = ParseChange(parser.NewChildObjectParser("change"))

	return nil
}

func (t *TimeChange) Validate(validator data.Validator) error {
	if err := t.Device.Validate(validator); err != nil {
		return err
	}

	if t.Change != nil {
		t.Change.Validate(validator.NewChildValidator("change"))
	}

	return nil
}
