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

var _ = Describe("Battery", func() {
	It("BatteryRemainingPercentMaximum is expected", func() {
		Expect(dataTypesStatusPump.BatteryRemainingPercentMaximum).To(Equal(100))
	})

	It("BatteryRemainingPercentMinimum is expected", func() {
		Expect(dataTypesStatusPump.BatteryRemainingPercentMinimum).To(Equal(0))
	})

	It("BatteryUnitsPercent is expected", func() {
		Expect(dataTypesStatusPump.BatteryUnitsPercent).To(Equal("percent"))
	})

	It("BatteryUnits returns expected", func() {
		Expect(dataTypesStatusPump.BatteryUnits()).To(Equal([]string{"percent"}))
	})

	Context("ParseBattery", func() {
		// TODO
	})

	Context("NewBattery", func() {
		It("is successful", func() {
			Expect(dataTypesStatusPump.NewBattery()).To(Equal(&dataTypesStatusPump.Battery{}))
		})
	})

	Context("Battery", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesStatusPump.Battery), expectedErrors ...error) {
					datum := dataTypesStatusPumpTest.RandomBattery()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesStatusPump.Battery) {},
				),
				Entry("remaining missing",
					func(datum *dataTypesStatusPump.Battery) { datum.Remaining = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/remaining"),
				),
				Entry("remaining below minimum",
					func(datum *dataTypesStatusPump.Battery) {
						datum.Remaining = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 100), "/remaining"),
				),
				Entry("remaining above maximum",
					func(datum *dataTypesStatusPump.Battery) {
						datum.Remaining = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0, 100), "/remaining"),
				),
				Entry("units missing",
					func(datum *dataTypesStatusPump.Battery) { datum.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *dataTypesStatusPump.Battery) { datum.Units = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"percent"}), "/units"),
				),
				Entry("units percent",
					func(datum *dataTypesStatusPump.Battery) { datum.Units = pointer.FromString("percent") },
				),
				Entry("multiple errors",
					func(datum *dataTypesStatusPump.Battery) {
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
