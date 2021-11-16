package pump

import (
	"time"

	"github.com/tidepool-org/platform/structure"
)

const (
	ReservoirRemainingUnitsMaximum = 10000
	ReservoirRemainingUnitsMinimum = 0
	ReservoirUnitsUnits            = "Units"
)

func ReservoirUnits() []string {
	return []string{
		ReservoirUnitsUnits,
	}
}

type Reservoir struct {
	Time      *time.Time `json:"time,omitempty" bson:"time,omitempty"`
	Remaining *float64   `json:"remaining,omitempty" bson:"remaining,omitempty"`
	Units     *string    `json:"units,omitempty" bson:"units,omitempty"`
}

func ParseReservoir(parser structure.ObjectParser) *Reservoir {
	if !parser.Exists() {
		return nil
	}
	datum := NewReservoir()
	parser.Parse(datum)
	return datum
}

func NewReservoir() *Reservoir {
	return &Reservoir{}
}

func (r *Reservoir) Parse(parser structure.ObjectParser) {
	r.Time = parser.Time("time", TimeFormat)
	r.Remaining = parser.Float64("remaining")
	r.Units = parser.String("units")
}

func (r *Reservoir) Validate(validator structure.Validator) {
	validator.Float64("remaining", r.Remaining).Exists().InRange(ReservoirRemainingUnitsMinimum, ReservoirRemainingUnitsMaximum)
	validator.String("units", r.Units).Exists().OneOf(ReservoirUnits()...)
}
