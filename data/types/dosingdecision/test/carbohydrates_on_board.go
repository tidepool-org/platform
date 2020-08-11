package test

import (
	dataTypesDosingDecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomCarbohydratesOnBoard() *dataTypesDosingDecision.CarbohydratesOnBoard {
	datum := dataTypesDosingDecision.NewCarbohydratesOnBoard()
	datum.StartTime = pointer.FromTime(test.RandomTime())
	datum.EndTime = pointer.FromTime(test.RandomTimeFromRange(*datum.StartTime, test.RandomTimeMaximum()))
	datum.Amount = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDosingDecision.CarbohydratesOnBoardAmountMinimum, dataTypesDosingDecision.CarbohydratesOnBoardAmountMaximum))
	return datum
}

func CloneCarbohydratesOnBoard(datum *dataTypesDosingDecision.CarbohydratesOnBoard) *dataTypesDosingDecision.CarbohydratesOnBoard {
	if datum == nil {
		return nil
	}
	clone := dataTypesDosingDecision.NewCarbohydratesOnBoard()
	clone.StartTime = pointer.CloneTime(datum.StartTime)
	clone.EndTime = pointer.CloneTime(datum.EndTime)
	clone.Amount = pointer.CloneFloat64(datum.Amount)
	return clone
}
