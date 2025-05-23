package dosingdecision

import (
	dataTypesBolusCombination "github.com/tidepool-org/platform/data/types/bolus/combination"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	// DEPRECATED: Use BolusNormalMaximum and BolusNormalMinimum
	BolusAmountMaximum = BolusNormalMaximum
	BolusAmountMinimum = BolusNormalMinimum

	BolusDurationMaximum = dataTypesBolusCombination.DurationMaximum
	BolusDurationMinimum = dataTypesBolusCombination.DurationMinimum
	BolusExtendedMaximum = dataTypesBolusCombination.ExtendedMaximum
	BolusExtendedMinimum = dataTypesBolusCombination.ExtendedMinimum
	BolusNormalMaximum   = dataTypesBolusCombination.NormalMaximum
	BolusNormalMinimum   = dataTypesBolusCombination.NormalMinimum
)

type Bolus struct {
	// DEPRECATED: Use Normal
	Amount *float64 `json:"amount,omitempty" bson:"amount,omitempty"`

	Duration *int     `json:"duration,omitempty" bson:"duration,omitempty"`
	Extended *float64 `json:"extended,omitempty" bson:"extended,omitempty"`
	Normal   *float64 `json:"normal,omitempty" bson:"normal,omitempty"`
}

func ParseBolus(parser structure.ObjectParser) *Bolus {
	if !parser.Exists() {
		return nil
	}
	datum := NewBolus()
	parser.Parse(datum)
	return datum
}

func NewBolus() *Bolus {
	return &Bolus{}
}

func (b *Bolus) Parse(parser structure.ObjectParser) {
	// DEPRECATED: Use Normal
	b.Amount = parser.Float64("amount")

	b.Duration = parser.Int("duration")
	b.Extended = parser.Float64("extended")
	b.Normal = parser.Float64("normal")
}

func (b *Bolus) Validate(validator structure.Validator) {
	// DEPRECATED: Use Normal
	if b.Amount != nil {
		validator.Float64("amount", b.Amount).Exists().InRange(BolusAmountMinimum, BolusAmountMaximum)
		validator.Int("duration", b.Duration).NotExists()
		validator.Float64("extended", b.Extended).NotExists()
		validator.Float64("normal", b.Normal).NotExists()
		return
	}

	if b.Extended != nil {
		validator.Int("duration", b.Duration).Exists().InRange(BolusDurationMinimum, BolusDurationMaximum)
		validator.Float64("extended", b.Extended).Exists().InRange(BolusExtendedMinimum, BolusExtendedMaximum)
	} else {
		validator.Int("duration", b.Duration).NotExists()
	}
	validator.Float64("normal", b.Normal).InRange(BolusNormalMinimum, BolusNormalMaximum)

	if b.Extended == nil && b.Normal == nil {
		validator.ReportError(structureValidator.ErrorValuesNotExistForAny("normal", "extended"))
	}
}
