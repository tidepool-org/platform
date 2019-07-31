package physical

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	StepCountMaximum = 100000
	StepCountMinimum = 0
)

type Step struct {
	Count *int `json:"count,omitempty" bson:"count,omitempty"`
}

func ParseStep(parser structure.ObjectParser) *Step {
	if !parser.Exists() {
		return nil
	}
	datum := NewStep()
	parser.Parse(datum)
	return datum
}

func NewStep() *Step {
	return &Step{}
}

func (s *Step) Parse(parser structure.ObjectParser) {
	s.Count = parser.Int("count")
}

func (s *Step) Validate(validator structure.Validator) {
	validator.Int("count", s.Count).Exists().InRange(StepCountMinimum, StepCountMaximum)
}

func (s *Step) Normalize(normalizer data.Normalizer) {}
