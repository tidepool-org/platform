package pumpstatus_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"time"

	dataTypesPumpStatus "github.com/tidepool-org/platform/data/types/pumpstatus"
	dataTypesPumpStatusTest "github.com/tidepool-org/platform/data/types/pumpstatus/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Reservoir", func() {
	It("ReservoirRemainingUnitsMaximum is expected", func() {
		Expect(dataTypesPumpStatus.ReservoirRemainingUnitsMaximum).To(Equal(10000))
	})

	It("ReservoirRemainingUnitsMinimum is expected", func() {
		Expect(dataTypesPumpStatus.ReservoirRemainingUnitsMinimum).To(Equal(0))
	})

	It("ReservoirUnitsUnits is expected", func() {
		Expect(dataTypesPumpStatus.ReservoirUnitsUnits).To(Equal("Units"))
	})

	It("ReservoirUnits returns expected", func() {
		Expect(dataTypesPumpStatus.ReservoirUnits()).To(Equal([]string{"Units"}))
	})

	Context("ParseReservoir", func() {
		// TODO
	})

	Context("NewReservoir", func() {
		It("is successful", func() {
			Expect(dataTypesPumpStatus.NewReservoir()).To(Equal(&dataTypesPumpStatus.Reservoir{}))
		})
	})

	Context("Reservoir", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesPumpStatus.Reservoir), expectedErrors ...error) {
					datum := dataTypesPumpStatusTest.RandomReservoir()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesPumpStatus.Reservoir) {},
				),
				Entry("time invalid",
					func(datum *dataTypesPumpStatus.Reservoir) { datum.Time = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339Nano), "/time"),
				),
				Entry("remaining missing",
					func(datum *dataTypesPumpStatus.Reservoir) { datum.Remaining = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/remaining"),
				),
				Entry("remaining below minimum",
					func(datum *dataTypesPumpStatus.Reservoir) {
						datum.Remaining = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 10000), "/remaining"),
				),
				Entry("remaining above maximum",
					func(datum *dataTypesPumpStatus.Reservoir) { datum.Remaining = pointer.FromFloat64(10000.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10000.1, 0, 10000), "/remaining"),
				),
				Entry("units missing",
					func(datum *dataTypesPumpStatus.Reservoir) { datum.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *dataTypesPumpStatus.Reservoir) { datum.Units = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"Units"}), "/units"),
				),
				Entry("units Units",
					func(datum *dataTypesPumpStatus.Reservoir) { datum.Units = pointer.FromString("Units") },
				),
				Entry("multiple errors",
					func(datum *dataTypesPumpStatus.Reservoir) {
						datum.Time = pointer.FromString("invalid")
						datum.Remaining = nil
						datum.Units = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339Nano), "/time"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/remaining"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
			)
		})
	})
})
