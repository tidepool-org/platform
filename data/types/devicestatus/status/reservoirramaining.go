package status

import "github.com/tidepool-org/platform/structure"

type ReservoirRemaining struct {
	Unit   *string  `json:"units,omitempty" bson:"units,omitempty"`
	Amount *float64 `json:"amount,omitempty" bson:"amount,omitempty"`
}

type ReservoirRemainingStruct struct {
	ReservoirRemaining *ReservoirRemaining `json:"reservoirRemaining,omitempty" bson:"reservoirRemaining,omitempty"`
}

func (r *ReservoirRemainingStruct) statusObject() {
}

func ParseReservoirRemainingStruct(parser structure.ObjectParser) *ReservoirRemainingStruct {
	if !parser.Exists() {
		return nil
	}
	datum := NewReservoirRemainingStruct()
	parser.Parse(datum)
	return datum
}

func NewReservoirRemainingStruct() *ReservoirRemainingStruct {
	return &ReservoirRemainingStruct{}
}

func (r *ReservoirRemainingStruct) Parse(parser structure.ObjectParser) {
	r.ReservoirRemaining = ParseReservoirRemaining(parser.WithReferenceObjectParser("battery"))
}

func ParseReservoirRemaining(parser structure.ObjectParser) *ReservoirRemaining {
	if !parser.Exists() {
		return nil
	}
	datum := NewReservoirRemaining()
	parser.Parse(datum)
	return datum
}
func NewReservoirRemaining() *ReservoirRemaining {
	return &ReservoirRemaining{}
}
func (r *ReservoirRemaining) Parse(parser structure.ObjectParser) {
	r.Unit = parser.String("unit")
	r.Amount = parser.Float64("value")
}
