package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewSuspendThreshold() *pump.SuspendThreshold {
	datum := pump.NewSuspendThreshold()
	datum.Units = pointer.FromString(test.RandomStringFromArray(dataBloodGlucose.Units()))
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(datum.Units)))

	return datum
}

func CloneSuspendThreshold(datum *pump.SuspendThreshold) *pump.SuspendThreshold {
	if datum == nil {
		return nil
	}
	clone := pump.NewSuspendThreshold()
	clone.Units = pointer.CloneString(datum.Units)
	clone.Value = pointer.CloneFloat64(datum.Value)
	return clone
}

var _ = Describe("SuspendThreshold", func() {

	Context("ParseSuspendThreshold", func() {
		// TODO
	})

	Context("NewSuspendThreshold", func() {
		It("is successful", func() {
			Expect(pump.NewSuspendThreshold()).To(Equal(&pump.SuspendThreshold{}))
		})
	})

	Context("SuspendThreshold", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.SuspendThreshold), expectedErrors ...error) {
					datum := NewSuspendThreshold()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.SuspendThreshold) {},
				),
				Entry("units missing",
					func(datum *pump.SuspendThreshold) { datum.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *pump.SuspendThreshold) { datum.Units = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dataBloodGlucose.Units()), "/units"),
				),
				Entry("units missing; value missing",
					func(datum *pump.SuspendThreshold) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units missing; value out of range (lower)",
					func(datum *pump.SuspendThreshold) {
						datum.Units = nil
						min, _ := dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L"))
						datum.Value = pointer.FromFloat64(min - 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (lower)",
					func(datum *pump.SuspendThreshold) {
						datum.Units = nil
						min, _ := dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L"))
						datum.Value = pointer.FromFloat64(min)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (upper)",
					func(datum *pump.SuspendThreshold) {
						datum.Units = nil
						_, max := dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L"))
						datum.Value = pointer.FromFloat64(max)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value out of range (upper)",
					func(datum *pump.SuspendThreshold) {
						datum.Units = nil
						_, max := dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L"))
						datum.Value = pointer.FromFloat64(max + 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; value missing",
					func(datum *pump.SuspendThreshold) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dataBloodGlucose.Units()), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units invalid; value out of range (lower)",
					func(datum *pump.SuspendThreshold) {
						datum.Units = pointer.FromString("invalid")
						min, _ := dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L"))
						datum.Value = pointer.FromFloat64(min - 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dataBloodGlucose.Units()), "/units"),
				),
				Entry("units invalid; value in range (lower)",
					func(datum *pump.SuspendThreshold) {
						datum.Units = pointer.FromString("invalid")
						min, _ := dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L"))
						datum.Value = pointer.FromFloat64(min)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dataBloodGlucose.Units()), "/units"),
				),
				Entry("units invalid; value in range (upper)",
					func(datum *pump.SuspendThreshold) {
						datum.Units = pointer.FromString("invalid")
						_, max := dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L"))
						datum.Value = pointer.FromFloat64(max)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dataBloodGlucose.Units()), "/units"),
				),
				Entry("units invalid; value out of range (upper)",
					func(datum *pump.SuspendThreshold) {
						datum.Units = pointer.FromString("invalid")
						_, max := dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L"))
						datum.Value = pointer.FromFloat64(max + 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dataBloodGlucose.Units()), "/units"),
				),
				Entry("units Units; value missing",
					func(datum *pump.SuspendThreshold) {
						datum.Units = pointer.FromString("mmol/L")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units Units; value out of range (lower)",
					func(datum *pump.SuspendThreshold) {
						datum.Units = pointer.FromString("mmol/L")
						min, _ := dataBloodGlucose.ValueRangeForUnits(datum.Units)
						datum.Value = pointer.FromFloat64(min - 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(low(dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L")))-1,
						low(dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L"))),
						high(dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L")))), "/value"),
				),
				Entry("units Units; value in range (lower)",
					func(datum *pump.SuspendThreshold) {
						datum.Units = pointer.FromString("mmol/L")
						min, _ := dataBloodGlucose.ValueRangeForUnits(datum.Units)
						datum.Value = pointer.FromFloat64(min)
					},
				),
				Entry("units Units; value in range (upper)",
					func(datum *pump.SuspendThreshold) {
						datum.Units = pointer.FromString("mmol/L")
						_, max := dataBloodGlucose.ValueRangeForUnits(datum.Units)
						datum.Value = pointer.FromFloat64(max)
					},
				),
				Entry("units Units; value out of range (upper)",
					func(datum *pump.SuspendThreshold) {
						datum.Units = pointer.FromString("mmol/L")
						_, max := dataBloodGlucose.ValueRangeForUnits(datum.Units)
						datum.Value = pointer.FromFloat64(max + 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(high(dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L")))+1,
						low(dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L"))),
						high(dataBloodGlucose.ValueRangeForUnits(pointer.FromString("mmol/L")))), "/value"),
				),
				Entry("multiple errors",
					func(datum *pump.SuspendThreshold) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.SuspendThreshold)) {
					for _, origin := range structure.Origins() {
						datum := NewSuspendThreshold()
						mutator(datum)
						expectedDatum := CloneSuspendThreshold(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.SuspendThreshold) {},
				),
				Entry("does not modify the datum; units missing",
					func(datum *pump.SuspendThreshold) { datum.Units = nil },
				),
				Entry("does not modify the datum; value missing",
					func(datum *pump.SuspendThreshold) { datum.Value = nil },
				),
			)
		})
	})
})
