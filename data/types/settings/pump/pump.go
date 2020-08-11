package pump

import (
	"sort"

	"github.com/tidepool-org/platform/data/blood/glucose"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	Type = "pumpSettings"

	ManufacturerLengthMaximum     = 100
	ManufacturersLengthMaximum    = 10
	ModelLengthMaximum            = 100
	ScheduleTimeZoneOffsetMaximum = 7 * 24 * 60 // TODO: Make sure same as all time zone offsets
	ScheduleTimeZoneOffsetMinimum = -7 * 24 * 60
	SerialNumberLengthMaximum     = 100
)

// TODO: Consider collapsing *Array objects into ArrayMap objects with "default" name

type Pump struct {
	types.Base `bson:",inline"`

	ActiveScheduleName                 *string                          `json:"activeSchedule,omitempty" bson:"activeSchedule,omitempty"` // TODO: Rename to activeScheduleName; move into Basal struct
	AutomatedDelivery                  *bool                            `json:"automatedDelivery,omitempty" bson:"automatedDelivery,omitempty"`
	Basal                              *Basal                           `json:"basal,omitempty" bson:"basal,omitempty"`
	BasalRateSchedule                  *BasalRateStartArray             `json:"basalSchedule,omitempty" bson:"basalSchedule,omitempty"`   // TODO: Move into Basal struct; rename schedule
	BasalRateSchedules                 *BasalRateStartArrayMap          `json:"basalSchedules,omitempty" bson:"basalSchedules,omitempty"` // TODO: Move into Basal struct; rename schedules
	BloodGlucoseSuspendThreshold       *float64                         `json:"bgSuspendThreshold,omitempty" bson:"bgSuspendThreshold,omitempty"`
	BloodGlucoseTargetPhysicalActivity *glucose.Target                  `json:"bgTargetPhysicalActivity,omitempty" bson:"bgTargetPhysicalActivity,omitempty"`
	BloodGlucoseTargetPreprandial      *glucose.Target                  `json:"bgTargetPreprandial,omitempty" bson:"bgTargetPreprandial,omitempty"`
	BloodGlucoseTargetSchedule         *BloodGlucoseTargetStartArray    `json:"bgTarget,omitempty" bson:"bgTarget,omitempty"`   // TODO: Move into BolusCalculator struct; rename bloodGlucoseTarget
	BloodGlucoseTargetSchedules        *BloodGlucoseTargetStartArrayMap `json:"bgTargets,omitempty" bson:"bgTargets,omitempty"` // TODO: Move into BolusCalculator struct; rename bloodGlucoseTargets
	Bolus                              *Bolus                           `json:"bolus,omitempty" bson:"bolus,omitempty"`
	CarbohydrateRatioSchedule          *CarbohydrateRatioStartArray     `json:"carbRatio,omitempty" bson:"carbRatio,omitempty"`   // TODO: Move into BolusCalculator struct; rename carbohydrateRatio
	CarbohydrateRatioSchedules         *CarbohydrateRatioStartArrayMap  `json:"carbRatios,omitempty" bson:"carbRatios,omitempty"` // TODO: Move into BolusCalculator struct; rename carbohydrateRatios
	Display                            *Display                         `json:"display,omitempty" bson:"display,omitempty"`
	InsulinModel                       *InsulinModel                    `json:"insulinModel,omitempty" bson:"insulinModel,omitempty"`
	InsulinSensitivitySchedule         *InsulinSensitivityStartArray    `json:"insulinSensitivity,omitempty" bson:"insulinSensitivity,omitempty"`     // TODO: Move into BolusCalculator struct
	InsulinSensitivitySchedules        *InsulinSensitivityStartArrayMap `json:"insulinSensitivities,omitempty" bson:"insulinSensitivities,omitempty"` // TODO: Move into BolusCalculator struct
	Manufacturers                      *[]string                        `json:"manufacturers,omitempty" bson:"manufacturers,omitempty"`
	Model                              *string                          `json:"model,omitempty" bson:"model,omitempty"`
	ScheduleTimeZoneOffset             *int                             `json:"scheduleTimeZoneOffset,omitempty" bson:"scheduleTimeZoneOffset,omitempty"`
	SerialNumber                       *string                          `json:"serialNumber,omitempty" bson:"serialNumber,omitempty"`
	Units                              *Units                           `json:"units,omitempty" bson:"units,omitempty"` // TODO: Move into appropriate structs
}

