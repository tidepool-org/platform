package pumpstatus

import (
	"time"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	BasalDeliveryStateCancelingTemporary  = "cancelingTemporary"
	BasalDeliveryStateInitiatingTemporary = "initiatingTemporary"
	BasalDeliveryStateResuming            = "resuming"
	BasalDeliveryStateScheduled           = "scheduled"
	BasalDeliveryStateSuspended           = "suspended"
	BasalDeliveryStateSuspending          = "suspending"
	BasalDeliveryStateTemporary           = "temporary"
	BasalDoseAmountDeliveredMaximum       = 1000
	BasalDoseAmountDeliveredMinimum       = 0
	BasalDoseRateMaximum                  = 100
	BasalDoseRateMinimum                  = 0
)

func BasalDeliveryStates() []string {
	return []string{
		BasalDeliveryStateCancelingTemporary,
		BasalDeliveryStateInitiatingTemporary,
		BasalDeliveryStateResuming,
		BasalDeliveryStateScheduled,
		BasalDeliveryStateSuspended,
		BasalDeliveryStateSuspending,
		BasalDeliveryStateTemporary,
	}
}

type BasalDelivery struct {
	State *string    `json:"state,omitempty" bson:"state,omitempty"`
	Time  *string    `json:"time,omitempty" bson:"time,omitempty"`
	Dose  *BasalDose `json:"dose,omitempty" bson:"dose,omitempty"`
}

func ParseBasalDelivery(parser structure.ObjectParser) *BasalDelivery {
	if !parser.Exists() {
		return nil
	}
	datum := NewBasalDelivery()
	parser.Parse(datum)
	return datum
}

func NewBasalDelivery() *BasalDelivery {
	return &BasalDelivery{}
}

func (b *BasalDelivery) Parse(parser structure.ObjectParser) {
	b.State = parser.String("state")
	b.Time = parser.String("time")
	b.Dose = ParseBasalDose(parser.WithReferenceObjectParser("dose"))
}

func (b *BasalDelivery) Validate(validator structure.Validator) {
	validator.String("state", b.State).Exists().OneOf(BasalDeliveryStates()...)
	if dateValidator := validator.String("time", b.Time); b.State != nil && (*b.State == BasalDeliveryStateScheduled || *b.State == BasalDeliveryStateSuspended) {
		dateValidator.Exists().AsTime(time.RFC3339Nano)
	} else if b.Time != nil {
		dateValidator.NotExists()
	}
	if doseValidator := validator.WithReference("dose"); b.State != nil && *b.State == BasalDeliveryStateTemporary {
		if b.Dose != nil {
			b.Dose.Validate(doseValidator)
		} else {
			doseValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	} else if b.Dose != nil {
		doseValidator.ReportError(structureValidator.ErrorValueExists())
	}
}

type BasalDose struct {
	StartTime       *string  `json:"startTime,omitempty" bson:"startTime,omitempty"`
	EndTime         *string  `json:"endTime,omitempty" bson:"endTime,omitempty"`
	Rate            *float64 `json:"rate,omitempty" bson:"rate,omitempty"`
	AmountDelivered *float64 `json:"amountDelivered,omitempty" bson:"amountDelivered,omitempty"`
}

func ParseBasalDose(parser structure.ObjectParser) *BasalDose {
	if !parser.Exists() {
		return nil
	}
	datum := NewBasalDose()
	parser.Parse(datum)
	return datum
}

func NewBasalDose() *BasalDose {
	return &BasalDose{}
}

func (b *BasalDose) Parse(parser structure.ObjectParser) {
	b.StartTime = parser.String("startTime")
	b.EndTime = parser.String("endTime")
	b.Rate = parser.Float64("rate")
	b.AmountDelivered = parser.Float64("amountDelivered")
}

func (b *BasalDose) Validate(validator structure.Validator) {
	var startTime time.Time

	if b.StartTime != nil {
		startTime, _ = time.Parse(time.RFC3339Nano, *b.StartTime)
	}

	validator.String("startTime", b.StartTime).AsTime(time.RFC3339Nano)
	validator.String("endTime", b.EndTime).AsTime(time.RFC3339Nano).After(startTime)
	validator.Float64("rate", b.Rate).Exists().InRange(BasalDoseRateMinimum, BasalDoseRateMaximum)
	validator.Float64("amountDelivered", b.AmountDelivered).InRange(BasalDoseAmountDeliveredMinimum, BasalDoseAmountDeliveredMaximum)
}
