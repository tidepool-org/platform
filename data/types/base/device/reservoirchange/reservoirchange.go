package reservoirchange

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

type ReservoirChange struct {
	device.Device `bson:",inline"`

	StatusID *string `json:"status,omitempty" bson:"status,omitempty"`
}

func SubType() string {
	return "reservoirChange"
}

func NewDatum() data.Datum {
	return New()
}

func New() *ReservoirChange {
	return &ReservoirChange{}
}

func Init() *ReservoirChange {
	reservoirChange := New()
	reservoirChange.Init()
	return reservoirChange
}

func (r *ReservoirChange) Init() {
	r.Device.Init()
	r.Device.SubType = SubType()

	r.StatusID = nil
}

func (r *ReservoirChange) Parse(parser data.ObjectParser) error {
	if err := r.Device.Parse(parser); err != nil {
		return err
	}

	r.StatusID = parser.ParseString("status")

	return nil
}

func (r *ReservoirChange) Validate(validator data.Validator) error {
	if err := r.Device.Validate(validator); err != nil {
		return err
	}

	validator.ValidateString("status", r.StatusID).LengthGreaterThan(1) // TODO_DATA: .Exists() does not exist in Animas currently

	return nil
}
