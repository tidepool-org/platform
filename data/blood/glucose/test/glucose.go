package test

import (
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/test"
)

func RandomUnits() string {
	return test.RandomStringFromArray(dataBloodGlucose.Units())
}
