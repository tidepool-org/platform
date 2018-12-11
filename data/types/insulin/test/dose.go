package test

import (
	"github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewDose() *insulin.Dose {
	datum := insulin.NewDose()
	datum.Active = pointer.FromFloat64(test.RandomFloat64FromRange(insulin.DoseActiveUnitsMinimum, insulin.DoseActiveUnitsMaximum))
	datum.Correction = pointer.FromFloat64(test.RandomFloat64FromRange(insulin.DoseCorrectionUnitsMinimum, insulin.DoseCorrectionUnitsMaximum))
	datum.Food = pointer.FromFloat64(test.RandomFloat64FromRange(insulin.DoseFoodUnitsMinimum, insulin.DoseFoodUnitsMaximum))
	datum.Total = pointer.FromFloat64(test.RandomFloat64FromRange(insulin.DoseTotalUnitsMinimum, insulin.DoseTotalUnitsMaximum))
	datum.Units = pointer.FromString(test.RandomStringFromArray(insulin.DoseUnits()))
	return datum
}

func CloneDose(datum *insulin.Dose) *insulin.Dose {
	if datum == nil {
		return nil
	}
	clone := insulin.NewDose()
	clone.Active = test.CloneFloat64(datum.Active)
	clone.Correction = test.CloneFloat64(datum.Correction)
	clone.Food = test.CloneFloat64(datum.Food)
	clone.Total = test.CloneFloat64(datum.Total)
	clone.Units = test.CloneString(datum.Units)
	return clone
}
