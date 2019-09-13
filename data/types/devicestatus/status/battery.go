package status

import (
	"github.com/tidepool-org/platform/structure"
)

type Battery struct {
	Unit  *string  `json:"unit,omitempty" bson:"unit,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

type BatteryStruct struct {
	Battery *Battery `json:"battery,omitempty" bson:"battery,omitempty"`
}

func (b *BatteryStruct) statusObject() {
}

func ParseBatteryStruct(parser structure.ObjectParser) TypeStatusInterface {
	if !parser.Exists() {
		return nil
	}
	datum := NewBatteryStruct()
	parser.Parse(datum)
	return datum
}

func NewBatteryStruct() *BatteryStruct {
	return &BatteryStruct{}
}

func (b *BatteryStruct) Parse(parser structure.ObjectParser) {
	b.Battery = ParseBattery(parser.WithReferenceObjectParser("battery"))
}

func ParseBattery(parser structure.ObjectParser) *Battery {
	if !parser.Exists() {
		return nil
	}
	datum := NewBattery()
	parser.Parse(datum)
	return datum
}
func NewBattery() *Battery {
	return &Battery{}
}
func (b *Battery) Parse(parser structure.ObjectParser) {
	b.Unit = parser.String("unit")
	b.Value = parser.Float64("value")
}
