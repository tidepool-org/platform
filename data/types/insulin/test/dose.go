package test

import (
	"github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewDose() *insulin.Dose {
	datum := insulin.NewDose()
	datum.Active = pointer.Float64(test.RandomFloat64FromRange(insulin.DoseActiveMinimum, insulin.DoseActiveMaximum))
	datum.Correction = pointer.Float64(test.RandomFloat64FromRange(insulin.DoseCorrectionMinimum, insulin.DoseCorrectionMaximum))
	datum.Food = pointer.Float64(test.RandomFloat64FromRange(insulin.DoseFoodMinimum, insulin.DoseFoodMaximum))
	datum.Total = pointer.Float64(test.RandomFloat64FromRange(insulin.DoseTotalMinimum, insulin.DoseTotalMaximum))
	datum.Units = pointer.String(test.RandomStringFromArray(insulin.DoseUnits()))
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
