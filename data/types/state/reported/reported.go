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
	r.Type = Type

	r.States = nil
}

func (r *Reported) Parse(parser data.ObjectParser) error {
	parser.SetMeta(r.Meta())

	if err := r.Base.Parse(parser); err != nil {
		return err
	}

	r.States = ParseStateArray(parser.NewChildArrayParser("states"))

	return nil
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

func (r *Reported) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(r.Meta())
	}

	r.Base.Normalize(normalizer)

	if r.States != nil {
		r.States.Normalize(normalizer.WithReference("states"))
	}
}
