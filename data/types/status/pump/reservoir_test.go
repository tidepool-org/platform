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
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
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

	Context("Reservoir", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesStatusPump.Reservoir)) {
				datum := dataTypesStatusPumpTest.RandomReservoir()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesStatusPumpTest.NewObjectFromReservoir(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesStatusPumpTest.NewObjectFromReservoir(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesStatusPump.Reservoir) {},
			),
			Entry("empty",
				func(datum *dataTypesStatusPump.Reservoir) {
					*datum = *dataTypesStatusPump.NewReservoir()
				},
			),
			Entry("all",
				func(datum *dataTypesStatusPump.Reservoir) {
					datum.Time = pointer.FromTime(test.RandomTime())
					datum.Remaining = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesStatusPump.ReservoirRemainingUnitsMinimum, dataTypesStatusPump.ReservoirRemainingUnitsMaximum))
					datum.Units = pointer.FromString(test.RandomStringFromArray(dataTypesStatusPump.ReservoirUnits()))
				},
			),
		)

		Context("ParseReservoir", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesStatusPump.ParseReservoir(structureParser.NewObject(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesStatusPumpTest.RandomReservoir()
				object := dataTypesStatusPumpTest.NewObjectFromReservoir(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(&object)
				Expect(dataTypesStatusPump.ParseReservoir(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewReservoir", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesStatusPump.NewReservoir()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Time).To(BeNil())
				Expect(datum.Remaining).To(BeNil())
				Expect(datum.Units).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesStatusPump.Reservoir), expectedErrors ...error) {
					expectedDatum := dataTypesStatusPumpTest.RandomReservoir()
					object := dataTypesStatusPumpTest.NewObjectFromReservoir(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesStatusPump.NewReservoir()
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesStatusPump.Reservoir) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesStatusPump.Reservoir) {
						object["time"] = true
						object["remaining"] = true
						object["units"] = true
						expectedDatum.Time = nil
						expectedDatum.Remaining = nil
						expectedDatum.Units = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/time"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/remaining"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/units"),
				),
			)
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
				Entry("remaining out of range (lower)",
					func(datum *dataTypesStatusPump.Reservoir) { datum.Remaining = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 10000), "/remaining"),
				),
				Entry("remaining in range (lower)",
					func(datum *dataTypesStatusPump.Reservoir) { datum.Remaining = pointer.FromFloat64(0) },
				),
				Entry("remaining in range (upper)",
					func(datum *dataTypesStatusPump.Reservoir) { datum.Remaining = pointer.FromFloat64(10000) },
				),
				Entry("remaining out of range (upper)",
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
