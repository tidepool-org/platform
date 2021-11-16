package pump

import (
	"time"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	BolusDeliveryStateCanceling     = "canceling"
	BolusDeliveryStateDelivering    = "delivering"
	BolusDeliveryStateInitiating    = "initiating"
	BolusDeliveryStateNone          = "none"
	BolusDoseAmountDeliveredMaximum = 1000
	BolusDoseAmountDeliveredMinimum = 0
	BolusDoseAmountMaximum          = 1000
	BolusDoseAmountMinimum          = 0
)

func BolusDeliveryStates() []string {
	return []string{
		BolusDeliveryStateCanceling,
		BolusDeliveryStateDelivering,
		BolusDeliveryStateInitiating,
		BolusDeliveryStateNone,
	}
}

type BolusDelivery struct {
	State *string    `json:"state,omitempty" bson:"state,omitempty"`
	Dose  *BolusDose `json:"dose,omitempty" bson:"dose,omitempty"`
}

func ParseBolusDelivery(parser structure.ObjectParser) *BolusDelivery {
	if !parser.Exists() {
		return nil
	}
	datum := NewBolusDelivery()
	parser.Parse(datum)
	return datum
}

func NewBolusDelivery() *BolusDelivery {
	return &BolusDelivery{}
}

func (b *BolusDelivery) Parse(parser structure.ObjectParser) {
	b.State = parser.String("state")
	b.Dose = ParseBolusDose(parser.WithReferenceObjectParser("dose"))
}

func (b *BolusDelivery) Validate(validator structure.Validator) {
	validator.String("state", b.State).Exists().OneOf(BolusDeliveryStates()...)
	if doseValidator := validator.WithReference("dose"); b.State != nil && *b.State == BolusDeliveryStateDelivering {
		if b.Dose != nil {
			b.Dose.Validate(doseValidator)
		} else {
			doseValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	} else if b.Dose != nil {
		doseValidator.ReportError(structureValidator.ErrorValueExists())
	}
}

type BolusDose struct {
	StartTime       *time.Time `json:"startTime,omitempty" bson:"startTime,omitempty"`
	Amount          *float64   `json:"amount,omitempty" bson:"amount,omitempty"`
	AmountDelivered *float64   `json:"amountDelivered,omitempty" bson:"amountDelivered,omitempty"`
}

func ParseBolusDose(parser structure.ObjectParser) *BolusDose {
	if !parser.Exists() {
		return nil
	}
	datum := NewBolusDose()
	parser.Parse(datum)
	return datum
}

func NewBolusDose() *BolusDose {
	return &BolusDose{}
}

func (b *BolusDose) Parse(parser structure.ObjectParser) {
	b.StartTime = parser.Time("startTime", time.RFC3339Nano)
	b.Amount = parser.Float64("amount")
	b.AmountDelivered = parser.Float64("amountDelivered")
}

func (b *BolusDose) Validate(validator structure.Validator) {
	validator.Float64("amount", b.Amount).Exists().InRange(BolusDoseAmountMinimum, BolusDoseAmountMaximum)
	validator.Float64("amountDelivered", b.AmountDelivered).InRange(BolusDoseAmountDeliveredMinimum, BolusDoseAmountDeliveredMaximum)
}
