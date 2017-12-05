package continuous

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
)

type Continuous struct {
	glucose.Glucose `bson:",inline"`
}

func Type() string {
	return "cbg"
}

func NewDatum() data.Datum {
	return New()
}

func New() *Continuous {
	return &Continuous{}
}

func Init() *Continuous {
	continuous := New()
	continuous.Init()
	return continuous
}

func (c *Continuous) Init() {
	c.Glucose.Init()
	c.Type = Type()
}

func (c *Continuous) Validate(validator data.Validator) error {
	if err := c.Glucose.Validate(validator); err != nil {
		return err
	}

	validator.ValidateString("type", &c.Type).EqualTo(Type())

	return nil
}

func (c *Continuous) Normalize(normalizer data.Normalizer) {
	normalizer = normalizer.WithMeta(c.Meta())

	c.Glucose.Normalize(normalizer)
}
