package dosingdecision

import (
	"github.com/tidepool-org/platform/data"
	dataTypes "github.com/tidepool-org/platform/data/types"
	dataTypesSettingsPump "github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	Type = "dosingDecision"

	ReasonLengthMaximum           = 100
	ScheduleTimeZoneOffsetMaximum = 7 * 24 * 60
	ScheduleTimeZoneOffsetMinimum = -7 * 24 * 60
)

type DosingDecision struct {
	dataTypes.Base `bson:",inline"`

	Reason                     *string                                             `json:"reason,omitempty" bson:"reason,omitempty"`
	OriginalFood               *Food                                               `json:"originalFood,omitempty" bson:"originalFood,omitempty"`
	Food                       *Food                                               `json:"food,omitempty" bson:"food,omitempty"`
	SelfMonitoredBloodGlucose  *BloodGlucose                                       `json:"smbg,omitempty" bson:"smbg,omitempty"`
	CarbohydratesOnBoard       *CarbohydratesOnBoard                               `json:"carbsOnBoard,omitempty" bson:"carbsOnBoard,omitempty"`
	InsulinOnBoard             *InsulinOnBoard                                     `json:"insulinOnBoard,omitempty" bson:"insulinOnBoard,omitempty"`
	BloodGlucoseTargetSchedule *dataTypesSettingsPump.BloodGlucoseTargetStartArray `json:"bgTargetSchedule,omitempty" bson:"bgTargetSchedule,omitempty"`
	HistoricalBloodGlucose     *BloodGlucoseArray                                  `json:"bgHistorical,omitempty" bson:"bgHistorical,omitempty"`
	ForecastBloodGlucose       *ForecastBloodGlucoseArray                          `json:"bgForecast,omitempty" bson:"bgForecast,omitempty"`
	RecommendedBasal           *RecommendedBasal                                   `json:"recommendedBasal,omitempty" bson:"recommendedBasal,omitempty"`
	RecommendedBolus           *Bolus                                              `json:"recommendedBolus,omitempty" bson:"recommendedBolus,omitempty"`
	RequestedBolus             *Bolus                                              `json:"requestedBolus,omitempty" bson:"requestedBolus,omitempty"`
	Warnings                   *IssueArray                                         `json:"warnings,omitempty" bson:"warnings,omitempty"`
	Errors                     *IssueArray                                         `json:"errors,omitempty" bson:"errors,omitempty"`
	ScheduleTimeZoneOffset     *int                                                `json:"scheduleTimeZoneOffset,omitempty" bson:"scheduleTimeZoneOffset,omitempty"`
	Units                      *Units                                              `json:"units,omitempty" bson:"units,omitempty"`
}

func New() *DosingDecision {
	return &DosingDecision{
		Base: dataTypes.New(Type),
	}
}

func (d *DosingDecision) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(d.Meta())
	}

	d.Base.Parse(parser)

	d.Reason = parser.String("reason")
	d.OriginalFood = ParseFood(parser.WithReferenceObjectParser("originalFood"))
	d.Food = ParseFood(parser.WithReferenceObjectParser("food"))
	d.SelfMonitoredBloodGlucose = ParseBloodGlucose(parser.WithReferenceObjectParser("smbg"))
	d.CarbohydratesOnBoard = ParseCarbohydratesOnBoard(parser.WithReferenceObjectParser("carbsOnBoard"))
	d.InsulinOnBoard = ParseInsulinOnBoard(parser.WithReferenceObjectParser("insulinOnBoard"))
	d.BloodGlucoseTargetSchedule = dataTypesSettingsPump.ParseBloodGlucoseTargetStartArray(parser.WithReferenceArrayParser("bgTargetSchedule"))
	d.HistoricalBloodGlucose = ParseBloodGlucoseArray(parser.WithReferenceArrayParser("bgHistorical"))
	d.ForecastBloodGlucose = ParseForecastBloodGlucoseArray(parser.WithReferenceArrayParser("bgForecast"))
	d.RecommendedBasal = ParseRecommendedBasal(parser.WithReferenceObjectParser("recommendedBasal"))
	d.RecommendedBolus = ParseBolus(parser.WithReferenceObjectParser("recommendedBolus"))
	d.RequestedBolus = ParseBolus(parser.WithReferenceObjectParser("requestedBolus"))
	d.Warnings = ParseIssueArray(parser.WithReferenceArrayParser("warnings"))
	d.Errors = ParseIssueArray(parser.WithReferenceArrayParser("errors"))
	d.ScheduleTimeZoneOffset = parser.Int("scheduleTimeZoneOffset")
	d.Units = ParseUnits(parser.WithReferenceObjectParser("units"))
}

