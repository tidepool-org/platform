package pump

import (
	"sort"

	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataTypes "github.com/tidepool-org/platform/data/types"
	dataTypesInsulin "github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	Type = "pumpSettings"

	FirmwareVersionLengthMaximum  = 100
	HardwareVersionLengthMaximum  = 100
	ManufacturerLengthMaximum     = 100
	ManufacturersLengthMaximum    = 10
	ModelLengthMaximum            = 100
	NameLengthMaximum             = 100
	ScheduleTimeZoneOffsetMaximum = 7 * 24 * 60 // TODO: Make sure same as all time zone offsets
	ScheduleTimeZoneOffsetMinimum = -7 * 24 * 60
	SerialNumberLengthMaximum     = 100
	SoftwareVersionLengthMaximum  = 100
)

// TODO: Consider collapsing *Array objects into ArrayMap objects with "default" name

type Pump struct {
	dataTypes.Base `bson:",inline"`

	ActiveScheduleName                 *string                          `json:"activeSchedule,omitempty" bson:"activeSchedule,omitempty"` // TODO: Rename to activeScheduleName; move into Basal struct
	AutomatedDelivery                  *bool                            `json:"automatedDelivery,omitempty" bson:"automatedDelivery,omitempty"`
	Basal                              *Basal                           `json:"basal,omitempty" bson:"basal,omitempty"`
	BasalRateSchedule                  *BasalRateStartArray             `json:"basalSchedule,omitempty" bson:"basalSchedule,omitempty"`   // TODO: Move into Basal struct; rename schedule
	BasalRateSchedules                 *BasalRateStartArrayMap          `json:"basalSchedules,omitempty" bson:"basalSchedules,omitempty"` // TODO: Move into Basal struct; rename schedules
	BloodGlucoseSafetyLimit            *float64                         `json:"bgSafetyLimit,omitempty" bson:"bgSafetyLimit,omitempty"`
	BloodGlucoseTargetPhysicalActivity *dataBloodGlucose.Target         `json:"bgTargetPhysicalActivity,omitempty" bson:"bgTargetPhysicalActivity,omitempty"`
	BloodGlucoseTargetPreprandial      *dataBloodGlucose.Target         `json:"bgTargetPreprandial,omitempty" bson:"bgTargetPreprandial,omitempty"`
	BloodGlucoseTargetSchedule         *BloodGlucoseTargetStartArray    `json:"bgTarget,omitempty" bson:"bgTarget,omitempty"`   // TODO: Move into BolusCalculator struct; rename bloodGlucoseTarget
	BloodGlucoseTargetSchedules        *BloodGlucoseTargetStartArrayMap `json:"bgTargets,omitempty" bson:"bgTargets,omitempty"` // TODO: Move into BolusCalculator struct; rename bloodGlucoseTargets
	Bolus                              *Bolus                           `json:"bolus,omitempty" bson:"bolus,omitempty"`
	Boluses                            *BolusMap                        `json:"boluses,omitempty" bson:"boluses,omitempty"`
	CarbohydrateRatioSchedule          *CarbohydrateRatioStartArray     `json:"carbRatio,omitempty" bson:"carbRatio,omitempty"`   // TODO: Move into BolusCalculator struct; rename carbohydrateRatio
	CarbohydrateRatioSchedules         *CarbohydrateRatioStartArrayMap  `json:"carbRatios,omitempty" bson:"carbRatios,omitempty"` // TODO: Move into BolusCalculator struct; rename carbohydrateRatios
	Display                            *Display                         `json:"display,omitempty" bson:"display,omitempty"`
	FirmwareVersion                    *string                          `json:"firmwareVersion,omitempty" bson:"firmwareVersion,omitempty"`
	HardwareVersion                    *string                          `json:"hardwareVersion,omitempty" bson:"hardwareVersion,omitempty"`
	InsulinFormulation                 *dataTypesInsulin.Formulation    `json:"insulinFormulation,omitempty" bson:"insulinFormulation,omitempty"`
	InsulinModel                       *InsulinModel                    `json:"insulinModel,omitempty" bson:"insulinModel,omitempty"`
	InsulinSensitivitySchedule         *InsulinSensitivityStartArray    `json:"insulinSensitivity,omitempty" bson:"insulinSensitivity,omitempty"`     // TODO: Move into BolusCalculator struct
	InsulinSensitivitySchedules        *InsulinSensitivityStartArrayMap `json:"insulinSensitivities,omitempty" bson:"insulinSensitivities,omitempty"` // TODO: Move into BolusCalculator struct
	Manufacturers                      *[]string                        `json:"manufacturers,omitempty" bson:"manufacturers,omitempty"`
	Model                              *string                          `json:"model,omitempty" bson:"model,omitempty"`
	Name                               *string                          `json:"name,omitempty" bson:"name,omitempty"`
	OverridePresets                    *OverridePresetMap               `json:"overridePresets,omitempty" bson:"overridePresets,omitempty"`
	ScheduleTimeZoneOffset             *int                             `json:"scheduleTimeZoneOffset,omitempty" bson:"scheduleTimeZoneOffset,omitempty"`
	SerialNumber                       *string                          `json:"serialNumber,omitempty" bson:"serialNumber,omitempty"`
	SleepSchedules                     *SleepScheduleMap                `json:"sleepSchedules,omitempty" bson:"sleepSchedules,omitempty"`
	SoftwareVersion                    *string                          `json:"softwareVersion,omitempty" bson:"softwareVersion,omitempty"`
	Units                              *Units                           `json:"units,omitempty" bson:"units,omitempty"` // TODO: Move into appropriate structs
}

