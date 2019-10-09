package data_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/test"
)

const (
	InvalidTimeScale = -1
	InvalidStartTime = "invalid"
	InvalidType      = "invalid"
)

func RandomForecast() *data.Forecast {
	startTime := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now().Add(-30*24*time.Hour))
	values := []float64{0.0}

	forecast := data.NewForecast()
	forecast.StartTime = pointer.FromString(startTime.Format(time.RFC3339Nano))
	forecast.TimeScale = pointer.FromInt(test.RandomIntFromRange(data.MinimumTimeScale, data.MaximumTimeScale))
	forecast.Type = pointer.FromString(test.RandomStringFromArray(data.Types()))
	forecast.Unit = pointer.FromString("")
	forecast.Values = &values

	return forecast
}

var _ = Describe("Forecast", func() {
	Context("Forecast", func() {
		DescribeTable("return the expected results when the input",
			func(mutator func(datum *data.Forecast), expectedErrors ...error) {
				datum := RandomForecast()
				mutator(datum)
				dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
			},
			Entry("succeeds",
				func(datum *data.Forecast) {},
			),
			Entry("succeeds",
				func(datum *data.Forecast) {},
			),
			Entry("start time invalid",
				func(datum *data.Forecast) { datum.StartTime = pointer.FromString("invalid") },
				errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid(InvalidStartTime, time.RFC3339Nano), "/startTime"),
			),
			Entry("start time invalid",
				func(datum *data.Forecast) {
					datum.TimeScale = pointer.FromInt(InvalidTimeScale)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(InvalidTimeScale, data.MinimumTimeScale, data.MaximumTimeScale), "/timeScale"),
			),
			Entry("invalid Type",
				func(datum *data.Forecast) {
					datum.Type = pointer.FromString(InvalidType)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf(InvalidType, data.Types()), "/type"),
			),
		)
	})

})
