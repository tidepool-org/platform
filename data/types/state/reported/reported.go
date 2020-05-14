package reported

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/structure"
)

const (
	Type = "reportedState" // TODO: Change to "state/reported"
)

type Reported struct {
	types.Base `bson:",inline"`

	States *StateArray `json:"states,omitempty" bson:"states,omitempty"`
}

func New() *Reported {
	return &Reported{
		Base: types.New(Type),
	}
}

func (r *Reported) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(r.Meta())
	}

	r.Base.Parse(parser)

	r.States = ParseStateArray(parser.WithReferenceArrayParser("states"))
}

func (r *Reported) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(r.Meta())
	}

	r.Base.Validate(validator)

	if r.Type != "" {
		validator.String("type", &r.Type).EqualTo(Type)
	}

	if r.States != nil {
		r.States.Validate(validator.WithReference("states"))
	}
}

// IsValid returns true if there is no error in the validator
func (r *Reported) IsValid(validator structure.Validator) bool {
	return !(validator.HasError())
}

func (r *Reported) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(r.Meta())
	}

	r.Base.Normalize(normalizer)

	if r.States != nil {
		r.States.Normalize(normalizer.WithReference("states"))
	}
}
