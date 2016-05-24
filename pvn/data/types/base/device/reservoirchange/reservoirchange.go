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
	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/types/base/device"
)

type ReservoirChange struct {
	device.Device `bson:",inline"`

	StatusID *string `json:"status" bson:"status"`
}

func Type() string {
	return device.Type()
}

func SubType() string {
	return "reservoirChange"
}

func New() *ReservoirChange {
	reservoirChangeType := Type()
	reservoirChangeSubType := SubType()

	reservoirChange := &ReservoirChange{}
	reservoirChange.Type = &reservoirChangeType
	reservoirChange.SubType = &reservoirChangeSubType
	return reservoirChange
}

func (r *ReservoirChange) Parse(parser data.ObjectParser) {
	r.Device.Parse(parser)
	r.StatusID = parser.ParseString("status")
}

func (r *ReservoirChange) Validate(validator data.Validator) {
	r.Device.Validate(validator)
	validator.ValidateString("status", r.StatusID).Exists().LengthGreaterThan(1)
}

func (r *ReservoirChange) Normalize(normalizer data.Normalizer) {
	r.Device.Normalize(normalizer)
}
