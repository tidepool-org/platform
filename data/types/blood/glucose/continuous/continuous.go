package continuous

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/structure"
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

func (c *Continuous) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(c.Meta())
	}

	c.Glucose.Validate(validator)

	if c.Type != "" {
		validator.String("type", &c.Type).EqualTo(Type())
	}
}

func (c *Continuous) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(c.Meta())
	}

	c.Glucose.Normalize(normalizer)
}
