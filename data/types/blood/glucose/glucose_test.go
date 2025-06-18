package glucose_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	dataTypesBloodGlucoseTest "github.com/tidepool-org/platform/data/types/blood/glucose/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Glucose", func() {
	Context("New", func() {
		It("creates a new datum with all values initialized", func() {
			typ := dataTypesTest.NewType()
			datum := glucose.New(typ)
			Expect(datum.Type).To(Equal(typ))
			Expect(datum.Units).To(BeNil())
			Expect(datum.Value).To(BeNil())
		})
	})

	Context("with new datum", func() {
		var typ string
		var datum glucose.Glucose

		BeforeEach(func() {
			typ = dataTypesTest.NewType()
			datum = glucose.New(typ)
		})

		Context("Meta", func() {
			It("returns the meta with delivery type", func() {
				Expect(datum.Meta()).To(Equal(&types.Meta{Type: typ}))
			})
		})
	})

	Context("Validate", func() {
		DescribeTable("validates the datum",
			func(units *string, mutator func(datum *glucose.Glucose, units *string), expectedErrors ...error) {
				datum := dataTypesBloodGlucoseTest.NewGlucose(units)
				mutator(datum, units)
				dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
			},
			Entry("succeeds",
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) {},
			),
			Entry("type missing",
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) { datum.Type = "" },
				errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/type"),
			),
			Entry("type exists",
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) { datum.Type = dataTypesTest.NewType() },
			),
			Entry("units missing; value missing",
				nil,
				func(datum *glucose.Glucose, units *string) {
					datum.Value = nil
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
			),
			Entry("units missing; value out of range (lower)",
				nil,
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(-0.1)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
			),
			Entry("units missing; value in range (lower)",
				nil,
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(0.0)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
			),
			Entry("units missing; value in range (upper)",
				nil,
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(55.0)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
			),
			Entry("units missing; value out of range (upper)",
				nil,
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(1000.1)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
			),
			Entry("units invalid; value missing",
				pointer.FromString("invalid"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = nil
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
			),
			Entry("units invalid; value out of range (lower)",
				pointer.FromString("invalid"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(-0.1)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units"),
			),
			Entry("units invalid; value in range (lower)",
				pointer.FromString("invalid"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(0.0)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units"),
			),
			Entry("units invalid; value in range (upper)",
				pointer.FromString("invalid"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(55.0)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units"),
			),
			Entry("units invalid; value out of range (upper)",
				pointer.FromString("invalid"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(1000.1)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units"),
			),
			Entry("units mmol/L; value missing",
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = nil
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
			),
			Entry("units mmol/L; value out of range (lower)",
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(-0.1)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/value"),
			),
			Entry("units mmol/L; value in range (lower)",
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(0.0)
				},
			),
			Entry("units mmol/L; value in range (upper)",
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(55.0)
				},
			),
			Entry("units mmol/L; value out of range (upper)",
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(55.1)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/value"),
			),
			Entry("units mmol/l; value missing",
				pointer.FromString("mmol/l"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = nil
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
			),
			Entry("units mmol/l; value out of range (lower)",
				pointer.FromString("mmol/l"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(-0.1)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/value"),
			),
			Entry("units mmol/l; value in range (lower)",
				pointer.FromString("mmol/l"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(0.0)
				},
			),
			Entry("units mmol/l; value in range (upper)",
				pointer.FromString("mmol/l"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(55.0)
				},
			),
			Entry("units mmol/l; value out of range (upper)",
				pointer.FromString("mmol/l"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(55.1)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/value"),
			),
			Entry("units mg/dL; value missing",
				pointer.FromString("mg/dL"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = nil
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
			),
			Entry("units mg/dL; value out of range (lower)",
				pointer.FromString("mg/dL"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(-0.1)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value"),
			),
			Entry("units mg/dL; value in range (lower)",
				pointer.FromString("mg/dL"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(0.0)
				},
			),
			Entry("units mg/dL; value in range (upper)",
				pointer.FromString("mg/dL"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(1000.0)
				},
			),
			Entry("units mg/dL; value out of range (upper)",
				pointer.FromString("mg/dL"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(1000.1)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value"),
			),
			Entry("units mg/dl; value missing",
				pointer.FromString("mg/dl"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = nil
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
			),
			Entry("units mg/dl; value out of range (lower)",
				pointer.FromString("mg/dl"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(-0.1)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value"),
			),
			Entry("units mg/dl; value in range (lower)",
				pointer.FromString("mg/dl"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(0.0)
				},
			),
			Entry("units mg/dl; value in range (upper)",
				pointer.FromString("mg/dl"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(1000.0)
				},
			),
			Entry("units mg/dl; value out of range (upper)",
				pointer.FromString("mg/dl"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(1000.1)
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value"),
			),
			Entry("multiple errors",
				nil,
				func(datum *glucose.Glucose, units *string) {
					datum.Type = ""
					datum.Value = nil
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/type"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
			),
		)
	})

	Context("Normalize", func() {
		DescribeTable("normalizes the datum",
			func(units *string, mutator func(datum *glucose.Glucose, units *string), expectator func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string)) {
				for _, origin := range structure.Origins() {
					datum := dataTypesBloodGlucoseTest.NewGlucose(units)
					mutator(datum, units)
					expectedDatum := dataTypesBloodGlucoseTest.CloneGlucose(datum)
					normalizer := dataNormalizer.New(logTest.NewLogger())
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(origin))
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum, units)
					}
					Expect(datum).To(Equal(expectedDatum))
				}
			},
			Entry("does not modify the datum; units missing",
				nil,
				func(datum *glucose.Glucose, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units missing, value missing",
				nil,
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units mmol/L",
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units mmol/L; value missing",
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				nil,
			),
		)

		DescribeTable("normalizes the datum with origin external",
			func(units *string, mutator func(datum *glucose.Glucose, units *string), expectator func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string)) {
				datum := dataTypesBloodGlucoseTest.NewGlucose(units)
				mutator(datum, units)
				expectedDatum := dataTypesBloodGlucoseTest.CloneGlucose(datum)
				normalizer := dataNormalizer.New(logTest.NewLogger())
				Expect(normalizer).ToNot(BeNil())
				datum.Normalize(normalizer.WithOrigin(structure.OriginExternal))
				Expect(normalizer.Error()).To(BeNil())
				Expect(normalizer.Data()).To(BeEmpty())
				if expectator != nil {
					expectator(datum, expectedDatum, units)
				}
				Expect(datum).To(Equal(expectedDatum))
			},
			Entry("does not modify the datum; units missing",
				nil,
				func(datum *glucose.Glucose, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units missing, value missing",
				nil,
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units mmol/L",
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units mmol/L; value missing",
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				nil,
			),
			Entry("modifies the datum; units mmol/l",
				pointer.FromString("mmol/l"),
				func(datum *glucose.Glucose, units *string) {},
				func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
				},
			),
			Entry("modifies the datum; units mmol/l; value missing",
				pointer.FromString("mmol/l"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
				},
			),
			Entry("modifies the datum; units mg/dL",
				pointer.FromString("mg/dL"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(units)))
				},
				func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					dataBloodGlucoseTest.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
				},
			),
			Entry("modifies the datum; units mg/dL; value missing",
				pointer.FromString("mg/dL"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
				},
			),
			Entry("modifies the datum; units mg/dl",
				pointer.FromString("mg/dl"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(units)))
				},
				func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					dataBloodGlucoseTest.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
				},
			),
			Entry("modifies the datum; units mg/dl; value missing",
				pointer.FromString("mg/dl"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
				},
			),
		)

		DescribeTable("normalizes the datum with origin internal/store",
			func(units *string, mutator func(datum *glucose.Glucose, units *string), expectator func(datum *glucose.Glucose, expectedDatum *glucose.Glucose, units *string)) {
				for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
					datum := dataTypesBloodGlucoseTest.NewGlucose(units)
					mutator(datum, units)
					expectedDatum := dataTypesBloodGlucoseTest.CloneGlucose(datum)
					normalizer := dataNormalizer.New(logTest.NewLogger())
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(origin))
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum, units)
					}
					Expect(datum).To(Equal(expectedDatum))
				}
			},
			Entry("does not modify the datum; units missing",
				nil,
				func(datum *glucose.Glucose, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units missing, value missing",
				nil,
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units mmol/L",
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units mmol/L; value missing",
				pointer.FromString("mmol/L"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units mmol/l",
				pointer.FromString("mmol/l"),
				func(datum *glucose.Glucose, units *string) {},
				nil,
			),
			Entry("does not modify the datum; units mmol/l; value missing",
				pointer.FromString("mmol/l"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units mg/dL",
				pointer.FromString("mg/dL"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(units)))
				},
				nil,
			),
			Entry("does not modify the datum; units mg/dL; value missing",
				pointer.FromString("mg/dL"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units mg/dl",
				pointer.FromString("mg/dl"),
				func(datum *glucose.Glucose, units *string) {
					datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(units)))
				},
				nil,
			),
			Entry("does not modify the datum; units mg/dl; value missing",
				pointer.FromString("mg/dl"),
				func(datum *glucose.Glucose, units *string) { datum.Value = nil },
				nil,
			),
		)
	})

	Context("Classify", func() {
		var MmolL = pointer.FromAny(dataBloodGlucose.MmolL)
		var MgdL = pointer.FromAny(dataBloodGlucose.MgdL)

		checkClassification := func(value float64, expectedRange glucose.RangeClassification) {
			GinkgoHelper()
			datum := dataTypesBloodGlucoseTest.NewGlucose(MmolL)
			datum.Value = pointer.FromAny(value)
			got, err := datum.Classify()
			Expect(err).To(Succeed())
			Expect(got).To(Equal(expectedRange))
		}

		It("classifies 2.9 as very low", func() {
			checkClassification(2.9, glucose.RangeVeryLow)
		})

		It("classifies 3.0 as low", func() {
			checkClassification(3.0, glucose.RangeLow)
		})

		It("classifies 3.8 as low", func() {
			checkClassification(3.8, glucose.RangeLow)
		})

		It("classifies 3.9 as on target", func() {
			checkClassification(3.9, glucose.RangeTarget)
		})

		It("classifies 10.0 as on target", func() {
			checkClassification(10.0, glucose.RangeTarget)
		})

		It("classifies 10.1 as high", func() {
			checkClassification(10.1, glucose.RangeHigh)
		})

		It("classifies 13.9 as high", func() {
			checkClassification(13.9, glucose.RangeHigh)
		})

		It("classifies 14.0 as very high", func() {
			checkClassification(14.0, glucose.RangeVeryHigh)
		})

		It("classifies 19.4 as very high", func() {
			checkClassification(19.4, glucose.RangeVeryHigh)
		})

		It("classifies 19.5 as extremely high", func() {
			checkClassification(19.5, glucose.RangeExtremelyHigh)
		})

		When("it's value doesn't require rounding", func() {
			It("classifies 2.95 as very low", func() {
				checkClassification(2.95, glucose.RangeVeryLow)
			})
		})

		When("it's value requires rounding", func() {
			It("classifies 3.85 as low", func() {
				checkClassification(3.85, glucose.RangeLow)
			})

			It("classifies 10.05 as on target", func() {
				checkClassification(10.05, glucose.RangeTarget)
			})
		})

		When("it doesn't recognize the units", func() {
			It("returns an error", func() {
				datum := dataTypesBloodGlucoseTest.NewGlucose(pointer.FromAny("blah"))
				datum.Value = pointer.FromAny(5.0)
				_, err := datum.Classify()
				Expect(err).To(MatchError(ContainSubstring("unhandled units")))
			})
		})

		It("can handle values in mg/dL", func() {
			datum := dataTypesBloodGlucoseTest.NewGlucose(MgdL)
			datum.Value = pointer.FromAny(100.0)
			got, err := datum.Classify()
			Expect(err).To(Succeed())
			Expect(got).To(Equal(glucose.RangeTarget))
		})

		When("it's value is nil", func() {
			It("returns an error", func() {
				datum := dataTypesBloodGlucoseTest.NewGlucose(MmolL)
				datum.Value = nil
				_, err := datum.Classify()
				Expect(err).To(MatchError(ContainSubstring("unhandled value: nil")))
			})
		})

		When("it's units are nil", func() {
			It("returns an error", func() {
				datum := dataTypesBloodGlucoseTest.NewGlucose(nil)
				datum.Value = pointer.FromAny(5.0)
				_, err := datum.Classify()
				Expect(err).To(MatchError(ContainSubstring("unhandled units: nil")))
			})
		})
	})
})
