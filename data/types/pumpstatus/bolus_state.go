package pumpstatus

import (
	"time"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

func BolusStates() []string {
	return []string{
		"none",
		"initiating",
		"inProgress",
		"canceling",
	}
}

type BolusState struct {
	State     *string         `json:"state,omitempty" bson:"state,omitempty"`
	DoseEntry *data.DoseEntry `json:"doseEntry,omitempty" bson:"doseEntry,omitempty"`
	Date      *string         `json:"date,omitempty" bson:"date,omitempty"`
}

func ParseBolusState(parser structure.ObjectParser) *BolusState {
	if !parser.Exists() {
		return nil
	}
	datum := NewBolusState()
	parser.Parse(datum)
	return datum
}
func NewBolusState() *BolusState {
	return &BolusState{}
}
func (b *BolusState) Parse(parser structure.ObjectParser) {
	b.State = parser.String("unit")
	b.DoseEntry = data.ParseDoseEntry(parser.WithReferenceObjectParser("doseEntry"))
	b.Date = parser.String("Date")
}

func (b *BolusState) Validate(validator structure.Validator) {
	validator.String("state", b.State).Exists().OneOf(BolusStates()...)
	if b.Date != nil {
		validator.String("date", b.Date).AsTime(time.RFC3339Nano)
	}
	if b.DoseEntry != nil {
		b.DoseEntry.Validate(validator.WithReference("doseEntry"))
	}
}

func (b *BolusState) Normalize(normalizer data.Normalizer) {
}
