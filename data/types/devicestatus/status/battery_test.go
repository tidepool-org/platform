package status_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/devicestatus/status"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewBattery() *status.Battery {
	datum := *status.NewBattery()
	datum.Unit = pointer.FromString("grams")
	datum.Value = pointer.FromFloat64(5.0)
	return &datum
}

var _ = Describe("Battery", func() {

	Context("Battery", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",

				func(mutator func(datum *status.Battery), expectedErrors ...error) {
					datum := NewBattery()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *status.Battery) {},
				),
				Entry("Unit missing",
					func(datum *status.Battery) { datum.Unit = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/unit"),
				),
				Entry("Value missing",
					func(datum *status.Battery) { datum.Value = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
			)
		})
	})
})
