package pump

import "github.com/tidepool-org/platform/pointer"

func CloneUnits(datum *Units) *Units {
	if datum == nil {
		return nil
	}
	clone := NewUnits()
	clone.BloodGlucose = pointer.CloneString(datum.BloodGlucose)
	clone.Carbohydrate = pointer.CloneString(datum.Carbohydrate)
	return clone
}
