package test

import (
	dataTypesDosingDecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	dataTypesSettingsPumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomDosingDecision(unitsBloodGlucose *string) *dataTypesDosingDecision.DosingDecision {
	datum := randomDosingDecision(unitsBloodGlucose)
	datum.Base = *dataTypesTest.RandomBase()
	datum.Type = "dosingDecision"
	return datum
}

func RandomDosingDecisionForParser(unitsBloodGlucose *string) *dataTypesDosingDecision.DosingDecision {
	datum := randomDosingDecision(unitsBloodGlucose)
	datum.Base = *dataTypesTest.RandomBaseForParser()
	datum.Type = "dosingDecision"
	return datum
}

func randomDosingDecision(unitsBloodGlucose *string) *dataTypesDosingDecision.DosingDecision {
	datum := dataTypesDosingDecision.New()
	datum.Reason = pointer.FromString(test.RandomStringFromRange(1, dataTypesDosingDecision.ReasonLengthMaximum))
	datum.OriginalFood = RandomFood()
	datum.Food = RandomFood()
	datum.SelfMonitoredBloodGlucose = RandomBloodGlucose(unitsBloodGlucose)
	datum.CarbohydratesOnBoard = RandomCarbohydratesOnBoard()
	datum.InsulinOnBoard = RandomInsulinOnBoard()
	datum.BloodGlucoseTargetSchedule = dataTypesSettingsPumpTest.RandomBloodGlucoseTargetStartArray(unitsBloodGlucose)
	datum.HistoricalBloodGlucose = RandomBloodGlucoseArray(unitsBloodGlucose)
	datum.ForecastBloodGlucose = RandomForecastBloodGlucoseArray(unitsBloodGlucose)
	datum.RecommendedBasal = RandomRecommendedBasal()
	datum.RecommendedBolus = RandomBolus()
	datum.RequestedBolus = RandomBolus()
	datum.Warnings = RandomIssueArray()
	datum.Errors = RandomIssueArray()
	datum.ScheduleTimeZoneOffset = pointer.FromInt(test.RandomIntFromRange(dataTypesDosingDecision.ScheduleTimeZoneOffsetMinimum, dataTypesDosingDecision.ScheduleTimeZoneOffsetMaximum))
	datum.Units = RandomUnits(unitsBloodGlucose)
	return datum
}

func CloneDosingDecision(datum *dataTypesDosingDecision.DosingDecision) *dataTypesDosingDecision.DosingDecision {
	if datum == nil {
		return nil
	}
	clone := dataTypesDosingDecision.New()
	clone.Base = *dataTypesTest.CloneBase(&datum.Base)
	clone.Reason = pointer.CloneString(datum.Reason)
	clone.OriginalFood = CloneFood(datum.OriginalFood)
	clone.Food = CloneFood(datum.Food)
	clone.SelfMonitoredBloodGlucose = CloneBloodGlucose(datum.SelfMonitoredBloodGlucose)
	clone.CarbohydratesOnBoard = CloneCarbohydratesOnBoard(datum.CarbohydratesOnBoard)
	clone.InsulinOnBoard = CloneInsulinOnBoard(datum.InsulinOnBoard)
	clone.BloodGlucoseTargetSchedule = dataTypesSettingsPumpTest.CloneBloodGlucoseTargetStartArray(datum.BloodGlucoseTargetSchedule)
	clone.HistoricalBloodGlucose = CloneBloodGlucoseArray(datum.HistoricalBloodGlucose)
	clone.ForecastBloodGlucose = CloneForecastBloodGlucoseArray(datum.ForecastBloodGlucose)
	clone.RecommendedBasal = CloneRecommendedBasal(datum.RecommendedBasal)
	clone.RecommendedBolus = CloneBolus(datum.RecommendedBolus)
	clone.RequestedBolus = CloneBolus(datum.RequestedBolus)
	clone.Warnings = CloneIssueArray(datum.Warnings)
	clone.Errors = CloneIssueArray(datum.Errors)
	clone.ScheduleTimeZoneOffset = pointer.CloneInt(datum.ScheduleTimeZoneOffset)
	clone.Units = CloneUnits(datum.Units)
	return clone
}

func NewObjectFromDosingDecision(datum *dataTypesDosingDecision.DosingDecision, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := dataTypesTest.NewObjectFromBase(&datum.Base, objectFormat)
	if datum.Reason != nil {
		object["reason"] = test.NewObjectFromString(*datum.Reason, objectFormat)
	}
	if datum.OriginalFood != nil {
		object["originalFood"] = NewObjectFromFood(datum.OriginalFood, objectFormat)
	}
	if datum.Food != nil {
		object["food"] = NewObjectFromFood(datum.Food, objectFormat)
	}
	if datum.SelfMonitoredBloodGlucose != nil {
		object["smbg"] = NewObjectFromBloodGlucose(datum.SelfMonitoredBloodGlucose, objectFormat)
	}
	if datum.CarbohydratesOnBoard != nil {
		object["carbsOnBoard"] = NewObjectFromCarbohydratesOnBoard(datum.CarbohydratesOnBoard, objectFormat)
	}
	if datum.InsulinOnBoard != nil {
		object["insulinOnBoard"] = NewObjectFromInsulinOnBoard(datum.InsulinOnBoard, objectFormat)
	}
	if datum.BloodGlucoseTargetSchedule != nil {
		object["bgTargetSchedule"] = dataTypesSettingsPumpTest.NewArrayFromBloodGlucoseTargetStartArray(datum.BloodGlucoseTargetSchedule, objectFormat)
	}
	if datum.HistoricalBloodGlucose != nil {
		object["bgHistorical"] = NewArrayFromBloodGlucoseArray(datum.HistoricalBloodGlucose, objectFormat)
	}
	if datum.ForecastBloodGlucose != nil {
		object["bgForecast"] = NewArrayFromForecastBloodGlucoseArray(datum.ForecastBloodGlucose, objectFormat)
	}
	if datum.RecommendedBasal != nil {
		object["recommendedBasal"] = NewObjectFromRecommendedBasal(datum.RecommendedBasal, objectFormat)
	}
	if datum.RecommendedBolus != nil {
		object["recommendedBolus"] = NewObjectFromBolus(datum.RecommendedBolus, objectFormat)
	}
	if datum.RequestedBolus != nil {
		object["requestedBolus"] = NewObjectFromBolus(datum.RequestedBolus, objectFormat)
	}
	if datum.Warnings != nil {
		object["warnings"] = NewArrayFromIssueArray(datum.Warnings, objectFormat)
	}
	if datum.Errors != nil {
		object["errors"] = NewArrayFromIssueArray(datum.Errors, objectFormat)
	}
	if datum.ScheduleTimeZoneOffset != nil {
		object["scheduleTimeZoneOffset"] = test.NewObjectFromInt(*datum.ScheduleTimeZoneOffset, objectFormat)
	}
	if datum.Units != nil {
		object["units"] = NewObjectFromUnits(datum.Units, objectFormat)
	}
	return object
}
