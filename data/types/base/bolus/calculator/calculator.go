package calculator

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/bloodglucose"
	"github.com/tidepool-org/platform/data/types/base"
	"github.com/tidepool-org/platform/data/types/base/bolus"
	"github.com/tidepool-org/platform/data/types/base/bolus/combination"
	"github.com/tidepool-org/platform/data/types/base/bolus/extended"
	"github.com/tidepool-org/platform/data/types/base/bolus/normal"
	"github.com/tidepool-org/platform/service"
)

type Calculator struct {
	base.Base `bson:",inline"`

	Recommended        *Recommended         `json:"recommended,omitempty" bson:"recommended,omitempty"`
	BloodGlucoseTarget *bloodglucose.Target `json:"bgTarget,omitempty" bson:"bgTarget,omitempty"`

	BolusID                  *string  `json:"bolus,omitempty" bson:"bolus,omitempty"`
	CarbohydrateInput        *int     `json:"carbInput,omitempty" bson:"carbInput,omitempty"`
	InsulinOnBoard           *float64 `json:"insulinOnBoard,omitempty" bson:"insulinOnBoard,omitempty"`
	InsulinSensitivity       *float64 `json:"insulinSensitivity,omitempty" bson:"insulinSensitivity,omitempty"`
	InsulinCarbohydrateRatio *int     `json:"insulinCarbRatio,omitempty" bson:"insulinCarbRatio,omitempty"`
	BloodGlucoseInput        *float64 `json:"bgInput,omitempty" bson:"bgInput,omitempty"`
	Units                    *string  `json:"units,omitempty" bson:"units,omitempty"`

	// Embedded bolus
	bolus *data.Datum
}

func Type() string {
	return "wizard"
}

func NewDatum() data.Datum {
	return New()
}

func New() *Calculator {
	return &Calculator{}
}

func Init() *Calculator {
	calculator := New()
	calculator.Init()
	return calculator
}

func (c *Calculator) Init() {
	c.Base.Init()
	c.Base.Type = Type()

	c.Recommended = nil
	c.BloodGlucoseTarget = nil

	c.BolusID = nil
	c.CarbohydrateInput = nil
	c.InsulinOnBoard = nil
	c.InsulinSensitivity = nil
	c.InsulinCarbohydrateRatio = nil
	c.BloodGlucoseInput = nil
	c.Units = nil

	c.bolus = nil
}

func (c *Calculator) Parse(parser data.ObjectParser) error {
	parser.SetMeta(c.Meta())

	if err := c.Base.Parse(parser); err != nil {
		return err
	}

	c.CarbohydrateInput = parser.ParseInteger("carbInput")
	c.InsulinOnBoard = parser.ParseFloat("insulinOnBoard")
	c.InsulinSensitivity = parser.ParseFloat("insulinSensitivity")
	c.InsulinCarbohydrateRatio = parser.ParseInteger("insulinCarbRatio")
	c.BloodGlucoseInput = parser.ParseFloat("bgInput")
	c.Units = parser.ParseString("units")

	c.Recommended = ParseRecommended(parser.NewChildObjectParser("recommended"))
	c.BloodGlucoseTarget = bloodglucose.ParseTarget(parser.NewChildObjectParser("bgTarget"))

	// TODO: This is a bit hacky to ensure we only parse true bolus data. Is there a better way?

	if bolusParser := parser.NewChildObjectParser("bolus"); bolusParser.Object() != nil {
		if bolusType := bolusParser.ParseString("type"); bolusType == nil {
			bolusParser.AppendError("type", service.ErrorValueNotExists())
		} else if *bolusType != bolus.Type() {
			bolusParser.AppendError("type", service.ErrorValueStringNotOneOf(*bolusType, []string{bolus.Type()}))
		} else {
			c.bolus = parser.ParseDatum("bolus")
		}
	}

	return nil
}

func (c *Calculator) Validate(validator data.Validator) error {
	validator.SetMeta(c.Meta())

	if err := c.Base.Validate(validator); err != nil {
		return err
	}

	validator.ValidateInteger("carbInput", c.CarbohydrateInput).InRange(0, 1000)
	validator.ValidateFloat("insulinOnBoard", c.InsulinOnBoard).InRange(0.0, 250.0)
	validator.ValidateInteger("insulinCarbRatio", c.InsulinCarbohydrateRatio).InRange(0, 250)

	validator.ValidateStringAsBloodGlucoseUnits("units", c.Units).Exists()
	validator.ValidateFloatAsBloodGlucoseValue("bgInput", c.BloodGlucoseInput).InRangeForUnits(c.Units)
	validator.ValidateFloatAsBloodGlucoseValue("insulinSensitivity", c.InsulinSensitivity).InRangeForUnits(c.Units)

	if c.Recommended != nil {
		c.Recommended.Validate(validator.NewChildValidator("recommended"))
	}

	if c.BloodGlucoseTarget != nil {
		c.BloodGlucoseTarget.Validate(validator.NewChildValidator("bgTarget"), c.Units)
	}

	if c.bolus != nil {
		(*c.bolus).Validate(validator.NewChildValidator("bolus"))
	}

	return nil
}

func (c *Calculator) Normalize(normalizer data.Normalizer) error {
	normalizer.SetMeta(c.Meta())

	if err := c.Base.Normalize(normalizer); err != nil {
		return err
	}

	units := c.Units

	bloodGlucoseNormalizer := normalizer.NormalizeBloodGlucose(c.Units)
	c.Units = bloodGlucoseNormalizer.Units()
	c.InsulinSensitivity = bloodGlucoseNormalizer.Value(c.InsulinSensitivity)
	c.BloodGlucoseInput = bloodGlucoseNormalizer.Value(c.BloodGlucoseInput)

	if c.Recommended != nil {
		c.Recommended.Normalize(normalizer.NewChildNormalizer("recommended"))
	}

	if c.BloodGlucoseTarget != nil {
		c.BloodGlucoseTarget.Normalize(normalizer.NewChildNormalizer("bgTarget"), units)
	}

	if c.bolus != nil {
		if err := (*c.bolus).Normalize(normalizer.NewChildNormalizer("bolus")); err != nil {
			return err
		}

		switch (*c.bolus).(type) {
		case *extended.Extended:
			c.BolusID = &(*c.bolus).(*extended.Extended).ID
		case *normal.Normal:
			c.BolusID = &(*c.bolus).(*normal.Normal).ID
		case *combination.Combination:
			c.BolusID = &(*c.bolus).(*combination.Combination).ID
		}

		normalizer.AppendDatum(*c.bolus)
		c.bolus = nil
	}

	return nil
}
