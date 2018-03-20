package reported

import (
	"strconv"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	StateSeverityMaximum            = 10
	StateSeverityMinimum            = 0
	StateStateAlcohol               = "alcohol"
	StateStateCycle                 = "cycle"
	StateStateHyperglycemiaSymptoms = "hyperglycemiaSymptoms"
	StateStateHypoglycemiaSymptoms  = "hypoglycemiaSymptoms"
	StateStateIllness               = "illness"
	StateStateOther                 = "other"
	StateStateOtherLengthMaximum    = 100
	StateStateStress                = "stress"
)

func StateStates() []string {
	return []string{
		StateStateAlcohol,
		StateStateCycle,
		StateStateHyperglycemiaSymptoms,
		StateStateHypoglycemiaSymptoms,
		StateStateIllness,
		StateStateOther,
		StateStateStress,
	}
}

type State struct {
	Severity   *int    `json:"severity,omitempty" bson:"severity,omitempty"`
	State      *string `json:"state,omitempty" bson:"state,omitempty"`
	StateOther *string `json:"stateOther,omitempty" bson:"stateOther,omitempty"`
}

func ParseState(parser data.ObjectParser) *State {
	if parser.Object() == nil {
		return nil
	}
	datum := NewState()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewState() *State {
	return &State{}
}

func (s *State) Parse(parser data.ObjectParser) {
	s.Severity = parser.ParseInteger("severity")
	s.State = parser.ParseString("state")
	s.StateOther = parser.ParseString("stateOther")
}

func (s *State) Validate(validator structure.Validator) {
	validator.Int("severity", s.Severity).InRange(StateSeverityMinimum, StateSeverityMaximum)
	validator.String("state", s.State).Exists().OneOf(StateStates()...)
	if s.State != nil && *s.State == StateStateOther {
		validator.String("stateOther", s.StateOther).Exists().NotEmpty().LengthLessThanOrEqualTo(StateStateOtherLengthMaximum)
	} else {
		validator.String("stateOther", s.StateOther).NotExists()
	}
}

func (s *State) Normalize(normalizer data.Normalizer) {}

type StateArray []*State

func ParseStateArray(parser data.ArrayParser) *StateArray {
	if parser.Array() == nil {
		return nil
	}
	datum := NewStateArray()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewStateArray() *StateArray {
	return &StateArray{}
}

func (s *StateArray) Parse(parser data.ArrayParser) {
	for index := range *parser.Array() {
		*s = append(*s, ParseState(parser.NewChildObjectParser(index)))
	}
}

func (s *StateArray) Validate(validator structure.Validator) {
	// TODO: Validate no duplicates?
	for index, state := range *s {
		stateValidator := validator.WithReference(strconv.Itoa(index))
		if state != nil {
			state.Validate(stateValidator)
		} else {
			stateValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (s *StateArray) Normalize(normalizer data.Normalizer) {
	for index, state := range *s {
		if state != nil {
			state.Normalize(normalizer.WithReference(strconv.Itoa(index)))
		}
	}
}
