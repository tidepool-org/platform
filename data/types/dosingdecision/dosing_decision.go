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

	Errors                                      *[]string                                           `json:"errors,omitempty" bson:"errors,omitempty"`
	InsulinOnBoard                              *InsulinOnBoard                                     `json:"insulinOnBoard,omitempty" bson:"insulinOnBoard,omitempty"`
	CarbohydratesOnBoard                        *CarbohydratesOnBoard                               `json:"carbohydratesOnBoard,omitempty" bson:"carbohydratesOnBoard,omitempty"`
	BloodGlucoseTargetSchedule                  *dataTypesSettingsPump.BloodGlucoseTargetStartArray `json:"bloodGlucoseTargetSchedule,omitempty" bson:"bloodGlucoseTargetSchedule,omitempty"`
	BloodGlucoseForecast                        *ForecastArray                                      `json:"bloodGlucoseForecast,omitempty" bson:"bloodGlucoseForecast,omitempty"`
	BloodGlucoseForecastIncludingPendingInsulin *ForecastArray                                      `json:"bloodGlucoseForecastIncludingPendingInsulin,omitempty" bson:"bloodGlucoseForecastIncludingPendingInsulin,omitempty"`
	RecommendedBasal                            *RecommendedBasal                                   `json:"recommendedBasal,omitempty" bson:"recommendedBasal,omitempty"`
	RecommendedBolus                            *RecommendedBolus                                   `json:"recommendedBolus,omitempty" bson:"recommendedBolus,omitempty"`
	Units                                       *Units                                              `json:"units,omitempty" bson:"units,omitempty"`
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

	d.Errors = parser.StringArray("errors")
	d.InsulinOnBoard = ParseInsulinOnBoard(parser.WithReferenceObjectParser("insulinOnBoard"))
	d.CarbohydratesOnBoard = ParseCarbohydratesOnBoard(parser.WithReferenceObjectParser("carbohydratesOnBoard"))
	d.BloodGlucoseTargetSchedule = dataTypesSettingsPump.ParseBloodGlucoseTargetStartArray(parser.WithReferenceArrayParser("bloodGlucoseTargetSchedule"))
	d.BloodGlucoseForecast = ParseForecastArray(parser.WithReferenceArrayParser("bloodGlucoseForecast"))
	d.BloodGlucoseForecastIncludingPendingInsulin = ParseForecastArray(parser.WithReferenceArrayParser("bloodGlucoseForecastIncludingPendingInsulin"))
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
	if d.BloodGlucoseTargetSchedule != nil {
		d.BloodGlucoseTargetSchedule.Validate(validator.WithReference("bloodGlucoseTargetSchedule"), unitsBloodGlucose)
	}
	if d.BloodGlucoseForecast != nil {
		d.BloodGlucoseForecast.Validate(validator.WithReference("bloodGlucoseForecast"))
	}
	if d.BloodGlucoseForecastIncludingPendingInsulin != nil {
		d.BloodGlucoseForecastIncludingPendingInsulin.Validate(validator.WithReference("bloodGlucoseForecastIncludingPendingInsulin"))
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
