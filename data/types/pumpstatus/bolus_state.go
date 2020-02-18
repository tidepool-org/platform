package pumpstatus

import (
	"time"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
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
	validator.String("date", b.Date).Exists().AsTime(time.RFC3339Nano)
	if b.DoseEntry != nil {
		b.DoseEntry.Validate(validator.WithReference("doseEntry"))
	} else {
		validator.WithReference("doseEntry").ReportError(structureValidator.ErrorValueNotExists())
	}
}

func (b *BolusState) Normalize(normalizer data.Normalizer) {
}
