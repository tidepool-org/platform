package pump

import (
	"sort"

	"github.com/tidepool-org/platform/data/blood/glucose"

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

	TimeZoneOffsetMaximum = 7 * 24 * 60 // TODO: Make sure same as all time zone offsets
	TimeZoneOffsetMinimum = -7 * 24 * 60

	InsulinModelFiasp      = "fiasp"
	InsulinModelRapidAdult = "rapidAdult"
	InsulinModelRapidChild = "rapidChild"
)

func InsulinModels() []string {
	return []string{
		InsulinModelFiasp,
		InsulinModelRapidAdult,
		InsulinModelRapidChild,
	}
}

// TODO: Consider collapsing *Array objects into ArrayMap objects with "default" name

type Pump struct {
	types.Base `bson:",inline"`

	ActiveScheduleName                        *string                          `json:"activeSchedule,omitempty" bson:"activeSchedule,omitempty"` // TODO: Rename to activeScheduleName; move into Basal struct
	Basal                                     *Basal                           `json:"basal,omitempty" bson:"basal,omitempty"`
	BasalRateSchedule                         *BasalRateStartArray             `json:"basalSchedule,omitempty" bson:"basalSchedule,omitempty"`   // TODO: Move into Basal struct; rename schedule
	BasalRateSchedules                        *BasalRateStartArrayMap          `json:"basalSchedules,omitempty" bson:"basalSchedules,omitempty"` // TODO: Move into Basal struct; rename schedules
	BasalRateSchedulesTimeZoneOffset          *int                             `json:"basalSchedulesTimezoneOffset,omitempty" bson:"basalSchedulesTimezoneOffset,omitempty"`
	BloodGlucoseTargetPreprandial             *glucose.Target                  `json:"bgTargetPreprandial,omitempty" bson:"bgTargetPreprandial,omitempty"`
	BloodGlucoseTargetSchedule                *BloodGlucoseTargetStartArray    `json:"bgTarget,omitempty" bson:"bgTarget,omitempty"`   // TODO: Move into BolusCalculator struct; rename bloodGlucoseTarget
	BloodGlucoseTargetSchedules               *BloodGlucoseTargetStartArrayMap `json:"bgTargets,omitempty" bson:"bgTargets,omitempty"` // TODO: Move into BolusCalculator struct; rename bloodGlucoseTargets
	BloodGlucoseTargetSchedulesTimeZoneOffset *int                             `json:"bgTargetsTimezoneOffset,omitempty" bson:"bgTargetsTimezoneOffset,omitempty"`
	Bolus                                     *Bolus                           `json:"bolus,omitempty" bson:"bolus,omitempty"`
	CarbohydrateRatioSchedule                 *CarbohydrateRatioStartArray     `json:"carbRatio,omitempty" bson:"carbRatio,omitempty"`   // TODO: Move into BolusCalculator struct; rename carbohydrateRatio
	CarbohydrateRatioSchedules                *CarbohydrateRatioStartArrayMap  `json:"carbRatios,omitempty" bson:"carbRatios,omitempty"` // TODO: Move into BolusCalculator struct; rename carbohydrateRatios
	CarbohydrateRatioSchedulesTimeZoneOffset  *int                             `json:"carbRatiosTimezoneOffset,omitempty" bson:"carbRatiosTimezoneOffset,omitempty"`
	Display                                   *Display                         `json:"display,omitempty" bson:"display,omitempty"`
	DosingEnabled                             *bool                            `json:"dosingEnabled,omitempty" bson:"dosingEnabled,omitempty"`
	InsulinModel                              *string                          `json:"insulinModel,omitempty" bson:"insulinModel,omitempty"`
	InsulinSensitivitySchedule                *InsulinSensitivityStartArray    `json:"insulinSensitivity,omitempty" bson:"insulinSensitivity,omitempty"`     // TODO: Move into BolusCalculator struct
	InsulinSensitivitySchedules               *InsulinSensitivityStartArrayMap `json:"insulinSensitivities,omitempty" bson:"insulinSensitivities,omitempty"` // TODO: Move into BolusCalculator struct
	InsulinSensitivitySchedulesTimeZoneOffset *int                             `json:"insulinSensitivitiesTimezoneOffset,omitempty" bson:"insulinSensitivitiesTimezoneOffset,omitempty"`
	Manufacturers                             *[]string                        `json:"manufacturers,omitempty" bson:"manufacturers,omitempty"`
	Model                                     *string                          `json:"model,omitempty" bson:"model,omitempty"`
	SerialNumber                              *string                          `json:"serialNumber,omitempty" bson:"serialNumber,omitempty"`
	SuspendThreshold                          *SuspendThreshold                `json:"suspendThreshold,omitempty" bson:"suspendThreshold,omitempty"`
	Units                                     *Units                           `json:"units,omitempty" bson:"units,omitempty"` // TODO: Move into appropriate structs
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
	p.Basal = ParseBasal(parser.WithReferenceObjectParser("basal"))
	p.BasalRateSchedule = ParseBasalRateStartArray(parser.WithReferenceArrayParser("basalSchedule"))
	p.BasalRateSchedules = ParseBasalRateStartArrayMap(parser.WithReferenceObjectParser("basalSchedules"))
	p.BasalRateSchedulesTimeZoneOffset = parser.Int("basalSchedulesTimezoneOffset")
	p.BloodGlucoseTargetPreprandial = glucose.ParseTarget(parser.WithReferenceObjectParser("bgTargetPreprandial"))
	p.BloodGlucoseTargetSchedule = ParseBloodGlucoseTargetStartArray(parser.WithReferenceArrayParser("bgTarget"))
	p.BloodGlucoseTargetSchedules = ParseBloodGlucoseTargetStartArrayMap(parser.WithReferenceObjectParser("bgTargets"))
	p.BloodGlucoseTargetSchedulesTimeZoneOffset = parser.Int("bgTargetsTimezoneOffset")
	p.Bolus = ParseBolus(parser.WithReferenceObjectParser("bolus"))
	p.CarbohydrateRatioSchedule = ParseCarbohydrateRatioStartArray(parser.WithReferenceArrayParser("carbRatio"))
	p.CarbohydrateRatioSchedules = ParseCarbohydrateRatioStartArrayMap(parser.WithReferenceObjectParser("carbRatios"))
	p.CarbohydrateRatioSchedulesTimeZoneOffset = parser.Int("carbRatiosTimezoneOffset")
	p.Display = ParseDisplay(parser.WithReferenceObjectParser("display"))
	p.DosingEnabled = parser.Bool("dosingEnabled")
	p.InsulinModel = parser.String("insulinModel")
	p.InsulinSensitivitySchedule = ParseInsulinSensitivityStartArray(parser.WithReferenceArrayParser("insulinSensitivity"))
	p.InsulinSensitivitySchedules = ParseInsulinSensitivityStartArrayMap(parser.WithReferenceObjectParser("insulinSensitivities"))
	p.InsulinSensitivitySchedulesTimeZoneOffset = parser.Int("insulinSensitivitiesTimezoneOffset")
	p.Manufacturers = parser.StringArray("manufacturers")
	p.Model = parser.String("model")
	p.SerialNumber = parser.String("serialNumber")
	p.SuspendThreshold = ParseSuspendThreshold(parser.WithReferenceObjectParser("suspendThreshold"))
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
	validator.Int("basalSchedulesTimezoneOffset", p.BasalRateSchedulesTimeZoneOffset).InRange(TimeZoneOffsetMinimum, TimeZoneOffsetMaximum)
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
	validator.Int("bgTargetsTimezoneOffset", p.BloodGlucoseTargetSchedulesTimeZoneOffset).InRange(TimeZoneOffsetMinimum, TimeZoneOffsetMaximum)
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
	validator.Int("carbRatiosTimezoneOffset", p.CarbohydrateRatioSchedulesTimeZoneOffset).InRange(TimeZoneOffsetMinimum, TimeZoneOffsetMaximum)
	if p.Display != nil {
		p.Display.Validate(validator.WithReference("display"))
	}
	validator.String("insulinModel", p.InsulinModel).OneOf(InsulinModels()...)
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
	validator.Int("insulinSensitivitiesTimezoneOffset", p.InsulinSensitivitySchedulesTimeZoneOffset).InRange(TimeZoneOffsetMinimum, TimeZoneOffsetMaximum)
	validator.StringArray("manufacturers", p.Manufacturers).NotEmpty().LengthLessThanOrEqualTo(ManufacturersLengthMaximum).Each(func(stringValidator structure.String) {
		stringValidator.Exists().NotEmpty().LengthLessThanOrEqualTo(ManufacturerLengthMaximum)
	}).EachUnique()
	validator.String("model", p.Model).NotEmpty().LengthLessThanOrEqualTo(ModelLengthMaximum)
	validator.String("serialNumber", p.SerialNumber).NotEmpty().LengthLessThanOrEqualTo(SerialNumberLengthMaximum)
	if p.SuspendThreshold != nil {
		p.SuspendThreshold.Validate(validator.WithReference("suspendThreshold"))
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

	if p.Basal != nil {
		p.Basal.Normalize(normalizer.WithReference("basal"))
	}
	if p.BasalRateSchedule != nil {
		p.BasalRateSchedule.Normalize(normalizer.WithReference("basalSchedule"))
	}
	if p.BasalRateSchedules != nil {
		p.BasalRateSchedules.Normalize(normalizer.WithReference("basalSchedules"))
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
