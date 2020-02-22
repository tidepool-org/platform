package pumpstatus

import (
	"time"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	Active              = "active"
	InitiatingTempBasal = "initiatingTempBasal"
	TempBasal           = "tempBasal"
	CancelingTempBasal  = "cancelingTempBasal"
	Suspended           = "suspended"
	Suspending          = "suspending"
	Resuming            = "resuming"
)

func BasalDeliveryStates() []string {
	return []string{
		Active,
		InitiatingTempBasal,
		TempBasal,
		CancelingTempBasal,
		Suspended,
		Suspending,
		Resuming,
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
	if b.State != nil {
		// With Active and suspended states - we have a date parameter
		if *b.State == Active || *b.State == Suspended {
			validator.String("date", b.Date).Exists().AsTime(time.RFC3339Nano)
		} else {
			if b.Date != nil {
				validator.WithReference("date").ReportError(structureValidator.ErrorValueExists())
			}
		}

		// With tempBasal states - we have a date parameter
		if *b.State == TempBasal {
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

func (b *BasalDeliveryState) Normalize(normalizer data.Normalizer) {
}
