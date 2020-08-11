package test

import (
	dataTypesDosingDecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	dataTypesSettingsPumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomDosingDecision(unitsBloodGlucose *string) *dataTypesDosingDecision.DosingDecision {
	datum := dataTypesDosingDecision.New()
	datum.Base = *dataTypesTest.NewBase()
	datum.Type = "dosingDecision"
	datum.Errors = pointer.FromStringArray(test.RandomStringArray())
	datum.InsulinOnBoard = RandomInsulinOnBoard()
	datum.CarbohydratesOnBoard = RandomCarbohydratesOnBoard()
	datum.BloodGlucoseTargetSchedule = dataTypesSettingsPumpTest.RandomBloodGlucoseTargetStartArray(unitsBloodGlucose)
	datum.BloodGlucoseForecast = RandomForecastArray()
	datum.RecommendedBasal = RandomRecommendedBasal()
	datum.RecommendedBolus = RandomRecommendedBolus()
	datum.Units = RandomUnits(unitsBloodGlucose)
	return datum
}

func CloneDosingDecision(datum *dataTypesDosingDecision.DosingDecision) *dataTypesDosingDecision.DosingDecision {
	if datum == nil {
		return nil
	}
	clone := dataTypesDosingDecision.New()
	clone.Base = *dataTypesTest.CloneBase(&datum.Base)
	clone.Errors = pointer.CloneStringArray(datum.Errors)
	clone.InsulinOnBoard = CloneInsulinOnBoard(datum.InsulinOnBoard)
	clone.CarbohydratesOnBoard = CloneCarbohydratesOnBoard(datum.CarbohydratesOnBoard)
	clone.BloodGlucoseTargetSchedule = dataTypesSettingsPumpTest.CloneBloodGlucoseTargetStartArray(datum.BloodGlucoseTargetSchedule)
	clone.BloodGlucoseForecast = CloneForecastArray(datum.BloodGlucoseForecast)
	clone.RecommendedBasal = CloneRecommendedBasal(datum.RecommendedBasal)
	clone.RecommendedBolus = CloneRecommendedBolus(datum.RecommendedBolus)
	clone.Units = CloneUnits(datum.Units)
	return clone
}
