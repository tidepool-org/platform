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
	InvalidDate = "invalid"
	InvalidType = "invalid"
)

func RandomForecast() *data.Forecast {

	forecast := data.NewForecast()
	forecast.Date = pointer.FromString(test.RandomTime().Format(time.RFC3339Nano))
	forecast.Type = pointer.FromString(test.RandomStringFromArray(data.Types()))
	forecast.Value = pointer.FromFloat64(test.RandomFloat64FromRange(0, 5))

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
			Entry("start time invalid",
				func(datum *data.Forecast) { datum.Date = pointer.FromString("invalid") },
				errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid(InvalidDate, time.RFC3339Nano), "/date"),
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