func (d *DosingDecision) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(d.Meta())
	}

	d.Base.Validate(validator)

	if d.Type != "" {
		validator.String("type", &d.Type).EqualTo(Type)
	}

	var unitsBloodGlucose *string
	if d.Units != nil {
		unitsBloodGlucose = d.Units.BloodGlucose
	}

	validator.String("reason", d.Reason).Exists().NotEmpty().LengthLessThanOrEqualTo(ReasonLengthMaximum)
	if d.OriginalFood != nil {
		d.OriginalFood.Validate(validator.WithReference("originalFood"))
	}
	if d.Food != nil {
		d.Food.Validate(validator.WithReference("food"))
	}
	if d.SelfMonitoredBloodGlucose != nil {
		d.SelfMonitoredBloodGlucose.Validate(validator.WithReference("smbg"), unitsBloodGlucose)
	}
	if d.CarbohydratesOnBoard != nil {
		d.CarbohydratesOnBoard.Validate(validator.WithReference("carbsOnBoard"))
	}
	if d.InsulinOnBoard != nil {
		d.InsulinOnBoard.Validate(validator.WithReference("insulinOnBoard"))
	}
	if d.BloodGlucoseTargetSchedule != nil {
		d.BloodGlucoseTargetSchedule.Validate(validator.WithReference("bgTargetSchedule"), unitsBloodGlucose)
	}
	if d.HistoricalBloodGlucose != nil {
		d.HistoricalBloodGlucose.Validate(validator.WithReference("bgHistorical"), unitsBloodGlucose)
	}
	if d.ForecastBloodGlucose != nil {
		d.ForecastBloodGlucose.Validate(validator.WithReference("bgForecast"), unitsBloodGlucose)
	}
	if d.RecommendedBasal != nil {
		d.RecommendedBasal.Validate(validator.WithReference("recommendedBasal"))
	}
	if d.RecommendedBolus != nil {
		d.RecommendedBolus.Validate(validator.WithReference("recommendedBolus"))
	}
	if d.RequestedBolus != nil {
		d.RequestedBolus.Validate(validator.WithReference("requestedBolus"))
	}
	if d.Warnings != nil {
		d.Warnings.Validate(validator.WithReference("warnings"))
	}
	if d.Errors != nil {
		d.Errors.Validate(validator.WithReference("errors"))
	}
	validator.Int("scheduleTimeZoneOffset", d.ScheduleTimeZoneOffset).InRange(ScheduleTimeZoneOffsetMinimum, ScheduleTimeZoneOffsetMaximum)
	if unitsValidator := validator.WithReference("units"); d.Units != nil {
		d.Units.Validate(unitsValidator)
	} else {
		unitsValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
}

func (d *DosingDecision) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(d.Meta())
	}

	d.Base.Normalize(normalizer)

	var unitsBloodGlucose *string
	if d.Units != nil {
		unitsBloodGlucose = d.Units.BloodGlucose
	}

	if d.SelfMonitoredBloodGlucose != nil {
		d.SelfMonitoredBloodGlucose.Normalize(normalizer.WithReference("smbg"), unitsBloodGlucose)
	}
	if d.BloodGlucoseTargetSchedule != nil {
		d.BloodGlucoseTargetSchedule.Normalize(normalizer.WithReference("bgTargetSchedule"), unitsBloodGlucose)
	}
	if d.HistoricalBloodGlucose != nil {
		d.HistoricalBloodGlucose.Normalize(normalizer.WithReference("bgHistorical"), unitsBloodGlucose)
	}
	if d.ForecastBloodGlucose != nil {
		d.ForecastBloodGlucose.Normalize(normalizer.WithReference("bgForecast"), unitsBloodGlucose)
	}
	if d.Units != nil {
		d.Units.Normalize(normalizer.WithReference("units"))
	}
}
