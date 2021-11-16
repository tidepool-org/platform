package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataTypesStatusPump "github.com/tidepool-org/platform/data/types/status/pump"
	dataTypesStatusPumpTest "github.com/tidepool-org/platform/data/types/status/pump/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Reservoir", func() {
	It("ReservoirRemainingUnitsMaximum is expected", func() {
		Expect(dataTypesStatusPump.ReservoirRemainingUnitsMaximum).To(Equal(10000))
	})

	It("ReservoirRemainingUnitsMinimum is expected", func() {
		Expect(dataTypesStatusPump.ReservoirRemainingUnitsMinimum).To(Equal(0))
	})

	It("ReservoirUnitsUnits is expected", func() {
		Expect(dataTypesStatusPump.ReservoirUnitsUnits).To(Equal("Units"))
	})

	It("ReservoirUnits returns expected", func() {
		Expect(dataTypesStatusPump.ReservoirUnits()).To(Equal([]string{"Units"}))
	})

	Context("ParseReservoir", func() {
		// TODO
	})

	Context("NewReservoir", func() {
		It("is successful", func() {
			Expect(dataTypesStatusPump.NewReservoir()).To(Equal(&dataTypesStatusPump.Reservoir{}))
		})
	})

	Context("Reservoir", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesStatusPump.Reservoir), expectedErrors ...error) {
					datum := dataTypesStatusPumpTest.RandomReservoir()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesStatusPump.Reservoir) {},
				),
				Entry("remaining missing",
					func(datum *dataTypesStatusPump.Reservoir) { datum.Remaining = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/remaining"),
				),
				Entry("remaining below minimum",
					func(datum *dataTypesStatusPump.Reservoir) {
						datum.Remaining = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 10000), "/remaining"),
				),
				Entry("remaining above maximum",
					func(datum *dataTypesStatusPump.Reservoir) { datum.Remaining = pointer.FromFloat64(10000.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10000.1, 0, 10000), "/remaining"),
				),
				Entry("units missing",
					func(datum *dataTypesStatusPump.Reservoir) { datum.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *dataTypesStatusPump.Reservoir) { datum.Units = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"Units"}), "/units"),
				),
				Entry("units Units",
					func(datum *dataTypesStatusPump.Reservoir) { datum.Units = pointer.FromString("Units") },
				),
				Entry("multiple errors",
					func(datum *dataTypesStatusPump.Reservoir) {
						datum.Remaining = nil
						datum.Units = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/remaining"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
			)
		})
	})
})
