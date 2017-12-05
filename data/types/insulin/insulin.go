package insulin

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
)

type Insulin struct {
	types.Base `bson:",inline"`

	Dose *Dose `json:"dose,omitempty" bson:"dose,omitempty"`
}

func Type() string {
	return "insulin"
}

func NewDatum() data.Datum {
	return New()
}

func New() *Insulin {
	return &Insulin{}
}

func Init() *Insulin {
	insulin := New()
	insulin.Init()
	return insulin
}

func (i *Insulin) Init() {
	i.Base.Init()
	i.Type = Type()

	i.Dose = nil
}

func (i *Insulin) Parse(parser data.ObjectParser) error {
	parser.SetMeta(i.Meta())

	if err := i.Base.Parse(parser); err != nil {
		return err
	}

	i.Dose = ParseDose(parser.NewChildObjectParser("dose"))

	return nil
}

func (i *Insulin) Validate(validator data.Validator) error {
	validator.SetMeta(i.Meta())

	if err := i.Base.Validate(validator); err != nil {
		return err
	}

	validator.ValidateString("type", &i.Type).EqualTo(Type())
	if i.Dose != nil {
		i.Dose.Validate(validator.NewChildValidator("dose"))
	}

	return nil
}

func (i *Insulin) Normalize(normalizer data.Normalizer) {
	normalizer = normalizer.WithMeta(i.Meta())

	i.Base.Normalize(normalizer)

	if i.Dose != nil {
		i.Dose.Normalize(normalizer.WithReference("dose"))
	}
}
