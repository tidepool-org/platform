package reported

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
)

type Reported struct {
	types.Base `bson:",inline"`

	States *[]*State `json:"states,omitempty" bson:"states,omitempty"`
}

func Type() string {
	return "reportedState"
}

func NewDatum() data.Datum {
	return New()
}

func New() *Reported {
	return &Reported{}
}

func Init() *Reported {
	reported := New()
	reported.Init()
	return reported
}

func (r *Reported) Init() {
	r.Base.Init()
	r.Type = Type()

	r.States = nil
}

func (r *Reported) Parse(parser data.ObjectParser) error {
	parser.SetMeta(r.Meta())

	if err := r.Base.Parse(parser); err != nil {
		return err
	}

	r.States = ParseStates(parser.NewChildArrayParser("states"))

	return nil
}

func (r *Reported) Validate(validator data.Validator) error {
	validator.SetMeta(r.Meta())

	if err := r.Base.Validate(validator); err != nil {
		return err
	}

	validator.ValidateString("type", &r.Type).EqualTo(Type())
	if r.States != nil {
		statesValidator := validator.NewChildValidator("states")
		for index, state := range *r.States {
			if state != nil {
				state.Validate(statesValidator.NewChildValidator(index))
			}
		}
	}

	return nil
}

func (r *Reported) Normalize(normalizer data.Normalizer) error {
	normalizer.SetMeta(r.Meta())

	if err := r.Base.Normalize(normalizer); err != nil {
		return err
	}

	if r.States != nil {
		statesNormalizer := normalizer.NewChildNormalizer("states")
		for index, state := range *r.States {
			if state != nil {
				state.Normalize(statesNormalizer.NewChildNormalizer(index))
			}
		}
	}

	return nil
}
