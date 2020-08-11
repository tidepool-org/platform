package dosingdecision_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataTypesDosingDecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	dataTypesDosingDecisionTest "github.com/tidepool-org/platform/data/types/dosingdecision/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Forecast", func() {
	Context("Forecast", func() {
		Context("ParseForecast", func() {
			// TODO
		})

		Context("NewForecast", func() {
			It("is successful", func() {
				Expect(dataTypesDosingDecision.NewForecast()).To(Equal(&dataTypesDosingDecision.Forecast{}))
			})
		})

		Context("Forecast", func() {
			Context("Parse", func() {
				// TODO
			})

			Context("Validate", func() {
				DescribeTable("return the expected results when the input",
					func(mutator func(datum *dataTypesDosingDecision.Forecast), expectedErrors ...error) {
						datum := dataTypesDosingDecisionTest.RandomForecast()
						mutator(datum)
						dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
					},
					Entry("succeeds",
						func(datum *dataTypesDosingDecision.Forecast) {},
					),
					Entry("time missing",
						func(datum *dataTypesDosingDecision.Forecast) { datum.Time = nil },
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/time"),
					),
					Entry("value missing",
						func(datum *dataTypesDosingDecision.Forecast) { datum.Value = nil },
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
					),
					Entry("multiple errors",
						func(datum *dataTypesDosingDecision.Forecast) {
							datum.Time = nil
							datum.Value = nil
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/time"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
					),
				)
			})
		})
	})
})
