package test

import (
	dataTypesDosingDecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomInsulinOnBoard() *dataTypesDosingDecision.InsulinOnBoard {
	datum := dataTypesDosingDecision.NewInsulinOnBoard()
	datum.StartTime = pointer.FromTime(test.RandomTime())
	datum.Amount = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDosingDecision.InsulinOnBoardAmountMinimum, dataTypesDosingDecision.InsulinOnBoardAmountMaximum))
	return datum
}

func CloneInsulinOnBoard(datum *dataTypesDosingDecision.InsulinOnBoard) *dataTypesDosingDecision.InsulinOnBoard {
	if datum == nil {
		return nil
	}
	clone := dataTypesDosingDecision.NewInsulinOnBoard()
	clone.StartTime = pointer.CloneTime(datum.StartTime)
	clone.Amount = pointer.CloneFloat64(datum.Amount)
	return clone
}
