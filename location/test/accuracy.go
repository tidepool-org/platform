package test

import (
	"github.com/tidepool-org/platform/location"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomAccuracy() *location.Accuracy {
	datum := location.NewAccuracy()
	datum.Units = pointer.FromString(RandomAccuracyUnits())
	datum.Value = pointer.FromFloat64(RandomAccuracyValue(datum.Units))
	return datum
}

func CloneAccuracy(datum *location.Accuracy) *location.Accuracy {
	if datum == nil {
		return nil
	}
	clone := location.NewAccuracy()
	clone.Units = pointer.CloneString(datum.Units)
	clone.Value = pointer.CloneFloat64(datum.Value)
	return clone
}

func NewObjectFromAccuracy(datum *location.Accuracy, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Units != nil {
		object["units"] = test.NewObjectFromString(*datum.Units, objectFormat)
	}
	if datum.Value != nil {
		object["value"] = test.NewObjectFromFloat64(*datum.Value, objectFormat)
	}
	return object
}

func RandomAccuracyUnits() string {
	return test.RandomStringFromArray(location.AccuracyUnits())
}

func RandomAccuracyValue(units *string) float64 {
	return test.RandomFloat64FromRange(location.AccuracyValueRangeForUnits(units))
}
