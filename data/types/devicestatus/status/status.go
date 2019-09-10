package status

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

type Array []Interface

type Interface interface {
	statusObject()
}

type Battery struct {
	Unit  *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

type BatteryStruct struct {
	Battery *Battery `json:"battery,omitempty" bson:"battery,omitempty"`
}

func (b *BatteryStruct) statusObject() {
}

type ReservoirRemaining struct {
	Unit   *string  `json:"units,omitempty" bson:"units,omitempty"`
	Amount *float64 `json:"amount,omitempty" bson:"amount,omitempty"`
}

type ReservoirRemainingStruct struct {
	ReservoirRemaining *ReservoirRemaining `json:"reservoirRemaining,omitempty" bson:"reservoirRemaining,omitempty"`
}

func (r *ReservoirRemainingStruct) statusObject() {
}

type SignalStrength struct {
	Unit  *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

type SignalStrengthStruct struct {
	SignalStrength *SignalStrength `json:"signalStrength,omitempty" bson:"signalStrength,omitempty"`
}

func (s *SignalStrengthStruct) statusObject() {
}

type AlertsStruct struct {
	Alerts []string `json:"alerts,omitempty" bson:"alerts,omitempty"`
}

func (a *AlertsStruct) statusObject() {
}

type ForecastStruct struct {
	Forecast *data.Forecast `json:"forecast,omitempty" bson:"forecast,omitempty"`
}

func (f *ForecastStruct) statusObject() {
}

func NewStatusArray() *Array {
	return &Array{}
}

func (a *Array) Parse(parser structure.ObjectParser) {
}

func (a *Array) Validate(validator structure.Validator) {
}

func (a *Array) Normalize(normalizer data.Normalizer) {}
