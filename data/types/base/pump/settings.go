package pump

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/base"
)

type Settings struct {
	base.Base `bson:",inline"`

	*Units `json:"units,omitempty" bson:"units,omitempty"`

	//TODO: sorting out validation
	//BasalSchedules *map[string][]*BasalSchedule `json:"basalSchedules,omitempty" bson:"basalSchedules,omitempty"`

	CarbohydrateRatios   *[]*CarbohydrateRatio  `json:"carbRatio,omitempty" bson:"carbRatio,omitempty"`
	InsulinSensitivities *[]*InsulinSensitivity `json:"insulinSensitivity,omitempty" bson:"insulinSensitivity,omitempty"`
	BloodGlucoseTargets  *[]*BloodGlucoseTarget `json:"bgTarget,omitempty" bson:"bgTarget,omitempty"`

	ActiveSchedule *string `json:"activeSchedule,omitempty" bson:"activeSchedule,omitempty"`
}

func Type() string {
	return "pumpSettings"
}

func New() (*Settings, error) {
	settingsBase, err := base.New(Type())
	if err != nil {
		return nil, err
	}

	return &Settings{
		Base: *settingsBase,
	}, nil
}

func (s *Settings) Parse(parser data.ObjectParser) {
	s.Base.Parse(parser)

	s.ActiveSchedule = parser.ParseString("activeSchedule")

	s.Units = ParseUnits(parser.NewChildObjectParser("units"))

	s.CarbohydrateRatios = ParseCarbohydrateRatioArray(parser.NewChildArrayParser("carbRatio"))
	s.InsulinSensitivities = ParseInsulinSensitivityArray(parser.NewChildArrayParser("insulinSensitivity"))
	s.BloodGlucoseTargets = ParseBloodGlucoseTargetArray(parser.NewChildArrayParser("bgTarget"))

	//s.BasalSchedules = ParseBasalScheduleArray(parser.NewChildArrayParser("basalSchedules"))
}

func (s *Settings) Validate(validator data.Validator) {
	s.Base.Validate(validator)

	validator.ValidateString("activeSchedule", s.ActiveSchedule).Exists().LengthGreaterThanOrEqualTo(1)

	if s.Units != nil {
		s.Units.Validate(validator.NewChildValidator("units"))
	}

	if s.CarbohydrateRatios != nil {
		carbohydrateRatiosValidator := validator.NewChildValidator("carbRatio")
		for index, carbohydrateRatio := range *s.CarbohydrateRatios {
			if carbohydrateRatio != nil {
				carbohydrateRatio.Validate(carbohydrateRatiosValidator.NewChildValidator(index))
			}
		}
	}

	if s.InsulinSensitivities != nil {
		insulinSensitivitiesValidator := validator.NewChildValidator("insulinSensitivity")
		for index, insulinSensitivity := range *s.InsulinSensitivities {
			if insulinSensitivity != nil {
				insulinSensitivity.amountUnits = s.Units.BloodGlucose
				insulinSensitivity.Validate(insulinSensitivitiesValidator.NewChildValidator(index))
			}
		}
	}

	if s.BloodGlucoseTargets != nil {
		bloodGlucoseTargetsValidator := validator.NewChildValidator("bgTarget")
		for index, bgTarget := range *s.BloodGlucoseTargets {
			if bgTarget != nil {
				bgTarget.targetUnits = s.Units.BloodGlucose
				bgTarget.Validate(bloodGlucoseTargetsValidator.NewChildValidator(index))
			}
		}
	}

}

func (s *Settings) Normalize(normalizer data.Normalizer) {
	s.Base.Normalize(normalizer)

	var originalUnits *string

	if s.Units != nil {
		originalUnits = s.Units.BloodGlucose
		s.Units.Normalize(normalizer.NewChildNormalizer("units"))
	}

	if s.BloodGlucoseTargets != nil {
		bloodGlucoseTargetsNormalizer := normalizer.NewChildNormalizer("bgTarget")
		for index, bgTarget := range *s.BloodGlucoseTargets {
			if bgTarget != nil {
				bgTarget.targetUnits = originalUnits
				bgTarget.Normalize(bloodGlucoseTargetsNormalizer.NewChildNormalizer(index))
			}
		}
	}

}