func New() *Pump {
	return &Pump{
		Base: dataTypes.New(Type),
	}
}

func (p *Pump) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(p.Meta())
	}

	p.Base.Parse(parser)

	p.ActiveScheduleName = parser.String("activeSchedule")
	p.AutomatedDelivery = parser.Bool("automatedDelivery")
	p.Basal = ParseBasal(parser.WithReferenceObjectParser("basal"))
	p.BasalRateSchedule = ParseBasalRateStartArray(parser.WithReferenceArrayParser("basalSchedule"))
	p.BasalRateSchedules = ParseBasalRateStartArrayMap(parser.WithReferenceObjectParser("basalSchedules"))
	p.BloodGlucoseSafetyLimit = parser.Float64("bgSafetyLimit")
	p.BloodGlucoseTargetPhysicalActivity = dataBloodGlucose.ParseTarget(parser.WithReferenceObjectParser("bgTargetPhysicalActivity"))
	p.BloodGlucoseTargetPreprandial = dataBloodGlucose.ParseTarget(parser.WithReferenceObjectParser("bgTargetPreprandial"))
	p.BloodGlucoseTargetSchedule = ParseBloodGlucoseTargetStartArray(parser.WithReferenceArrayParser("bgTarget"))
	p.BloodGlucoseTargetSchedules = ParseBloodGlucoseTargetStartArrayMap(parser.WithReferenceObjectParser("bgTargets"))
	p.Bolus = ParseBolus(parser.WithReferenceObjectParser("bolus"))
	p.Boluses = ParseBolusMap(parser.WithReferenceObjectParser("boluses"))
	p.CarbohydrateRatioSchedule = ParseCarbohydrateRatioStartArray(parser.WithReferenceArrayParser("carbRatio"))
	p.CarbohydrateRatioSchedules = ParseCarbohydrateRatioStartArrayMap(parser.WithReferenceObjectParser("carbRatios"))
	p.Display = ParseDisplay(parser.WithReferenceObjectParser("display"))
	p.FirmwareVersion = parser.String("firmwareVersion")
	p.HardwareVersion = parser.String("hardwareVersion")
	p.InsulinFormulation = dataTypesInsulin.ParseFormulation(parser.WithReferenceObjectParser("insulinFormulation"))
	p.InsulinModel = ParseInsulinModel(parser.WithReferenceObjectParser("insulinModel"))
	p.InsulinSensitivitySchedule = ParseInsulinSensitivityStartArray(parser.WithReferenceArrayParser("insulinSensitivity"))
	p.InsulinSensitivitySchedules = ParseInsulinSensitivityStartArrayMap(parser.WithReferenceObjectParser("insulinSensitivities"))
	p.Manufacturers = parser.StringArray("manufacturers")
	p.Model = parser.String("model")
	p.Name = parser.String("name")
	p.OverridePresets = ParseOverridePresetMap(parser.WithReferenceObjectParser("overridePresets"))
	p.ScheduleTimeZoneOffset = parser.Int("scheduleTimeZoneOffset")
	p.SleepSchedules = ParseSleepScheduleMap(parser.WithReferenceObjectParser("sleepSchedules"))
	p.SerialNumber = parser.String("serialNumber")
	p.SoftwareVersion = parser.String("softwareVersion")
	p.Units = ParseUnits(parser.WithReferenceObjectParser("units"))
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
	}
	ValidateBloodGlucoseSafetyLimit(p.BloodGlucoseSafetyLimit, unitsBloodGlucose, "bgSafetyLimit", validator)
	if p.BloodGlucoseTargetPhysicalActivity != nil {
		p.BloodGlucoseTargetPhysicalActivity.Validate(validator.WithReference("bgTargetPhysicalActivity"), unitsBloodGlucose)
	}
	if p.BloodGlucoseTargetPreprandial != nil {
		p.BloodGlucoseTargetPreprandial.Validate(validator.WithReference("bgTargetPreprandial"), unitsBloodGlucose)
	}
	if p.BloodGlucoseTargetSchedule != nil {
		p.BloodGlucoseTargetSchedule.Validate(validator.WithReference("bgTarget"), unitsBloodGlucose)
		if p.BloodGlucoseTargetSchedules != nil {
			validator.WithReference("bgTargets").ReportError(structureValidator.ErrorValueExists())
		}
	} else if p.BloodGlucoseTargetSchedules != nil {
		p.BloodGlucoseTargetSchedules.Validate(validator.WithReference("bgTargets"), unitsBloodGlucose)
	}

	if p.Bolus != nil {
		p.Bolus.Validate(validator.WithReference("bolus"))
		if p.Boluses != nil {
			validator.WithReference("boluses").ReportError(structureValidator.ErrorValueExists())
		}
	} else if p.Boluses != nil {
		p.Boluses.Validate(validator.WithReference("boluses"))
	}

	if p.CarbohydrateRatioSchedule != nil {
		p.CarbohydrateRatioSchedule.Validate(validator.WithReference("carbRatio"))
		if p.CarbohydrateRatioSchedules != nil {
			validator.WithReference("carbRatios").ReportError(structureValidator.ErrorValueExists())
		}
	} else if p.CarbohydrateRatioSchedules != nil {
		p.CarbohydrateRatioSchedules.Validate(validator.WithReference("carbRatios"))
	}
	if p.Display != nil {
		p.Display.Validate(validator.WithReference("display"))
	}
	validator.String("firmwareVersion", p.FirmwareVersion).NotEmpty().LengthLessThanOrEqualTo(FirmwareVersionLengthMaximum)
	validator.String("hardwareVersion", p.HardwareVersion).NotEmpty().LengthLessThanOrEqualTo(HardwareVersionLengthMaximum)
	if p.InsulinFormulation != nil {
		p.InsulinFormulation.Validate(validator.WithReference("insulinFormulation"))
	}
	if p.InsulinModel != nil {
		p.InsulinModel.Validate(validator.WithReference("insulinModel"))
	}
	if p.InsulinSensitivitySchedule != nil {
		p.InsulinSensitivitySchedule.Validate(validator.WithReference("insulinSensitivity"), unitsBloodGlucose)
		if p.InsulinSensitivitySchedules != nil {
			validator.WithReference("insulinSensitivities").ReportError(structureValidator.ErrorValueExists())
		}
	} else if p.InsulinSensitivitySchedules != nil {
		p.InsulinSensitivitySchedules.Validate(validator.WithReference("insulinSensitivities"), unitsBloodGlucose)
	}
	validator.StringArray("manufacturers", p.Manufacturers).NotEmpty().LengthLessThanOrEqualTo(ManufacturersLengthMaximum).Each(func(stringValidator structure.String) {
		stringValidator.Exists().NotEmpty().LengthLessThanOrEqualTo(ManufacturerLengthMaximum)
	}).EachUnique()
	validator.String("model", p.Model).NotEmpty().LengthLessThanOrEqualTo(ModelLengthMaximum)
	validator.String("name", p.Name).NotEmpty().LengthLessThanOrEqualTo(NameLengthMaximum)
	if p.OverridePresets != nil {
		p.OverridePresets.Validate(validator.WithReference("overridePresets"), unitsBloodGlucose)
	}
	if p.SleepSchedules != nil {
		p.SleepSchedules.Validate(validator.WithReference("sleepSchedules"))
	}
	validator.Int("scheduleTimeZoneOffset", p.ScheduleTimeZoneOffset).InRange(ScheduleTimeZoneOffsetMinimum, ScheduleTimeZoneOffsetMaximum)
	validator.String("serialNumber", p.SerialNumber).NotEmpty().LengthLessThanOrEqualTo(SerialNumberLengthMaximum)
	validator.String("softwareVersion", p.SoftwareVersion).NotEmpty().LengthLessThanOrEqualTo(SoftwareVersionLengthMaximum)
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
	if normalizer.Origin() == structure.OriginExternal {
		p.BloodGlucoseSafetyLimit = dataBloodGlucose.NormalizeValueForUnits(p.BloodGlucoseSafetyLimit, unitsBloodGlucose)
	}
	if p.BloodGlucoseTargetPhysicalActivity != nil {
		p.BloodGlucoseTargetPhysicalActivity.Normalize(normalizer.WithReference("bgTargetPhysicalActivity"), unitsBloodGlucose)
	}
	if p.BloodGlucoseTargetPreprandial != nil {
		p.BloodGlucoseTargetPreprandial.Normalize(normalizer.WithReference("bgTargetPreprandial"), unitsBloodGlucose)
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
	if p.Boluses != nil {
		p.Boluses.Normalize(normalizer.WithReference("boluses"))
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
	if p.InsulinFormulation != nil {
		p.InsulinFormulation.Normalize(normalizer.WithReference("insulinFormulation"))
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
	if p.OverridePresets != nil {
		p.OverridePresets.Normalize(normalizer.WithReference("overridePresets"), unitsBloodGlucose)
	}
	if p.SleepSchedules != nil {
		p.SleepSchedules.Normalize(normalizer.WithReference("sleepSchedules"))
	}
	if p.Units != nil {
		p.Units.Normalize(normalizer.WithReference("units"))
	}
}
