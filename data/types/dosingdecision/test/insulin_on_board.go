package test

import (
	dataTypesDosingDecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomInsulinOnBoard() *dataTypesDosingDecision.InsulinOnBoard {
	datum := dataTypesDosingDecision.NewInsulinOnBoard()
	datum.Time = pointer.FromTime(test.RandomTime())
	datum.Amount = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDosingDecision.InsulinOnBoardAmountMinimum, dataTypesDosingDecision.InsulinOnBoardAmountMaximum))
	return datum
}

func CloneInsulinOnBoard(datum *dataTypesDosingDecision.InsulinOnBoard) *dataTypesDosingDecision.InsulinOnBoard {
	if datum == nil {
		return nil
	}
	clone := dataTypesDosingDecision.NewInsulinOnBoard()
	clone.Time = pointer.CloneTime(datum.Time)
	clone.Amount = pointer.CloneFloat64(datum.Amount)
	return clone
}

func NewObjectFromInsulinOnBoard(datum *dataTypesDosingDecision.InsulinOnBoard, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Time != nil {
		object["time"] = test.NewObjectFromTime(*datum.Time, objectFormat)
	}
	if datum.Amount != nil {
		object["amount"] = test.NewObjectFromFloat64(*datum.Amount, objectFormat)
	}
	return object
}
