package physical

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	LapCountMaximum = 10000
	LapCountMinimum = 0
)

type Lap struct {
	Count    *int      `json:"count,omitempty" bson:"count,omitempty"`
	Distance *Distance `json:"distance,omitempty" bson:"distance,omitempty"`
}

func ParseLap(parser structure.ObjectParser) *Lap {
	if !parser.Exists() {
		return nil
	}
	datum := NewLap()
	parser.Parse(datum)
	return datum
}

func NewLap() *Lap {
	return &Lap{}
}

func (l *Lap) Parse(parser structure.ObjectParser) {
	l.Count = parser.Int("count")
	l.Distance = ParseDistance(parser.WithReferenceObjectParser("distance"))
}

func (l *Lap) Validate(validator structure.Validator) {
	validator.Int("count", l.Count).Exists().InRange(LapCountMinimum, LapCountMaximum)
	if l.Distance != nil {
		l.Distance.Validate(validator.WithReference("distance"))
	} else {
		validator.WithReference("distance").ReportError(structureValidator.ErrorValueNotExists())
	}
}

func (l *Lap) Normalize(normalizer data.Normalizer) {
	if l.Distance != nil {
		l.Distance.Normalize(normalizer.WithReference("distance"))
	}
}
