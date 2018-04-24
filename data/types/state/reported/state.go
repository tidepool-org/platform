package reported

import (
	"strconv"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	StateAlcohol               = "alcohol"
	StateCycle                 = "cycle"
	StateHyperglycemiaSymptoms = "hyperglycemiaSymptoms"
	StateHypoglycemiaSymptoms  = "hypoglycemiaSymptoms"
	StateIllness               = "illness"
	StateStress                = "stress"
)

func States() []string {
	return []string{
		StateAlcohol,
		StateCycle,
		StateHyperglycemiaSymptoms,
		StateHypoglycemiaSymptoms,
		StateIllness,
		StateStress,
	}
}

type State struct {
	State *string `json:"state,omitempty" bson:"state,omitempty"`
}

func ParseState(parser data.ObjectParser) *State {
	if parser.Object() == nil {
		return nil
	}
	state := NewState()
	state.Parse(parser)
	parser.ProcessNotParsed()
	return state
}

func NewState() *State {
	return &State{}
}

func (s *State) Parse(parser data.ObjectParser) {
	s.State = parser.ParseString("state")
}

func (s *State) Validate(validator structure.Validator) {
	validator.String("state", s.State).Exists().OneOf(States()...)
}

func (s *State) Normalize(normalizer data.Normalizer) {}

type StateArray []*State

func ParseStateArray(parser data.ArrayParser) *StateArray {
	if parser.Array() == nil {
		return nil
	}
	stateArray := NewStateArray()
	stateArray.Parse(parser)
	parser.ProcessNotParsed()
	return stateArray
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
