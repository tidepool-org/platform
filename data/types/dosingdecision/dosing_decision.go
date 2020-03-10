package dosingdecision

import (
	"time"

	"github.com/tidepool-org/platform/data"
	dataTypes "github.com/tidepool-org/platform/data/types"
	dataTypesSettingsPump "github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	Type = "dosingDecision"

	TimeFormat = time.RFC3339Nano
)

type DosingDecision struct {
	dataTypes.Base `bson:",inline"`

	Alerts                          *[]string                                           `json:"alerts,omitempty" bson:"alerts,omitempty"`
	InsulinOnBoard                  *InsulinOnBoard                                     `json:"insulinOnBoard,omitempty" bson:"insulinOnBoard,omitempty"`
	CarbohydratesOnBoard            *CarbohydratesOnBoard                               `json:"carbohydratesOnBoard,omitempty" bson:"carbohydratesOnBoard,omitempty"`
	BloodGlucoseTargetRangeSchedule *dataTypesSettingsPump.BloodGlucoseTargetStartArray `json:"bloodGlucoseTargetRangeSchedule,omitempty" bson:"bloodGlucoseTargetRangeSchedule,omitempty"`
	BloodGlucoseForecast            *ForecastArray                                      `json:"bloodGlucoseForecast,omitempty" bson:"bloodGlucoseForecast,omitempty"`
	RecommendedBasal                *RecommendedBasal                                   `json:"recommendedBasal,omitempty" bson:"recommendedBasal,omitempty"`
	RecommendedBolus                *RecommendedBolus                                   `json:"recommendedBolus,omitempty" bson:"recommendedBolus,omitempty"`
	Units                           *Units                                              `json:"units,omitempty" bson:"units,omitempty"`
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

	d.Alerts = parser.StringArray("alerts")
	d.InsulinOnBoard = ParseInsulinOnBoard(parser.WithReferenceObjectParser("insulinOnBoard"))
	d.CarbohydratesOnBoard = ParseCarbohydratesOnBoard(parser.WithReferenceObjectParser("carbohydratesOnBoard"))
	d.BloodGlucoseTargetRangeSchedule = dataTypesSettingsPump.ParseBloodGlucoseTargetStartArray(parser.WithReferenceArrayParser("bloodGlucoseTargetRangeSchedule"))
	d.BloodGlucoseForecast = ParseForecastArray(parser.WithReferenceArrayParser("bloodGlucoseForecast"))
	d.RecommendedBasal = ParseRecommendedBasal(parser.WithReferenceObjectParser("recommendedBasal"))
	d.RecommendedBolus = ParseRecommendedBolus(parser.WithReferenceObjectParser("recommendedBolus"))
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

	if d.InsulinOnBoard != nil {
		d.InsulinOnBoard.Validate(validator.WithReference("insulinOnBoard"))
	}
	if d.CarbohydratesOnBoard != nil {
		d.CarbohydratesOnBoard.Validate(validator.WithReference("carbohydratesOnBoard"))
	}
	if d.BloodGlucoseTargetRangeSchedule != nil {
		d.BloodGlucoseTargetRangeSchedule.Validate(validator.WithReference("bloodGlucoseTargetRangeSchedule"), unitsBloodGlucose)
	}
	if d.BloodGlucoseForecast != nil {
		d.BloodGlucoseForecast.Validate(validator.WithReference("bloodGlucoseForecast"))
	}
	if d.RecommendedBasal != nil {
		d.RecommendedBasal.Validate(validator.WithReference("recommendedBasal"))
	}
	if d.RecommendedBolus != nil {
		d.RecommendedBolus.Validate(validator.WithReference("recommendedBolus"))
	}
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
}