func New() *Pump {
	return &Pump{
		Base: types.New(Type),
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
	p.BloodGlucoseSuspendThreshold = parser.Float64("bgSuspendThreshold")
	p.BloodGlucoseTargetPhysicalActivity = glucose.ParseTarget(parser.WithReferenceObjectParser("bgTargetPhysicalActivity"))
	p.BloodGlucoseTargetPreprandial = glucose.ParseTarget(parser.WithReferenceObjectParser("bgTargetPreprandial"))
	p.BloodGlucoseTargetSchedule = ParseBloodGlucoseTargetStartArray(parser.WithReferenceArrayParser("bgTarget"))
	p.BloodGlucoseTargetSchedules = ParseBloodGlucoseTargetStartArrayMap(parser.WithReferenceObjectParser("bgTargets"))
	p.Bolus = ParseBolus(parser.WithReferenceObjectParser("bolus"))
	p.CarbohydrateRatioSchedule = ParseCarbohydrateRatioStartArray(parser.WithReferenceArrayParser("carbRatio"))
	p.CarbohydrateRatioSchedules = ParseCarbohydrateRatioStartArrayMap(parser.WithReferenceObjectParser("carbRatios"))
	p.Display = ParseDisplay(parser.WithReferenceObjectParser("display"))
	p.InsulinModel = ParseInsulinModel(parser.WithReferenceObjectParser("insulinModel"))
	p.InsulinSensitivitySchedule = ParseInsulinSensitivityStartArray(parser.WithReferenceArrayParser("insulinSensitivity"))
	p.InsulinSensitivitySchedules = ParseInsulinSensitivityStartArrayMap(parser.WithReferenceObjectParser("insulinSensitivities"))
	p.Manufacturers = parser.StringArray("manufacturers")
	p.Model = parser.String("model")
	p.ScheduleTimeZoneOffset = parser.Int("scheduleTimeZoneOffset")
	p.SerialNumber = parser.String("serialNumber")
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
	} else {
		validator.WithReference("basalSchedule").ReportError(structureValidator.ErrorValueNotExists())
	}
	validator.Float64("bgSuspendThreshold", p.BloodGlucoseSuspendThreshold).InRange(dataBloodGlucose.ValueRangeForUnits(unitsBloodGlucose))
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
	} else {
		validator.WithReference("insulinSensitivity").ReportError(structureValidator.ErrorValueNotExists())
	}
	validator.StringArray("manufacturers", p.Manufacturers).NotEmpty().LengthLessThanOrEqualTo(ManufacturersLengthMaximum).Each(func(stringValidator structure.String) {
		stringValidator.Exists().NotEmpty().LengthLessThanOrEqualTo(ManufacturerLengthMaximum)
	}).EachUnique()
	validator.String("model", p.Model).NotEmpty().LengthLessThanOrEqualTo(ModelLengthMaximum)
	validator.Int("scheduleTimeZoneOffset", p.ScheduleTimeZoneOffset).InRange(ScheduleTimeZoneOffsetMinimum, ScheduleTimeZoneOffsetMaximum)
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
	if normalizer.Origin() == structure.OriginExternal {
		p.BloodGlucoseSuspendThreshold = dataBloodGlucose.NormalizeValueForUnits(p.BloodGlucoseSuspendThreshold, unitsBloodGlucose)
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
	if p.CarbohydrateRatioSchedule != nil {
		p.CarbohydrateRatioSchedule.Normalize(normalizer.WithReference("carbRatio"))
	}
	if p.CarbohydrateRatioSchedules != nil {
		p.CarbohydrateRatioSchedules.Normalize(normalizer.WithReference("carbRatios"))
	}
	if p.Display != nil {
		p.Display.Normalize(normalizer.WithReference("display"))
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
