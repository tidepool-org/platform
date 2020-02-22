package pumpstatus

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	None       = "none"
	Initiating = "initiating"
	InProgress = "inProgress"
	Canceling  = "canceling"
)

func BolusStates() []string {
	return []string{
		None,
		Initiating,
		InProgress,
		Canceling,
	}
}

type BolusState struct {
	State     *string         `json:"state,omitempty" bson:"state,omitempty"`
	DoseEntry *data.DoseEntry `json:"doseEntry,omitempty" bson:"doseEntry,omitempty"`
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
	b.State = parser.String("state")
	b.DoseEntry = data.ParseDoseEntry(parser.WithReferenceObjectParser("doseEntry"))
}

func (b *BolusState) Validate(validator structure.Validator) {
	validator.String("state", b.State).Exists().OneOf(BolusStates()...)
	if b.State != nil {
		if *b.State == InProgress {
			if b.DoseEntry != nil {
				b.DoseEntry.Validate(validator.WithReference("doseEntry"))
			} else {
				validator.WithReference("doseEntry").ReportError(structureValidator.ErrorValueNotExists())
			}
		} else {
			if b.DoseEntry != nil {
				validator.WithReference("doseEntry").ReportError(structureValidator.ErrorValueExists())
			}
		}
	}
}

func (b *BolusState) Normalize(normalizer data.Normalizer) {
}
