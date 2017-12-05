package reported

import "github.com/tidepool-org/platform/data"

const (
	StateAlcohol               = "alcohol"
	StateCycle                 = "cycle"
	StateHyperglycemiaSymptoms = "hyperglycemiaSymptoms"
	StateHypoglycemiaSymptoms  = "hypoglycemiaSymptoms"
	StateIllness               = "illness"
	StateStress                = "stress"
)

type State struct {
	State *string `json:"state,omitempty" bson:"state,omitempty"`
}

func NewState() *State {
	return &State{}
}

func (s *State) Parse(parser data.ObjectParser) {
	s.State = parser.ParseString("state")
}

func (s *State) Validate(validator data.Validator) {
	validator.ValidateString("state", s.State).Exists().OneOf([]string{StateAlcohol, StateCycle, StateHyperglycemiaSymptoms, StateHypoglycemiaSymptoms, StateIllness, StateStress})
}

func (s *State) Normalize(normalizer data.Normalizer) {}

func ParseState(parser data.ObjectParser) *State {
	if parser.Object() == nil {
		return nil
	}

	state := NewState()
	state.Parse(parser)
	parser.ProcessNotParsed()

	return state
}

func ParseStates(parser data.ArrayParser) *[]*State {
	if parser.Array() == nil {
		return nil
	}

	states := &[]*State{}
	for index := range *parser.Array() {
		*states = append(*states, ParseState(parser.NewChildObjectParser(index)))
	}
	parser.ProcessNotParsed()

	return states
}
