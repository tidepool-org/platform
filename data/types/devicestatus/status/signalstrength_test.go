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

func NewSignalStrength() *status.SignalStrength {
	datum := *status.NewSignalStrength()
	datum.Unit = pointer.FromString("ounces")
	datum.Value = pointer.FromFloat64(10.0)
	return &datum
}

var _ = Describe("SignalStrength", func() {

	Context("SignalStrength", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",

				func(mutator func(datum *status.SignalStrength), expectedErrors ...error) {
					datum := NewSignalStrength()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *status.SignalStrength) {},
				),
				Entry("Unit missing",
					func(datum *status.SignalStrength) { datum.Unit = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/unit"),
				),
				Entry("Value missing",
					func(datum *status.SignalStrength) { datum.Value = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
			)
		})
	})
})
