package test

import (
	dataTypesDosingDecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomCarbohydratesOnBoard() *dataTypesDosingDecision.CarbohydratesOnBoard {
	startTime := test.RandomTime()
	datum := dataTypesDosingDecision.NewCarbohydratesOnBoard()
	datum.StartTime = pointer.FromString(startTime.Format(dataTypesDosingDecision.TimeFormat))
	datum.EndTime = pointer.FromString(test.RandomTimeFromRange(startTime, test.RandomTimeMaximum()).Format(dataTypesDosingDecision.TimeFormat))
	datum.Amount = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDosingDecision.CarbohydratesOnBoardAmountMinimum, dataTypesDosingDecision.CarbohydratesOnBoardAmountMaximum))
	return datum
}

func CloneCarbohydratesOnBoard(datum *dataTypesDosingDecision.CarbohydratesOnBoard) *dataTypesDosingDecision.CarbohydratesOnBoard {
	if datum == nil {
		return nil
	}
	clone := dataTypesDosingDecision.NewCarbohydratesOnBoard()
	clone.StartTime = pointer.CloneString(datum.StartTime)
	clone.EndTime = pointer.CloneString(datum.EndTime)
	clone.Amount = pointer.CloneFloat64(datum.Amount)
	return clone
}
