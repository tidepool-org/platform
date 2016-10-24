package continuous

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/base"
)

type Continuous struct {
	base.Base `bson:",inline"`

	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
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
	c.Base.Init()
	c.Base.Type = Type()

	c.Units = nil
	c.Value = nil
}

func (c *Continuous) Parse(parser data.ObjectParser) error {
	parser.SetMeta(c.Meta())

	if err := c.Base.Parse(parser); err != nil {
		return err
	}

	c.Units = parser.ParseString("units")
	c.Value = parser.ParseFloat("value")

	return nil
}

func (c *Continuous) Validate(validator data.Validator) error {
	validator.SetMeta(c.Meta())

	if err := c.Base.Validate(validator); err != nil {
		return err
	}

	validator.ValidateStringAsBloodGlucoseUnits("units", c.Units).Exists()
	validator.ValidateFloatAsBloodGlucoseValue("value", c.Value).Exists().InRangeForUnits(c.Units)

	return nil
}

func (c *Continuous) Normalize(normalizer data.Normalizer) error {
	normalizer.SetMeta(c.Meta())

	if err := c.Base.Normalize(normalizer); err != nil {
		return err
	}

	c.Units, c.Value = normalizer.NormalizeBloodGlucose(c.Units).UnitsAndValue(c.Value)

	return nil
}
