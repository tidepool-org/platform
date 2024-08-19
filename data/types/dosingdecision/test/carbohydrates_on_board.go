package test

import (
	dataTypesDosingDecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomCarbohydratesOnBoard() *dataTypesDosingDecision.CarbohydratesOnBoard {
	datum := dataTypesDosingDecision.NewCarbohydratesOnBoard()
	datum.Time = pointer.FromTime(test.RandomTime())
	datum.Amount = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDosingDecision.CarbohydratesOnBoardAmountMinimum, dataTypesDosingDecision.CarbohydratesOnBoardAmountMaximum))
	return datum
}

func CloneCarbohydratesOnBoard(datum *dataTypesDosingDecision.CarbohydratesOnBoard) *dataTypesDosingDecision.CarbohydratesOnBoard {
	if datum == nil {
		return nil
	}
	clone := dataTypesDosingDecision.NewCarbohydratesOnBoard()
	clone.Time = pointer.CloneTime(datum.Time)
	clone.Amount = pointer.CloneFloat64(datum.Amount)
	return clone
}

func NewObjectFromCarbohydratesOnBoard(datum *dataTypesDosingDecision.CarbohydratesOnBoard, objectFormat test.ObjectFormat) map[string]interface{} {
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
