package dosingdecision_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataTypesDosingdecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	dataTypesDosingdecisionTest "github.com/tidepool-org/platform/data/types/dosingdecision/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Bolus", func() {
	It("BolusAmountMaximum is expected", func() {
		Expect(dataTypesDosingdecision.BolusAmountMaximum).To(Equal(100.0))
	})

	It("BolusAmountMinimum is expected", func() {
		Expect(dataTypesDosingdecision.BolusAmountMinimum).To(Equal(0.0))
	})

	It("BolusDurationMaximum is expected", func() {
		Expect(dataTypesDosingdecision.BolusDurationMaximum).To(Equal(86400000))
	})

	It("BolusDurationMinimum is expected", func() {
		Expect(dataTypesDosingdecision.BolusDurationMinimum).To(Equal(0))
	})

	It("BolusExtendedMaximum is expected", func() {
		Expect(dataTypesDosingdecision.BolusExtendedMaximum).To(Equal(100.0))
	})

	It("BolusExtendedMinimum is expected", func() {
		Expect(dataTypesDosingdecision.BolusExtendedMinimum).To(Equal(0.0))
	})

	It("BolusNormalMaximum is expected", func() {
		Expect(dataTypesDosingdecision.BolusNormalMaximum).To(Equal(100.0))
	})

	It("BolusNormalMinimum is expected", func() {
		Expect(dataTypesDosingdecision.BolusNormalMinimum).To(Equal(0.0))
	})

	Context("Bolus", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesDosingdecision.Bolus)) {
				datum := dataTypesDosingdecisionTest.RandomBolus()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesDosingdecisionTest.NewObjectFromBolus(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesDosingdecisionTest.NewObjectFromBolus(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesDosingdecision.Bolus) {},
			),
			Entry("empty",
				func(datum *dataTypesDosingdecision.Bolus) {
					*datum = *dataTypesDosingdecision.NewBolus()
				},
			),
			Entry("all",
				func(datum *dataTypesDosingdecision.Bolus) {
					datum.Duration = pointer.FromInt(test.RandomIntFromRange(dataTypesDosingdecision.BolusDurationMinimum, dataTypesDosingdecision.BolusDurationMaximum))
					datum.Extended = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDosingdecision.BolusExtendedMinimum, dataTypesDosingdecision.BolusExtendedMaximum))
					datum.Normal = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDosingdecision.BolusNormalMinimum, dataTypesDosingdecision.BolusNormalMaximum))
				},
			),
		)

		Context("ParseBolus", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesDosingdecision.ParseBolus(structureParser.NewObject(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesDosingdecisionTest.RandomBolus()
				object := dataTypesDosingdecisionTest.NewObjectFromBolus(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(logTest.NewLogger(), &object)
				Expect(dataTypesDosingdecision.ParseBolus(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewBolus", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesDosingdecision.NewBolus()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Amount).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *dataTypesDosingdecision.Bolus), expectedErrors ...error) {
					expectedDatum := dataTypesDosingdecisionTest.RandomBolus()
					object := dataTypesDosingdecisionTest.NewObjectFromBolus(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesDosingdecision.NewBolus()
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *dataTypesDosingdecision.Bolus) {},
				),
				Entry("multiple errors",
					func(object map[string]any, expectedDatum *dataTypesDosingdecision.Bolus) {
						object["duration"] = true
						object["extended"] = true
						object["normal"] = true
						expectedDatum.Duration = nil
						expectedDatum.Extended = nil
						expectedDatum.Normal = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/duration"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/extended"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/normal"),
				),
			)
		})
		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesDosingdecision.Bolus), expectedErrors ...error) {
					datum := dataTypesDosingdecisionTest.RandomBolus()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesDosingdecision.Bolus) {},
				),
				Entry("duration missing",
					func(datum *dataTypesDosingdecision.Bolus) {
						datum.Duration = nil
						datum.Extended = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDosingdecision.BolusExtendedMinimum, dataTypesDosingdecision.BolusExtendedMaximum))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/duration"),
				),
				Entry("duration; out of range (lower)",
					func(datum *dataTypesDosingdecision.Bolus) {
						datum.Duration = pointer.FromInt(dataTypesDosingdecision.BolusDurationMinimum - 1)
						datum.Extended = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDosingdecision.BolusExtendedMinimum, dataTypesDosingdecision.BolusExtendedMaximum))

					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dataTypesDosingdecision.BolusDurationMinimum-1, dataTypesDosingdecision.BolusDurationMinimum, dataTypesDosingdecision.BolusDurationMaximum), "/duration"),
				),
				Entry("duration; in range (lower)",
					func(datum *dataTypesDosingdecision.Bolus) {
						datum.Duration = pointer.FromInt(dataTypesDosingdecision.BolusDurationMinimum)
						datum.Extended = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDosingdecision.BolusExtendedMinimum, dataTypesDosingdecision.BolusExtendedMaximum))

					},
				),
				Entry("duration; in range (upper)",
					func(datum *dataTypesDosingdecision.Bolus) {
						datum.Duration = pointer.FromInt(dataTypesDosingdecision.BolusDurationMaximum)
						datum.Extended = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDosingdecision.BolusExtendedMinimum, dataTypesDosingdecision.BolusExtendedMaximum))

					},
				),
				Entry("duration; out of range (upper)",
					func(datum *dataTypesDosingdecision.Bolus) {
						datum.Duration = pointer.FromInt(dataTypesDosingdecision.BolusDurationMaximum + 1)
						datum.Extended = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDosingdecision.BolusExtendedMinimum, dataTypesDosingdecision.BolusExtendedMaximum))

					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dataTypesDosingdecision.BolusDurationMaximum+1, dataTypesDosingdecision.BolusDurationMinimum, dataTypesDosingdecision.BolusDurationMaximum), "/duration"),
				),
				Entry("extended; out of range (lower)",
					func(datum *dataTypesDosingdecision.Bolus) {
						datum.Duration = pointer.FromInt(test.RandomIntFromRange(dataTypesDosingdecision.BolusDurationMinimum, dataTypesDosingdecision.BolusDurationMaximum))
						datum.Extended = pointer.FromFloat64(dataTypesDosingdecision.BolusExtendedMinimum - 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dataTypesDosingdecision.BolusExtendedMinimum-0.1, dataTypesDosingdecision.BolusExtendedMinimum, dataTypesDosingdecision.BolusExtendedMaximum), "/extended"),
				),
				Entry("extended; in range (lower)",
					func(datum *dataTypesDosingdecision.Bolus) {
						datum.Duration = pointer.FromInt(test.RandomIntFromRange(dataTypesDosingdecision.BolusDurationMinimum, dataTypesDosingdecision.BolusDurationMaximum))
						datum.Extended = pointer.FromFloat64(dataTypesDosingdecision.BolusExtendedMinimum)
					},
				),
				Entry("extended; in range (upper)",
					func(datum *dataTypesDosingdecision.Bolus) {
						datum.Duration = pointer.FromInt(test.RandomIntFromRange(dataTypesDosingdecision.BolusDurationMinimum, dataTypesDosingdecision.BolusDurationMaximum))
						datum.Extended = pointer.FromFloat64(dataTypesDosingdecision.BolusExtendedMaximum)
					},
				),
				Entry("extended; out of range (upper)",
					func(datum *dataTypesDosingdecision.Bolus) {
						datum.Duration = pointer.FromInt(test.RandomIntFromRange(dataTypesDosingdecision.BolusDurationMinimum, dataTypesDosingdecision.BolusDurationMaximum))
						datum.Extended = pointer.FromFloat64(dataTypesDosingdecision.BolusExtendedMaximum + 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dataTypesDosingdecision.BolusExtendedMaximum+0.1, dataTypesDosingdecision.BolusExtendedMinimum, dataTypesDosingdecision.BolusAmountMaximum), "/extended"),
				),
				Entry("normal; out of range (lower)",
					func(datum *dataTypesDosingdecision.Bolus) {
						datum.Normal = pointer.FromFloat64(dataTypesDosingdecision.BolusNormalMinimum - 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dataTypesDosingdecision.BolusNormalMinimum-0.1, dataTypesDosingdecision.BolusNormalMinimum, dataTypesDosingdecision.BolusNormalMaximum), "/normal"),
				),
				Entry("normal; in range (lower)",
					func(datum *dataTypesDosingdecision.Bolus) {
						datum.Normal = pointer.FromFloat64(dataTypesDosingdecision.BolusNormalMinimum)
					},
				),
				Entry("normal; in range (upper)",
					func(datum *dataTypesDosingdecision.Bolus) {
						datum.Normal = pointer.FromFloat64(dataTypesDosingdecision.BolusNormalMaximum)
					},
				),
				Entry("normal; out of range (upper)",
					func(datum *dataTypesDosingdecision.Bolus) {
						datum.Normal = pointer.FromFloat64(dataTypesDosingdecision.BolusNormalMaximum + 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dataTypesDosingdecision.BolusNormalMaximum+0.1, dataTypesDosingdecision.BolusNormalMinimum, dataTypesDosingdecision.BolusAmountMaximum), "/normal"),
				),
				Entry("extended and normal missing",
					func(datum *dataTypesDosingdecision.Bolus) {
						datum.Duration = nil
						datum.Extended = nil
						datum.Normal = nil
					},
					structureValidator.ErrorValuesNotExistForAny("normal", "extended"),
				),
				Entry("multiple errors",
					func(datum *dataTypesDosingdecision.Bolus) {
						datum.Duration = nil
						datum.Extended = pointer.FromFloat64(dataTypesDosingdecision.BolusExtendedMinimum - 0.1)
						datum.Normal = pointer.FromFloat64(dataTypesDosingdecision.BolusNormalMinimum - 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/duration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dataTypesDosingdecision.BolusExtendedMinimum-0.1, dataTypesDosingdecision.BolusExtendedMinimum, dataTypesDosingdecision.BolusExtendedMaximum), "/extended"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dataTypesDosingdecision.BolusNormalMinimum-0.1, dataTypesDosingdecision.BolusNormalMinimum, dataTypesDosingdecision.BolusNormalMaximum), "/normal"),
				),
			)
		})
	})

	Context("BolusDEPRECATED", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesDosingdecision.Bolus)) {
				datum := dataTypesDosingdecisionTest.RandomBolusDEPRECATED()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesDosingdecisionTest.NewObjectFromBolus(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesDosingdecisionTest.NewObjectFromBolus(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesDosingdecision.Bolus) {},
			),
			Entry("empty",
				func(datum *dataTypesDosingdecision.Bolus) {
					*datum = *dataTypesDosingdecision.NewBolus()
				},
			),
			Entry("all",
				func(datum *dataTypesDosingdecision.Bolus) {
					datum.Amount = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesDosingdecision.BolusAmountMinimum, dataTypesDosingdecision.BolusAmountMaximum))
				},
			),
		)

		Context("ParseBolus", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesDosingdecision.ParseBolus(structureParser.NewObject(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesDosingdecisionTest.RandomBolusDEPRECATED()
				object := dataTypesDosingdecisionTest.NewObjectFromBolus(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(logTest.NewLogger(), &object)
				Expect(dataTypesDosingdecision.ParseBolus(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewBolus", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesDosingdecision.NewBolus()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Amount).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *dataTypesDosingdecision.Bolus), expectedErrors ...error) {
					expectedDatum := dataTypesDosingdecisionTest.RandomBolusDEPRECATED()
					object := dataTypesDosingdecisionTest.NewObjectFromBolus(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesDosingdecision.NewBolus()
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *dataTypesDosingdecision.Bolus) {},
				),
				Entry("multiple errors",
					func(object map[string]any, expectedDatum *dataTypesDosingdecision.Bolus) {
						object["amount"] = true
						expectedDatum.Amount = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/amount"),
				),
			)
		})
		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesDosingdecision.Bolus), expectedErrors ...error) {
					datum := dataTypesDosingdecisionTest.RandomBolusDEPRECATED()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesDosingdecision.Bolus) {},
				),
				Entry("amount missing",
					func(datum *dataTypesDosingdecision.Bolus) { datum.Amount = nil },
					structureValidator.ErrorValuesNotExistForAny("normal", "extended"),
				),
				Entry("amount; out of range (lower)",
					func(datum *dataTypesDosingdecision.Bolus) {
						datum.Amount = pointer.FromFloat64(dataTypesDosingdecision.BolusAmountMinimum - 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dataTypesDosingdecision.BolusAmountMinimum-0.1, dataTypesDosingdecision.BolusAmountMinimum, dataTypesDosingdecision.BolusAmountMaximum), "/amount"),
				),
				Entry("amount; in range (lower)",
					func(datum *dataTypesDosingdecision.Bolus) {
						datum.Amount = pointer.FromFloat64(dataTypesDosingdecision.BolusAmountMinimum)
					},
				),
				Entry("amount; in range (upper)",
					func(datum *dataTypesDosingdecision.Bolus) {
						datum.Amount = pointer.FromFloat64(dataTypesDosingdecision.BolusAmountMaximum)
					},
				),
				Entry("amount; out of range (upper)",
					func(datum *dataTypesDosingdecision.Bolus) {
						datum.Amount = pointer.FromFloat64(dataTypesDosingdecision.BolusAmountMaximum + 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dataTypesDosingdecision.BolusAmountMaximum+0.1, dataTypesDosingdecision.BolusAmountMinimum, dataTypesDosingdecision.BolusAmountMaximum), "/amount"),
				),
			)
		})
	})
})
