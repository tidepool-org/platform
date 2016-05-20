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
	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/types/base"
	"github.com/tidepool-org/platform/pvn/data/types/common/bloodglucose"
)

type Calculator struct {
	base.Base `bson:",inline"`

	*Recommended        `json:"recommended,omitempty" bson:"recommended,omitempty"`
	*BloodGlucoseTarget `json:"bgTarget,omitempty" bson:"bgTarget,omitempty"`
	//*bolus.Bolus `json:"bolus,omitempty" bson:"bolus,omitempty"`

	BolusID                  *string  `json:"bolusId,omitempty" bson:"bolusId,omitempty"`
	CarbohydrateInput        *int     `json:"carbInput,omitempty" bson:"carbInput,omitempty"`
	InsulinOnBoard           *float64 `json:"insulinOnBoard,omitempty" bson:"insulinOnBoard,omitempty"`
	InsulinSensitivity       *float64 `json:"insulinSensitivity,omitempty" bson:"insulinSensitivity,omitempty"`
	InsulinCarbohydrateRatio *int     `json:"insulinCarbRatio,omitempty" bson:"insulinCarbRatio,omitempty"`
	BloodGlucoseInput        *float64 `json:"bgInput,omitempty" bson:"bgInput,omitempty"`
	Units                    *string  `json:"units" bson:"units"`
}

func Type() string {
	return "wizard"
}

func New() *Calculator {
	calculatorType := Type()

	calculator := &Calculator{}
	calculator.Type = &calculatorType
	return calculator
}

func (c *Calculator) Parse(parser data.ObjectParser) {
	c.Base.Parse(parser)
	c.CarbohydrateInput = parser.ParseInteger("carbInput")
	c.InsulinOnBoard = parser.ParseFloat("insulinOnBoard")
	c.InsulinSensitivity = parser.ParseFloat("insulinSensitivity")
	c.InsulinCarbohydrateRatio = parser.ParseInteger("insulinCarbRatio")
	c.BloodGlucoseInput = parser.ParseFloat("bgInput")
	c.Units = parser.ParseString("units")

	c.Recommended = ParseRecommended(parser.NewChildObjectParser("recommended"))
	c.BloodGlucoseTarget = ParseBloodGlucoseTarget(parser.NewChildObjectParser("bgTarget"))
}

func (c *Calculator) Validate(validator data.Validator) {
	c.Base.Validate(validator)
	validator.ValidateInteger("carbInput", c.CarbohydrateInput).InRange(0, 1000)
	validator.ValidateFloat("insulinOnBoard", c.InsulinOnBoard).InRange(0.0, 250.0)
	validator.ValidateInteger("insulinCarbRatio", c.InsulinCarbohydrateRatio).InRange(0, 250)

	validator.ValidateString("units", c.Units).Exists().OneOf([]string{common.Mmoll, common.MmolL, common.Mgdl, common.MgdL})
	switch c.Units {
	case &common.Mmoll, &common.MmolL:
		validator.ValidateFloat("bgInput", c.BloodGlucoseInput).InRange(common.MmolLFromValue, common.MmolLToValue)
		validator.ValidateFloat("insulinSensitivity", c.InsulinSensitivity).InRange(common.MmolLFromValue, common.MmolLToValue)
	default:
		validator.ValidateFloat("bgInput", c.BloodGlucoseInput).InRange(common.MgdLFromValue, common.MgdLToValue)
		validator.ValidateFloat("insulinSensitivity", c.InsulinSensitivity).InRange(common.MgdLFromValue, common.MgdLToValue)
	}

	if c.Recommended != nil {
		c.Recommended.Validate(validator.NewChildValidator("recommended"))
	}

	if c.BloodGlucoseTarget != nil {
		c.targetUnits = c.Units
		c.BloodGlucoseTarget.Validate(validator.NewChildValidator("bgTarget"))
	}
}

func (c *Calculator) Normalize(normalizer data.Normalizer) {
	c.Base.Normalize(normalizer)

	originalUnits := c.Units

	if c.BloodGlucoseTarget != nil {
		c.targetUnits = originalUnits
		c.BloodGlucoseTarget.Normalize(normalizer.NewChildNormalizer("bgTarget"))
	}

	bgNormalizer := normalizer.NormalizeBloodGlucose(Type(), c.Units)
	c.Units = bgNormalizer.NormalizeUnits()
	c.InsulinSensitivity = bgNormalizer.NormalizeValue(c.InsulinSensitivity)
	c.BloodGlucoseInput = bgNormalizer.NormalizeValue(c.BloodGlucoseInput)
}
