package pump

import (
	"sort"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
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

	ActiveScheduleName          *string                          `json:"activeSchedule,omitempty" bson:"activeSchedule,omitempty"` // TODO: Rename to activeScheduleName; move into basal struct
	Basal                       *Basal                           `json:"basal,omitempty" bson:"basal,omitempty"`
	BasalRateSchedule           *BasalRateStartArray             `json:"basalSchedule,omitempty" bson:"basalSchedule,omitempty"`   // TODO: Move into basal struct
	BasalRateSchedules          *BasalRateStartArrayMap          `json:"basalSchedules,omitempty" bson:"basalSchedules,omitempty"` // TODO: Move into basal struct
	BloodGlucoseTargetSchedule  *BloodGlucoseTargetStartArray    `json:"bgTarget,omitempty" bson:"bgTarget,omitempty"`             // TODO: Move into bolus struct
	BloodGlucoseTargetSchedules *BloodGlucoseTargetStartArrayMap `json:"bgTargets,omitempty" bson:"bgTargets,omitempty"`           // TODO: Move into bolus struct
	Bolus                       *Bolus                           `json:"bolus,omitempty" bson:"bolus,omitempty"`
	CarbohydrateRatioSchedule   *CarbohydrateRatioStartArray     `json:"carbRatio,omitempty" bson:"carbRatio,omitempty"`   // TODO: Move into bolus struct
	CarbohydrateRatioSchedules  *CarbohydrateRatioStartArrayMap  `json:"carbRatios,omitempty" bson:"carbRatios,omitempty"` // TODO: Move into bolus struct
	Display                     *Display                         `json:"display,omitempty" bson:"display,omitempty"`
	Insulin                     *Insulin                         `json:"insulin,omitempty" bson:"insulin,omitempty"`
	InsulinSensitivitySchedule  *InsulinSensitivityStartArray    `json:"insulinSensitivity,omitempty" bson:"insulinSensitivity,omitempty"`     // TODO: Move into bolus struct
	InsulinSensitivitySchedules *InsulinSensitivityStartArrayMap `json:"insulinSensitivities,omitempty" bson:"insulinSensitivities,omitempty"` // TODO: Move into bolus struct
	Manufacturers               *[]string                        `json:"manufacturers,omitempty" bson:"manufacturers,omitempty"`
	Model                       *string                          `json:"model,omitempty" bson:"model,omitempty"`
	SerialNumber                *string                          `json:"serialNumber,omitempty" bson:"serialNumber,omitempty"`
	Units                       *Units                           `json:"units,omitempty" bson:"units,omitempty"` // TODO: Move into all appropriate structs
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
	p.BasalRateSchedule = ParseBasalRateStartArray(parser.NewChildArrayParser("basalSchedule"))
	p.BasalRateSchedules = ParseBasalRateStartArrayMap(parser.NewChildObjectParser("basalSchedules"))
	p.BloodGlucoseTargetSchedule = ParseBloodGlucoseTargetStartArray(parser.NewChildArrayParser("bgTarget"))
	p.BloodGlucoseTargetSchedules = ParseBloodGlucoseTargetStartArrayMap(parser.NewChildObjectParser("bgTargets"))
	p.Bolus = ParseBolus(parser.NewChildObjectParser("bolus"))
	p.CarbohydrateRatioSchedule = ParseCarbohydrateRatioStartArray(parser.NewChildArrayParser("carbRatio"))
	p.CarbohydrateRatioSchedules = ParseCarbohydrateRatioStartArrayMap(parser.NewChildObjectParser("carbRatios"))
	p.Display = ParseDisplay(parser.NewChildObjectParser("display"))
	p.Insulin = ParseInsulin(parser.NewChildObjectParser("insulin"))
	p.InsulinSensitivitySchedule = ParseInsulinSensitivityStartArray(parser.NewChildArrayParser("insulinSensitivity"))
	p.InsulinSensitivitySchedules = ParseInsulinSensitivityStartArrayMap(parser.NewChildObjectParser("insulinSensitivities"))
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
	if p.BasalRateSchedule != nil {
		p.BasalRateSchedule.Validate(validator.WithReference("basalSchedule"))
		if p.BasalRateSchedules != nil {
			validator.WithReference("basalSchedules").ReportError(structureValidator.ErrorValueExists())
		}
	} else if p.BasalRateSchedules != nil {
		p.BasalRateSchedules.Validate(validator.WithReference("basalSchedules"))
	} else {
		validator.WithReference("basalSchedule").ReportError(structureValidator.ErrorValueNotExists())
	}
	if p.BloodGlucoseTargetSchedule != nil {
		p.BloodGlucoseTargetSchedule.Validate(validator.WithReference("bgTarget"), unitsBloodGlucose)
		if p.BloodGlucoseTargetSchedules != nil {
			validator.WithReference("bgTargets").ReportError(structureValidator.ErrorValueExists())
		}
	} else if p.BloodGlucoseTargetSchedules != nil {
		p.BloodGlucoseTargetSchedules.Validate(validator.WithReference("bgTargets"), unitsBloodGlucose)
	} else {
		validator.WithReference("bgTarget").ReportError(structureValidator.ErrorValueNotExists())
	}
	if p.Bolus != nil {
		p.Bolus.Validate(validator.WithReference("bolus"))
	}
	if p.CarbohydrateRatioSchedule != nil {
		p.CarbohydrateRatioSchedule.Validate(validator.WithReference("carbRatio"))
		if p.CarbohydrateRatioSchedules != nil {
			validator.WithReference("carbRatios").ReportError(structureValidator.ErrorValueExists())
		}
	} else if p.CarbohydrateRatioSchedules != nil {
		p.CarbohydrateRatioSchedules.Validate(validator.WithReference("carbRatios"))
	} else {
		validator.WithReference("carbRatio").ReportError(structureValidator.ErrorValueNotExists())
	}
	if p.Display != nil {
		p.Display.Validate(validator.WithReference("display"))
	}
	if p.Insulin != nil {
		p.Insulin.Validate(validator.WithReference("insulin"))
	}
	if p.InsulinSensitivitySchedule != nil {
		p.InsulinSensitivitySchedule.Validate(validator.WithReference("insulinSensitivity"), unitsBloodGlucose)
		if p.InsulinSensitivitySchedules != nil {
			validator.WithReference("insulinSensitivities").ReportError(structureValidator.ErrorValueExists())
		}
	} else if p.InsulinSensitivitySchedules != nil {
		p.InsulinSensitivitySchedules.Validate(validator.WithReference("insulinSensitivities"), unitsBloodGlucose)
	} else {
		validator.WithReference("insulinSensitivity").ReportError(structureValidator.ErrorValueNotExists())
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
	if p.BasalRateSchedule != nil {
		p.BasalRateSchedule.Normalize(normalizer.WithReference("basalSchedule"))
	}
	if p.BasalRateSchedules != nil {
		p.BasalRateSchedules.Normalize(normalizer.WithReference("basalSchedules"))
	}
	if p.BloodGlucoseTargetSchedule != nil {
		p.BloodGlucoseTargetSchedule.Normalize(normalizer.WithReference("bgTarget"), unitsBloodGlucose)
	}
	if p.BloodGlucoseTargetSchedules != nil {
		p.BloodGlucoseTargetSchedules.Normalize(normalizer.WithReference("bgTargets"), unitsBloodGlucose)
	}
	if p.Bolus != nil {
		p.Bolus.Normalize(normalizer.WithReference("bolus"))
	}
	if p.CarbohydrateRatioSchedule != nil {
		p.CarbohydrateRatioSchedule.Normalize(normalizer.WithReference("carbRatio"))
	}
	if p.CarbohydrateRatioSchedules != nil {
		p.CarbohydrateRatioSchedules.Normalize(normalizer.WithReference("carbRatios"))
	}
	if p.Display != nil {
		p.Display.Normalize(normalizer.WithReference("display"))
	}
	if p.Insulin != nil {
		p.Insulin.Normalize(normalizer.WithReference("insulin"))
	}
	if p.InsulinSensitivitySchedule != nil {
		p.InsulinSensitivitySchedule.Normalize(normalizer.WithReference("insulinSensitivity"), unitsBloodGlucose)
	}
	if p.InsulinSensitivitySchedules != nil {
		p.InsulinSensitivitySchedules.Normalize(normalizer.WithReference("insulinSensitivities"), unitsBloodGlucose)
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
