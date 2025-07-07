package dosingdecision

import (
	dataTypesBolusCombination "github.com/tidepool-org/platform/data/types/bolus/combination"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	BolusAmountMaximum   = dataTypesBolusCombination.NormalMaximum
	BolusAmountMinimum   = dataTypesBolusCombination.NormalMinimum
	BolusDurationMaximum = dataTypesBolusCombination.DurationMaximum
	BolusDurationMinimum = dataTypesBolusCombination.DurationMinimum
	BolusExtendedMaximum = dataTypesBolusCombination.ExtendedMaximum
	BolusExtendedMinimum = dataTypesBolusCombination.ExtendedMinimum
	BolusNormalMaximum   = dataTypesBolusCombination.NormalMaximum
	BolusNormalMinimum   = dataTypesBolusCombination.NormalMinimum
)

type Bolus struct {
	Amount   *float64 `json:"amount,omitempty" bson:"amount,omitempty"`
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
	b.Amount = parser.Float64("amount")
	b.Duration = parser.Int("duration")
	b.Extended = parser.Float64("extended")
	b.Normal = parser.Float64("normal")
}

func (b *Bolus) Validate(validator structure.Validator) {
	if b.Amount != nil {
		validator.Float64("amount", b.Amount).InRange(BolusAmountMinimum, BolusAmountMaximum)
		validator.Int("duration", b.Duration).NotExists()
		validator.Float64("extended", b.Extended).NotExists()
		validator.Float64("normal", b.Normal).NotExists()
	} else {
		if durationValidator := validator.Int("duration", b.Duration); b.Extended != nil {
			durationValidator.Exists().InRange(BolusDurationMinimum, BolusDurationMaximum)
		} else {
			durationValidator.NotExists()
		}
		validator.Float64("extended", b.Extended).InRange(BolusExtendedMinimum, BolusExtendedMaximum)
		validator.Float64("normal", b.Normal).InRange(BolusNormalMinimum, BolusNormalMaximum)
	}

	if b.Amount == nil && b.Extended == nil && b.Normal == nil {
		validator.ReportError(structureValidator.ErrorValuesNotExistForAny("amount", "extended", "normal"))
	}
}
