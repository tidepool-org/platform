package pump

import (
	"sort"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/structure"
)

const (
	Type = "pumpSettings"

	ManufacturerLengthMaximum  = 100
	ManufacturersLengthMaximum = 10
	ModelLengthMaximum         = 100
	SerialNumberLengthMaximum  = 100
)

type Pump struct {
	types.Base `bson:",inline"`

	ActiveScheduleName   *string                  `json:"activeSchedule,omitempty" bson:"activeSchedule,omitempty"` // TODO: Rename to activeScheduleName; move into basal struct
	Basal                *Basal                   `json:"basal,omitempty" bson:"basal,omitempty"`
	BasalSchedules       *BasalScheduleArrayMap   `json:"basalSchedules,omitempty" bson:"basalSchedules,omitempty"` // TODO: Move into basal struct
	BloodGlucoseTargets  *BloodGlucoseTargetArray `json:"bgTarget,omitempty" bson:"bgTarget,omitempty"`             // TODO: Move into bolus struct
	Bolus                *Bolus                   `json:"bolus,omitempty" bson:"bolus,omitempty"`
	CarbohydrateRatios   *CarbohydrateRatioArray  `json:"carbRatio,omitempty" bson:"carbRatio,omitempty"` // TODO: Move into bolus struct
	Display              *Display                 `json:"display,omitempty" bson:"display,omitempty"`
	Insulin              *Insulin                 `json:"insulin,omitempty" bson:"insulin,omitempty"`
	InsulinSensitivities *InsulinSensitivityArray `json:"insulinSensitivity,omitempty" bson:"insulinSensitivity,omitempty"` // TODO: Move into bolus struct
	Manufacturers        *[]string                `json:"manufacturers,omitempty" bson:"manufacturers,omitempty"`
	Model                *string                  `json:"model,omitempty" bson:"model,omitempty"`
	SerialNumber         *string                  `json:"serialNumber,omitempty" bson:"serialNumber,omitempty"`
	Units                *Units                   `json:"units,omitempty" bson:"units,omitempty"` // TODO: Move into all appropriate structs
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
	p.Basal = ParseBasal(parser.NewChildObjectParser("basal"))
	p.BasalSchedules = ParseBasalScheduleArrayMap(parser.NewChildObjectParser("basalSchedules"))
	p.BloodGlucoseTargets = ParseBloodGlucoseTargetArray(parser.NewChildArrayParser("bgTarget"))
	p.Bolus = ParseBolus(parser.NewChildObjectParser("bolus"))
	p.CarbohydrateRatios = ParseCarbohydrateRatioArray(parser.NewChildArrayParser("carbRatio"))
	p.Display = ParseDisplay(parser.NewChildObjectParser("display"))
	p.Insulin = ParseInsulin(parser.NewChildObjectParser("insulin"))
	p.InsulinSensitivities = ParseInsulinSensitivityArray(parser.NewChildArrayParser("insulinSensitivity"))
	p.Manufacturers = parser.ParseStringArray("manufacturers")
	p.Model = parser.ParseString("model")
	p.SerialNumber = parser.ParseString("serialNumber")
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
	if p.Basal != nil {
		p.Basal.Validate(validator.WithReference("basal"))
	}
	if p.BasalSchedules != nil {
		p.BasalSchedules.Validate(validator.WithReference("basalSchedules"))
	}
	if p.BloodGlucoseTargets != nil {
		p.BloodGlucoseTargets.Validate(validator.WithReference("bgTarget"), unitsBloodGlucose)
	}
	if p.Bolus != nil {
		p.Bolus.Validate(validator.WithReference("bolus"))
	}
	if p.CarbohydrateRatios != nil {
		p.CarbohydrateRatios.Validate(validator.WithReference("carbRatio"))
	}
	if p.Display != nil {
		p.Display.Validate(validator.WithReference("display"))
	}
	if p.Insulin != nil {
		p.Insulin.Validate(validator.WithReference("insulin"))
	}
	if p.InsulinSensitivities != nil {
		p.InsulinSensitivities.Validate(validator.WithReference("insulinSensitivity"), unitsBloodGlucose)
	}
	validator.StringArray("manufacturers", p.Manufacturers).NotEmpty().LengthLessThanOrEqualTo(ManufacturersLengthMaximum).Each(func(stringValidator structure.String) {
		stringValidator.Exists().NotEmpty().LengthLessThanOrEqualTo(ManufacturerLengthMaximum)
	}).EachUnique()
	validator.String("model", p.Model).NotEmpty().LengthLessThanOrEqualTo(ModelLengthMaximum)
	validator.String("serialNumber", p.SerialNumber).NotEmpty().LengthLessThanOrEqualTo(SerialNumberLengthMaximum)
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

	if p.Basal != nil {
		p.Basal.Normalize(normalizer.WithReference("basal"))
	}
	if p.BasalSchedules != nil {
		p.BasalSchedules.Normalize(normalizer.WithReference("basalSchedules"))
	}
	if p.BloodGlucoseTargets != nil {
		p.BloodGlucoseTargets.Normalize(normalizer.WithReference("bgTarget"), unitsBloodGlucose)
	}
	if p.Bolus != nil {
		p.Bolus.Normalize(normalizer.WithReference("bolus"))
	}
	if p.CarbohydrateRatios != nil {
		p.CarbohydrateRatios.Normalize(normalizer.WithReference("carbRatio"))
	}
	if p.Display != nil {
		p.Display.Normalize(normalizer.WithReference("display"))
	}
	if p.Insulin != nil {
		p.Insulin.Normalize(normalizer.WithReference("insulin"))
	}
	if p.InsulinSensitivities != nil {
		p.InsulinSensitivities.Normalize(normalizer.WithReference("insulinSensitivity"), unitsBloodGlucose)
	}
	if normalizer.Origin() == structure.OriginExternal {
		if p.Manufacturers != nil {
			sort.Strings(*p.Manufacturers)
		}
	}
	if p.Units != nil {
		p.Units.Normalize(normalizer.WithReference("units"))
	}
}
