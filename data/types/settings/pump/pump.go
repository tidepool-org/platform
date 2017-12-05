package pump

import (
	"strconv"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
)

type Pump struct {
	types.Base `bson:",inline"`

	*Units `json:"units,omitempty" bson:"units,omitempty"`

	BasalSchedules *map[string]*[]*BasalSchedule `json:"basalSchedules,omitempty" bson:"basalSchedules,omitempty"`

	CarbohydrateRatios   *[]*CarbohydrateRatio  `json:"carbRatio,omitempty" bson:"carbRatio,omitempty"`
	InsulinSensitivities *[]*InsulinSensitivity `json:"insulinSensitivity,omitempty" bson:"insulinSensitivity,omitempty"`
	BloodGlucoseTargets  *[]*BloodGlucoseTarget `json:"bgTarget,omitempty" bson:"bgTarget,omitempty"`

	ActiveSchedule *string `json:"activeSchedule,omitempty" bson:"activeSchedule,omitempty"`
}

func Type() string {
	return "pumpSettings"
}

func NewDatum() data.Datum {
	return New()
}

func New() *Pump {
	return &Pump{}
}

func Init() *Pump {
	pump := New()
	pump.Init()
	return pump
}

func (p *Pump) Init() {
	p.Base.Init()
	p.Type = Type()

	p.Units = nil

	p.BasalSchedules = nil

	p.CarbohydrateRatios = nil
	p.InsulinSensitivities = nil
	p.BloodGlucoseTargets = nil

	p.ActiveSchedule = nil
}

func (p *Pump) Parse(parser data.ObjectParser) error {
	parser.SetMeta(p.Meta())

	if err := p.Base.Parse(parser); err != nil {
		return err
	}

	p.ActiveSchedule = parser.ParseString("activeSchedule")

	p.Units = ParseUnits(parser.NewChildObjectParser("units"))

	p.CarbohydrateRatios = ParseCarbohydrateRatioArray(parser.NewChildArrayParser("carbRatio"))
	p.InsulinSensitivities = ParseInsulinSensitivityArray(parser.NewChildArrayParser("insulinSensitivity"))
	p.BloodGlucoseTargets = ParseBloodGlucoseTargetArray(parser.NewChildArrayParser("bgTarget"))
	p.BasalSchedules = ParseBasalSchedulesMap(parser.NewChildObjectParser("basalSchedules"))

	return nil
}

func (p *Pump) Validate(validator data.Validator) error {
	validator.SetMeta(p.Meta())

	if err := p.Base.Validate(validator); err != nil {
		return err
	}

	validator.ValidateString("type", &p.Type).EqualTo(Type())

	validator.ValidateString("activeSchedule", p.ActiveSchedule).Exists().NotEmpty()

	if p.Units != nil {
		p.Units.Validate(validator.NewChildValidator("units"))
	}

	if p.CarbohydrateRatios != nil {
		carbohydrateRatiosValidator := validator.NewChildValidator("carbRatio")
		for index, carbohydrateRatio := range *p.CarbohydrateRatios {
			if carbohydrateRatio != nil {
				carbohydrateRatio.Validate(carbohydrateRatiosValidator.NewChildValidator(index))
			}
		}
	}

	if p.InsulinSensitivities != nil {
		insulinSensitivitiesValidator := validator.NewChildValidator("insulinSensitivity")
		for index, insulinSensitivity := range *p.InsulinSensitivities {
			if insulinSensitivity != nil {
				insulinSensitivity.Validate(insulinSensitivitiesValidator.NewChildValidator(index), p.Units.BloodGlucose)
			}
		}
	}

	if p.BloodGlucoseTargets != nil {
		bloodGlucoseTargetsValidator := validator.NewChildValidator("bgTarget")
		for index, bgTarget := range *p.BloodGlucoseTargets {
			if bgTarget != nil {
				bgTarget.Validate(bloodGlucoseTargetsValidator.NewChildValidator(index), p.Units.BloodGlucose)
			}
		}
	}

	if p.BasalSchedules != nil {
		basalSchedulesValidator := validator.NewChildValidator("basalSchedules")
		for basalScheduleName, basalSchedule := range *p.BasalSchedules {
			basalSchedulesValidator.ValidateString("", &basalScheduleName).Exists().NotEmpty()
			if basalSchedule != nil {
				basalScheduleValidator := basalSchedulesValidator.NewChildValidator(basalScheduleName)
				for index, scheduleItem := range *basalSchedule {
					scheduleItem.Validate(basalScheduleValidator.NewChildValidator(index))
				}
			}
		}
	}

	return nil
}

func (p *Pump) Normalize(normalizer data.Normalizer) {
	normalizer = normalizer.WithMeta(p.Meta())

	p.Base.Normalize(normalizer)

	var originalUnits *string

	if p.Units != nil {
		originalUnits = p.Units.BloodGlucose
		p.Units.Normalize(normalizer.WithReference("units"))
	}

	if p.BasalSchedules != nil {
		basalSchedulesNormalizer := normalizer.WithReference("basalSchedules")
		for basalScheduleName, basalSchedule := range *p.BasalSchedules {
			if basalSchedule != nil {
				basalScheduleNormalizer := basalSchedulesNormalizer.WithReference(basalScheduleName)
				for index, scheduleItem := range *basalSchedule {
					if scheduleItem != nil {
						scheduleItem.Normalize(basalScheduleNormalizer.WithReference(strconv.Itoa(index)))
					}
				}
			}
		}
	}

	if p.CarbohydrateRatios != nil {
		carbohydrateRatiosNormalizer := normalizer.WithReference("carbRatio")
		for index, carbohydrateRatio := range *p.CarbohydrateRatios {
			if carbohydrateRatio != nil {
				carbohydrateRatio.Normalize(carbohydrateRatiosNormalizer.WithReference(strconv.Itoa(index)))
			}
		}
	}

	if p.InsulinSensitivities != nil {
		insulinSensitivitiesNormalizer := normalizer.WithReference("insulinSensitivity")
		for index, insulinSensitivity := range *p.InsulinSensitivities {
			if insulinSensitivity != nil {
				insulinSensitivity.Normalize(insulinSensitivitiesNormalizer.WithReference(strconv.Itoa(index)), originalUnits)
			}
		}
	}

	if p.BloodGlucoseTargets != nil {
		bloodGlucoseTargetsNormalizer := normalizer.WithReference("bgTarget")
		for index, bgTarget := range *p.BloodGlucoseTargets {
			if bgTarget != nil {
				bgTarget.Normalize(bloodGlucoseTargetsNormalizer.WithReference(strconv.Itoa(index)), originalUnits)
			}
		}
	}
}
