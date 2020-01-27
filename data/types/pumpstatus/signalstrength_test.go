package pumpstatus_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/pumpstatus"

	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewSignalStrength() *pumpstatus.SignalStrength {
	datum := *pumpstatus.NewSignalStrength()
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

				func(mutator func(datum *pumpstatus.SignalStrength), expectedErrors ...error) {
					datum := NewSignalStrength()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pumpstatus.SignalStrength) {},
				),
				Entry("Unit missing",
					func(datum *pumpstatus.SignalStrength) { datum.Unit = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/unit"),
				),
				Entry("Value missing",
					func(datum *pumpstatus.SignalStrength) { datum.Value = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
			)
		})
	})
})
