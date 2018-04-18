package calculator

import (
	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types"
	dataTypesBolusCombination "github.com/tidepool-org/platform/data/types/bolus/combination"
	dataTypesBolusExtended "github.com/tidepool-org/platform/data/types/bolus/extended"
	dataTypesBolusFactory "github.com/tidepool-org/platform/data/types/bolus/factory"
	dataTypesBolusNormal "github.com/tidepool-org/platform/data/types/bolus/normal"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	Type = "wizard" // TODO: Rename Type to "calculator"

	CarbohydrateInputMaximum        = 1000.0
	CarbohydrateInputMinimum        = 0.0
	InsulinCarbohydrateRatioMaximum = 250.0
	InsulinCarbohydrateRatioMinimum = 0.0
	InsulinOnBoardMaximum           = 250.0
	InsulinOnBoardMinimum           = 0.0
)

type Calculator struct {
	types.Base `bson:",inline"`

	BloodGlucoseInput        *float64                 `json:"bgInput,omitempty" bson:"bgInput,omitempty"`
	BloodGlucoseTarget       *dataBloodGlucose.Target `json:"bgTarget,omitempty" bson:"bgTarget,omitempty"`
	Bolus                    *data.Datum              `json:"-" bson:"-"`
	BolusID                  *string                  `json:"bolus,omitempty" bson:"bolus,omitempty"`
	CarbohydrateInput        *float64                 `json:"carbInput,omitempty" bson:"carbInput,omitempty"`
	InsulinCarbohydrateRatio *float64                 `json:"insulinCarbRatio,omitempty" bson:"insulinCarbRatio,omitempty"`
	InsulinOnBoard           *float64                 `json:"insulinOnBoard,omitempty" bson:"insulinOnBoard,omitempty"`
	InsulinSensitivity       *float64                 `json:"insulinSensitivity,omitempty" bson:"insulinSensitivity,omitempty"`
	Recommended              *Recommended             `json:"recommended,omitempty" bson:"recommended,omitempty"`
	Units                    *string                  `json:"units,omitempty" bson:"units,omitempty"`
}

func New() *Calculator {
	return &Calculator{
		Base: types.New(Type),
	}
}

func (c *Calculator) Parse(parser data.ObjectParser) error {
	parser.SetMeta(c.Meta())

	if err := c.Base.Parse(parser); err != nil {
		return err
	}

	c.BloodGlucoseInput = parser.ParseFloat("bgInput")
	c.BloodGlucoseTarget = dataBloodGlucose.ParseTarget(parser.NewChildObjectParser("bgTarget"))
	c.CarbohydrateInput = parser.ParseFloat("carbInput")
	c.InsulinCarbohydrateRatio = parser.ParseFloat("insulinCarbRatio")
	c.InsulinOnBoard = parser.ParseFloat("insulinOnBoard")
	c.InsulinSensitivity = parser.ParseFloat("insulinSensitivity")
	c.Recommended = ParseRecommended(parser.NewChildObjectParser("recommended"))
	c.Units = parser.ParseString("units")
	c.Bolus = dataTypesBolusFactory.ParseBolusDatum(parser.NewChildObjectParser("bolus"))

	return nil
}

func (c *Calculator) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(c.Meta())
	}

	c.Base.Validate(validator)

	if c.Type != "" {
		validator.String("type", &c.Type).EqualTo(Type)
	}

	units := c.Units

	validator.Float64("bgInput", c.BloodGlucoseInput).InRange(dataBloodGlucose.ValueRangeForUnits(units))
	if c.BloodGlucoseTarget != nil {
		c.BloodGlucoseTarget.Validate(validator.WithReference("bgTarget"), units)
	}

	if validator.Origin() == structure.OriginExternal {
		if c.Bolus != nil {
			(*c.Bolus).Validate(validator.WithReference("bolus"))
		}
		validator.String("bolusId", c.BolusID).NotExists()
	} else {
		if c.Bolus != nil {
			validator.WithReference("bolus").ReportError(structureValidator.ErrorValueExists())
		}
		validator.String("bolusId", c.BolusID).Using(id.Validate)
	}

	validator.Float64("carbInput", c.CarbohydrateInput).InRange(CarbohydrateInputMinimum, CarbohydrateInputMaximum)
	validator.Float64("insulinCarbRatio", c.InsulinCarbohydrateRatio).InRange(InsulinCarbohydrateRatioMinimum, InsulinCarbohydrateRatioMaximum)
	validator.Float64("insulinOnBoard", c.InsulinOnBoard).InRange(InsulinOnBoardMinimum, InsulinOnBoardMaximum)
	validator.Float64("insulinSensitivity", c.InsulinSensitivity).InRange(dataBloodGlucose.ValueRangeForUnits(units))
	if c.Recommended != nil {
		c.Recommended.Validate(validator.WithReference("recommended"))
	}
	validator.String("units", c.Units).Exists().OneOf(dataBloodGlucose.Units()...)
}

func (c *Calculator) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(c.Meta())
	}

	c.Base.Normalize(normalizer)

	if normalizer.Origin() == structure.OriginExternal {
		c.BloodGlucoseInput = dataBloodGlucose.NormalizeValueForUnits(c.BloodGlucoseInput, c.Units)
	}

	if c.BloodGlucoseTarget != nil {
		c.BloodGlucoseTarget.Normalize(normalizer.WithReference("bgTarget"), c.Units)
	}

	if c.Bolus != nil {
		(*c.Bolus).Normalize(normalizer.WithReference("bolus"))
	}

	if normalizer.Origin() == structure.OriginExternal {
		if c.Bolus != nil {
			normalizer.AddData(*c.Bolus)
			switch bolus := (*c.Bolus).(type) {
			case *dataTypesBolusCombination.Combination:
				c.BolusID = bolus.ID
			case *dataTypesBolusExtended.Extended:
				c.BolusID = bolus.ID
			case *dataTypesBolusNormal.Normal:
				c.BolusID = bolus.ID
			}
			c.Bolus = nil
		}
		c.InsulinSensitivity = dataBloodGlucose.NormalizeValueForUnits(c.InsulinSensitivity, c.Units)
	}

	if c.Recommended != nil {
		c.Recommended.Normalize(normalizer.WithReference("recommended"))
	}

	if normalizer.Origin() == structure.OriginExternal {
		c.Units = dataBloodGlucose.NormalizeUnits(c.Units)
	}
}
