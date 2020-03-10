package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataTypesSettingsPump "github.com/tidepool-org/platform/data/types/settings/pump"
	dataTypesSettingsPumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func low(a float64, b float64) float64  { return a }
func high(a float64, b float64) float64 { return b }

var _ = Describe("SuspendThreshold", func() {
	Context("ParseSuspendThreshold", func() {
		// TODO
	})

	Context("NewSuspendThreshold", func() {
		It("is successful", func() {
			Expect(dataTypesSettingsPump.NewSuspendThreshold()).To(Equal(&dataTypesSettingsPump.SuspendThreshold{}))
		})
	})

	Context("SuspendThreshold", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesSettingsPump.SuspendThreshold), expectedErrors ...error) {
					datum := dataTypesSettingsPumpTest.RandomSuspendThreshold()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesSettingsPump.SuspendThreshold) {},
				),
				Entry("units missing",
					func(datum *dataTypesSettingsPump.SuspendThreshold) { datum.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *dataTypesSettingsPump.SuspendThreshold) { datum.Units = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dataBloodGlucose.Units()), "/units"),
				),
				Entry("units missing; value missing",
					func(datum *dataTypesSettingsPump.SuspendThreshold) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units missing; value out of range (lower)",
					func(datum *dataTypesSettingsPump.SuspendThreshold) {
						datum.Units = nil
						min, _ := dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L"))
						datum.Value = pointer.FromFloat64(min - 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (lower)",
					func(datum *dataTypesSettingsPump.SuspendThreshold) {
						datum.Units = nil
						min, _ := dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L"))
						datum.Value = pointer.FromFloat64(min)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (upper)",
					func(datum *dataTypesSettingsPump.SuspendThreshold) {
						datum.Units = nil
						_, max := dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L"))
						datum.Value = pointer.FromFloat64(max)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value out of range (upper)",
					func(datum *dataTypesSettingsPump.SuspendThreshold) {
						datum.Units = nil
						_, max := dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L"))
						datum.Value = pointer.FromFloat64(max + 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; value missing",
					func(datum *dataTypesSettingsPump.SuspendThreshold) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dataBloodGlucose.Units()), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units invalid; value out of range (lower)",
					func(datum *dataTypesSettingsPump.SuspendThreshold) {
						datum.Units = pointer.FromString("invalid")
						min, _ := dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L"))
						datum.Value = pointer.FromFloat64(min - 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dataBloodGlucose.Units()), "/units"),
				),
				Entry("units invalid; value in range (lower)",
					func(datum *dataTypesSettingsPump.SuspendThreshold) {
						datum.Units = pointer.FromString("invalid")
						min, _ := dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L"))
						datum.Value = pointer.FromFloat64(min)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dataBloodGlucose.Units()), "/units"),
				),
				Entry("units invalid; value in range (upper)",
					func(datum *dataTypesSettingsPump.SuspendThreshold) {
						datum.Units = pointer.FromString("invalid")
						_, max := dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L"))
						datum.Value = pointer.FromFloat64(max)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dataBloodGlucose.Units()), "/units"),
				),
				Entry("units invalid; value out of range (upper)",
					func(datum *dataTypesSettingsPump.SuspendThreshold) {
						datum.Units = pointer.FromString("invalid")
						_, max := dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L"))
						datum.Value = pointer.FromFloat64(max + 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dataBloodGlucose.Units()), "/units"),
				),
				Entry("units Units; value missing",
					func(datum *dataTypesSettingsPump.SuspendThreshold) {
						datum.Units = pointer.FromString("mmol/L")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units Units; value out of range (lower)",
					func(datum *dataTypesSettingsPump.SuspendThreshold) {
						datum.Units = pointer.FromString("mmol/L")
						min, _ := dataBloodGlucose.ValueRangeForUnits(datum.Units)
						datum.Value = pointer.FromFloat64(min - 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(low(dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L")))-1,
						low(dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L"))),
						high(dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L")))), "/value"),
				),
				Entry("units Units; value in range (lower)",
					func(datum *dataTypesSettingsPump.SuspendThreshold) {
						datum.Units = pointer.FromString("mmol/L")
						min, _ := dataBloodGlucose.ValueRangeForUnits(datum.Units)
						datum.Value = pointer.FromFloat64(min)
					},
				),
				Entry("units Units; value in range (upper)",
					func(datum *dataTypesSettingsPump.SuspendThreshold) {
						datum.Units = pointer.FromString("mmol/L")
						_, max := dataBloodGlucose.ValueRangeForUnits(datum.Units)
						datum.Value = pointer.FromFloat64(max)
					},
				),
				Entry("units Units; value out of range (upper)",
					func(datum *dataTypesSettingsPump.SuspendThreshold) {
						datum.Units = pointer.FromString("mmol/L")
						_, max := dataBloodGlucose.ValueRangeForUnits(datum.Units)
						datum.Value = pointer.FromFloat64(max + 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(high(dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L")))+1,
						low(dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L"))),
						high(dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L")))), "/value"),
				),
				Entry("multiple errors",
					func(datum *dataTypesSettingsPump.SuspendThreshold) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
			)
		})
	})
})
