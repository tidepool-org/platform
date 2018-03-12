package pump

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/structure"
)

const (
	Type = "pumpSettings"
)

type Pump struct {
	types.Base `bson:",inline"`

	ActiveScheduleName   *string                  `json:"activeSchedule,omitempty" bson:"activeSchedule,omitempty"` // TODO: Rename to activeScheduleName
	BasalSchedules       *BasalScheduleArrayMap   `json:"basalSchedules,omitempty" bson:"basalSchedules,omitempty"`
	BloodGlucoseTargets  *BloodGlucoseTargetArray `json:"bgTarget,omitempty" bson:"bgTarget,omitempty"`
	CarbohydrateRatios   *CarbohydrateRatioArray  `json:"carbRatio,omitempty" bson:"carbRatio,omitempty"`
	InsulinSensitivities *InsulinSensitivityArray `json:"insulinSensitivity,omitempty" bson:"insulinSensitivity,omitempty"`
	Units                *Units                   `json:"units,omitempty" bson:"units,omitempty"`
}

func New() *Pump {
	return &Pump{
		Base: types.New(Type),
	}
}

func (p *Pump) Parse(parser data.ObjectParser) error {
	parser.SetMeta(p.Meta())

	if err := p.Base.Parse(parser); err != nil {
		return err
	}

	p.ActiveScheduleName = parser.ParseString("activeSchedule")
	p.BasalSchedules = ParseBasalScheduleArrayMap(parser.NewChildObjectParser("basalSchedules"))
	p.BloodGlucoseTargets = ParseBloodGlucoseTargetArray(parser.NewChildArrayParser("bgTarget"))
	p.CarbohydrateRatios = ParseCarbohydrateRatioArray(parser.NewChildArrayParser("carbRatio"))
	p.InsulinSensitivities = ParseInsulinSensitivityArray(parser.NewChildArrayParser("insulinSensitivity"))
	p.Units = ParseUnits(parser.NewChildObjectParser("units"))

	return nil
}

func (p *Pump) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(p.Meta())
	}

	p.Base.Validate(validator)

	if p.Type != "" {
		validator.String("type", &p.Type).EqualTo(Type)
	}

	var unitsBloodGlucose *string
	if p.Units != nil {
		unitsBloodGlucose = p.Units.BloodGlucose
	}

	validator.String("activeSchedule", p.ActiveScheduleName).Exists().NotEmpty()
	if p.BasalSchedules != nil {
		p.BasalSchedules.Validate(validator.WithReference("basalSchedules"))
	}
	if p.BloodGlucoseTargets != nil {
		p.BloodGlucoseTargets.Validate(validator.WithReference("bgTarget"), unitsBloodGlucose)
	}
	if p.CarbohydrateRatios != nil {
		p.CarbohydrateRatios.Validate(validator.WithReference("carbRatio"))
	}
	if p.InsulinSensitivities != nil {
		p.InsulinSensitivities.Validate(validator.WithReference("insulinSensitivity"), unitsBloodGlucose)
	}
	if p.Units != nil {
		p.Units.Validate(validator.WithReference("units"))
	}
}

func (p *Pump) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(p.Meta())
	}

	p.Base.Normalize(normalizer)

	var unitsBloodGlucose *string
	if p.Units != nil {
		unitsBloodGlucose = p.Units.BloodGlucose
	}

	if p.BasalSchedules != nil {
		p.BasalSchedules.Normalize(normalizer.WithReference("basalSchedules"))
	}
	if p.BloodGlucoseTargets != nil {
		p.BloodGlucoseTargets.Normalize(normalizer.WithReference("bgTarget"), unitsBloodGlucose)
	}
	if p.CarbohydrateRatios != nil {
		p.CarbohydrateRatios.Normalize(normalizer.WithReference("carbRatio"))
	}
	if p.InsulinSensitivities != nil {
		p.InsulinSensitivities.Normalize(normalizer.WithReference("insulinSensitivity"), unitsBloodGlucose)
	}
	if p.Units != nil {
		p.Units.Normalize(normalizer.WithReference("units"))
	}
}
