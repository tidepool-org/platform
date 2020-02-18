package pumpstatus_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/test"

	"github.com/tidepool-org/platform/data/types/pumpstatus"

	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func RandomReservoirRemaining() *pumpstatus.ReservoirRemaining {
	datum := *pumpstatus.NewReservoirRemaining()
	datum.Unit = pointer.FromString("mls")
	datum.Amount = pointer.FromFloat64(test.RandomFloat64FromRange(pumpstatus.MinReservoirAmount, pumpstatus.MaxReservoirAmount))
	return &datum
}

var _ = Describe("ReservoirRemaining", func() {

	Context("ReservoirRemaining", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",

				func(mutator func(datum *pumpstatus.ReservoirRemaining), expectedErrors ...error) {
					datum := RandomReservoirRemaining()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pumpstatus.ReservoirRemaining) {},
				),
				Entry("Unit missing",
					func(datum *pumpstatus.ReservoirRemaining) { datum.Unit = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/unit"),
				),
				Entry("Amount missing",
					func(datum *pumpstatus.ReservoirRemaining) { datum.Amount = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("Value belove Minimum",
					func(datum *pumpstatus.ReservoirRemaining) {
						datum.Amount = pointer.FromFloat64(pumpstatus.MinReservoirAmount - 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(pumpstatus.MinReservoirAmount-1, pumpstatus.MinReservoirAmount, pumpstatus.MaxReservoirAmount), "/amount"),
				),
				Entry("Value above Maximum",
					func(datum *pumpstatus.ReservoirRemaining) {
						datum.Amount = pointer.FromFloat64(pumpstatus.MaxReservoirAmount + 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(pumpstatus.MaxReservoirAmount+1, pumpstatus.MinReservoirAmount, pumpstatus.MaxReservoirAmount), "/amount"),
				),
				Entry("MultipleErrors",
					func(datum *pumpstatus.ReservoirRemaining) {
						datum.Unit = nil
						datum.Amount = pointer.FromFloat64(pumpstatus.MaxReservoirAmount + 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/unit"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(pumpstatus.MaxReservoirAmount+1, pumpstatus.MinReservoirAmount, pumpstatus.MaxReservoirAmount), "/amount"),
				),
			)
		})
	})
})
