package pumpstatus

import (
	"time"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

func BasalDeliveryStates() []string {
	return []string{
		"active",
		"initiatingTempBasal",
		"tempBasal",
		"cancelingTempBasal",
		"suspended",
		"suspending",
		"resuming",
	}
}

type BasalDeliveryState struct {
	State     *string         `json:"state,omitempty" bson:"state,omitempty"`
	DoseEntry *data.DoseEntry `json:"doseEntry,omitempty" bson:"doseEntry,omitempty"`
	Date      *string         `json:"date,omitempty" bson:"date,omitempty"`
}

func ParseBasalDeliveryState(parser structure.ObjectParser) *BasalDeliveryState {
	if !parser.Exists() {
		return nil
	}
	datum := NewBasalDeliveryState()
	parser.Parse(datum)
	return datum
}
func NewBasalDeliveryState() *BasalDeliveryState {
	return &BasalDeliveryState{}
}
func (b *BasalDeliveryState) Parse(parser structure.ObjectParser) {
	b.State = parser.String("unit")
	b.DoseEntry = data.ParseDoseEntry(parser.WithReferenceObjectParser("doseEntry"))
	b.Date = parser.String("Date")
}

func (b *BasalDeliveryState) Validate(validator structure.Validator) {
	validator.String("state", b.State).Exists().OneOf(BasalDeliveryStates()...)
	if b.Date != nil {
		validator.String("date", b.Date).AsTime(time.RFC3339Nano)
	}
	if b.DoseEntry != nil {
		b.DoseEntry.Validate(validator.WithReference("doseEntry"))
	}
}

func (b *BasalDeliveryState) Normalize(normalizer data.Normalizer) {
}
