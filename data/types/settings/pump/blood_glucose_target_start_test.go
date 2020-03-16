package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	pumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("BloodGlucoseTargetStart", func() {
	It("BloodGlucoseTargetStartStartMaximum is expected", func() {
		Expect(pump.BloodGlucoseTargetStartStartMaximum).To(Equal(86400000))
	})

	It("BloodGlucoseTargetStartStartMinimum is expected", func() {
		Expect(pump.BloodGlucoseTargetStartStartMinimum).To(Equal(0))
	})

	Context("ParseBloodGlucoseTargetStart", func() {
		// TODO
	})

	Context("NewBloodGlucoseTargetStart", func() {
		It("is successful", func() {
			Expect(pump.NewBloodGlucoseTargetStart()).To(Equal(&pump.BloodGlucoseTargetStart{}))
		})
	})

	Context("BloodGlucoseTargetStart", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *pump.BloodGlucoseTargetStart, units *string), expectedErrors ...error) {
					datum := pumpTest.NewBloodGlucoseTargetStart(units, pump.BloodGlucoseTargetStartStartMinimum)
					mutator(datum, units)
					dataTypesTest.ValidateWithExpectedOrigins(pumpTest.NewValidatableWithUnitsAndStartMinimumAdapter(datum, units, pointer.FromInt(pump.BloodGlucoseTargetStartStartMinimum)), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStart, units *string) {},
				),
				Entry("target missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStart, units *string) {
						datum.Target = *dataBloodGlucose.NewTarget()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/target"),
				),
				Entry("start missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStart, units *string) { datum.Start = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStart, units *string) {
						datum.Target = *dataBloodGlucose.NewTarget()
						datum.Start = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/target"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
				),
			)

			DescribeTable("validates the datum with minimum start",
				func(units *string, mutator func(datum *pump.BloodGlucoseTargetStart, units *string), expectedErrors ...error) {
					datum := pumpTest.NewBloodGlucoseTargetStart(units, pump.BloodGlucoseTargetStartStartMinimum)
					mutator(datum, units)
					dataTypesTest.ValidateWithExpectedOrigins(pumpTest.NewValidatableWithUnitsAndStartMinimumAdapter(datum, units, pointer.FromInt(pump.BloodGlucoseTargetStartStartMinimum)), structure.Origins(), expectedErrors...)
				},
				Entry("start out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStart, units *string) { datum.Start = pointer.FromInt(-1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo(-1, 0), "/start"),
				),
				Entry("start in range",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStart, units *string) { datum.Start = pointer.FromInt(0) },
				),
				Entry("start out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStart, units *string) { datum.Start = pointer.FromInt(1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo(1, 0), "/start"),
				),
			)

			DescribeTable("validates the datum with non-minimum start",
				func(units *string, mutator func(datum *pump.BloodGlucoseTargetStart, units *string), expectedErrors ...error) {
					datum := pumpTest.NewBloodGlucoseTargetStart(units, pump.BloodGlucoseTargetStartStartMinimum+1)
					mutator(datum, units)
					dataTypesTest.ValidateWithExpectedOrigins(pumpTest.NewValidatableWithUnitsAndStartMinimumAdapter(datum, units, pointer.FromInt(pump.BloodGlucoseTargetStartStartMinimum+1)), structure.Origins(), expectedErrors...)
				},
				Entry("start out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStart, units *string) { datum.Start = pointer.FromInt(0) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0, 1, 86400000), "/start"),
				),
				Entry("start in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStart, units *string) { datum.Start = pointer.FromInt(1) },
				),
				Entry("start in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStart, units *string) { datum.Start = pointer.FromInt(86400000) },
				),
				Entry("start out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStart, units *string) { datum.Start = pointer.FromInt(86400001) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(86400001, 1, 86400000), "/start"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *pump.BloodGlucoseTargetStart, units *string), expectator func(datum *pump.BloodGlucoseTargetStart, expectedDatum *pump.BloodGlucoseTargetStart, units *string)) {
					for _, origin := range structure.Origins() {
						datum := pumpTest.NewBloodGlucoseTargetStart(units, pump.BloodGlucoseTargetStartStartMinimum+1)
						mutator(datum, units)
						expectedDatum := pumpTest.CloneBloodGlucoseTargetStart(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin), units)
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum, units)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStart, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *pump.BloodGlucoseTargetStart, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *pump.BloodGlucoseTargetStart, units *string) {},
					nil,
				),
				Entry("does not modify the datum; start missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStart, units *string) { datum.Start = nil },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *pump.BloodGlucoseTargetStart, units *string), expectator func(datum *pump.BloodGlucoseTargetStart, expectedDatum *pump.BloodGlucoseTargetStart, units *string)) {
					datum := pumpTest.NewBloodGlucoseTargetStart(units, pump.BloodGlucoseTargetStartStartMinimum+1)
					mutator(datum, units)
					expectedDatum := pumpTest.CloneBloodGlucoseTargetStart(datum)
					normalizer := dataNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(structure.OriginExternal), units)
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum, units)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStart, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *pump.BloodGlucoseTargetStart, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *pump.BloodGlucoseTargetStart, units *string) {},
					func(datum *pump.BloodGlucoseTargetStart, expectedDatum *pump.BloodGlucoseTargetStart, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedTarget(&datum.Target, &expectedDatum.Target, units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *pump.BloodGlucoseTargetStart, units *string) {},
					func(datum *pump.BloodGlucoseTargetStart, expectedDatum *pump.BloodGlucoseTargetStart, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedTarget(&datum.Target, &expectedDatum.Target, units)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *pump.BloodGlucoseTargetStart, units *string), expectator func(datum *pump.BloodGlucoseTargetStart, expectedDatum *pump.BloodGlucoseTargetStart, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := pumpTest.NewBloodGlucoseTargetStart(units, pump.BloodGlucoseTargetStartStartMinimum+1)
						mutator(datum, units)
						expectedDatum := pumpTest.CloneBloodGlucoseTargetStart(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin), units)
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum, units)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStart, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *pump.BloodGlucoseTargetStart, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *pump.BloodGlucoseTargetStart, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *pump.BloodGlucoseTargetStart, units *string) {},
					nil,
				),
			)
		})
	})

	Context("ParseBloodGlucoseTargetStartArray", func() {
		// TODO
	})

	Context("NewBloodGlucoseTargetStartArray", func() {
		It("is successful", func() {
			Expect(pump.NewBloodGlucoseTargetStartArray()).To(Equal(&pump.BloodGlucoseTargetStartArray{}))
		})
	})

	Context("BloodGlucoseTargetStartArray", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *pump.BloodGlucoseTargetStartArray, units *string), expectedErrors ...error) {
					datum := pump.NewBloodGlucoseTargetStartArray()
					mutator(datum, units)
					dataTypesTest.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringAdapter(datum, units), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStartArray, units *string) {},
				),
				Entry("empty",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStartArray, units *string) {
						*datum = *pump.NewBloodGlucoseTargetStartArray()
					},
				),
				Entry("nil",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStartArray, units *string) { *datum = append(*datum, nil) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
				),
				Entry("single invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStartArray, units *string) {
						invalid := pumpTest.NewBloodGlucoseTargetStart(units, pump.BloodGlucoseTargetStartStartMinimum)
						invalid.Target = *dataBloodGlucose.NewTarget()
						*datum = append(*datum, invalid)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0/target"),
				),
				Entry("single valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStartArray, units *string) {
						*datum = append(*datum, pumpTest.NewBloodGlucoseTargetStart(units, pump.BloodGlucoseTargetStartStartMinimum))
					},
				),
				Entry("multiple invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStartArray, units *string) {
						*datum = append(*datum, pumpTest.NewBloodGlucoseTargetStart(units, pump.BloodGlucoseTargetStartStartMinimum))
						invalid := pumpTest.NewBloodGlucoseTargetStart(units, *datum.Last().Start+1)
						invalid.Target = *dataBloodGlucose.NewTarget()
						*datum = append(*datum, invalid)
						*datum = append(*datum, pumpTest.NewBloodGlucoseTargetStart(units, *datum.Last().Start+1))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/target"),
				),
				Entry("multiple valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStartArray, units *string) {
						*datum = append(*datum, pumpTest.NewBloodGlucoseTargetStart(units, pump.BloodGlucoseTargetStartStartMinimum))
						*datum = append(*datum, pumpTest.NewBloodGlucoseTargetStart(units, *datum.Last().Start+1))
						*datum = append(*datum, pumpTest.NewBloodGlucoseTargetStart(units, *datum.Last().Start+1))
					},
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStartArray, units *string) {
						invalid := pumpTest.NewBloodGlucoseTargetStart(units, pump.BloodGlucoseTargetStartStartMinimum)
						invalid.Target = *dataBloodGlucose.NewTarget()
						*datum = append(*datum, nil, invalid)
						*datum = append(*datum, nil, pumpTest.NewBloodGlucoseTargetStart(units, *datum.Last().Start+1))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/target"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/2"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *pump.BloodGlucoseTargetStartArray, units *string), expectator func(datum *pump.BloodGlucoseTargetStartArray, expectedDatum *pump.BloodGlucoseTargetStartArray, units *string)) {
					for _, origin := range structure.Origins() {
						datum := pumpTest.NewBloodGlucoseTargetStartArray(units)
						mutator(datum, units)
						expectedDatum := pumpTest.CloneBloodGlucoseTargetStartArray(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin), units)
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum, units)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStartArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *pump.BloodGlucoseTargetStartArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *pump.BloodGlucoseTargetStartArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; start missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStartArray, units *string) { (*datum)[0].Start = nil },
					nil,
				),
			)

			DescribeTable("normalizes the datum with structure external",
				func(units *string, mutator func(datum *pump.BloodGlucoseTargetStartArray, units *string), expectator func(datum *pump.BloodGlucoseTargetStartArray, expectedDatum *pump.BloodGlucoseTargetStartArray, units *string)) {
					datum := pumpTest.NewBloodGlucoseTargetStartArray(units)
					mutator(datum, units)
					expectedDatum := pumpTest.CloneBloodGlucoseTargetStartArray(datum)
					normalizer := dataNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(structure.OriginExternal), units)
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum, units)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStartArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *pump.BloodGlucoseTargetStartArray, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *pump.BloodGlucoseTargetStartArray, units *string) {},
					func(datum *pump.BloodGlucoseTargetStartArray, expectedDatum *pump.BloodGlucoseTargetStartArray, units *string) {
						for index := range *datum {
							dataBloodGlucoseTest.ExpectNormalizedTarget(&(*datum)[index].Target, &(*expectedDatum)[index].Target, units)
						}
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *pump.BloodGlucoseTargetStartArray, units *string) {},
					func(datum *pump.BloodGlucoseTargetStartArray, expectedDatum *pump.BloodGlucoseTargetStartArray, units *string) {
						for index := range *datum {
							dataBloodGlucoseTest.ExpectNormalizedTarget(&(*datum)[index].Target, &(*expectedDatum)[index].Target, units)
						}
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *pump.BloodGlucoseTargetStartArray, units *string), expectator func(datum *pump.BloodGlucoseTargetStartArray, expectedDatum *pump.BloodGlucoseTargetStartArray, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := pumpTest.NewBloodGlucoseTargetStartArray(units)
						mutator(datum, units)
						expectedDatum := pumpTest.CloneBloodGlucoseTargetStartArray(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin), units)
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum, units)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStartArray, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *pump.BloodGlucoseTargetStartArray, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *pump.BloodGlucoseTargetStartArray, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *pump.BloodGlucoseTargetStartArray, units *string) {},
					nil,
				),
			)
		})

		Context("First", func() {
			var datum *pump.BloodGlucoseTargetStartArray

			BeforeEach(func() {
				datum = pump.NewBloodGlucoseTargetStartArray()
			})

			It("returns nil if array is empty", func() {
				Expect(datum.First()).To(BeNil())
			})

			It("returns the first element if the array has one element", func() {
				*datum = append(*datum, pumpTest.NewBloodGlucoseTargetStart(pointer.FromString("mmol/L"), pump.BloodGlucoseTargetStartStartMinimum))
				Expect(datum.First()).To(Equal((*datum)[0]))
			})

			It("returns the first element if the array has multiple elements", func() {
				*datum = append(*datum, pumpTest.NewBloodGlucoseTargetStart(pointer.FromString("mmol/L"), pump.BloodGlucoseTargetStartStartMinimum))
				*datum = append(*datum, pumpTest.NewBloodGlucoseTargetStart(pointer.FromString("mmol/L"), *datum.Last().Start+1))
				*datum = append(*datum, pumpTest.NewBloodGlucoseTargetStart(pointer.FromString("mmol/L"), *datum.Last().Start+1))
				Expect(datum.First()).To(Equal((*datum)[0]))
			})
		})

		Context("Last", func() {
			var datum *pump.BloodGlucoseTargetStartArray

			BeforeEach(func() {
				datum = pump.NewBloodGlucoseTargetStartArray()
			})

			It("returns nil if array is empty", func() {
				Expect(datum.Last()).To(BeNil())
			})

			It("returns the last element if the array has one element", func() {
				*datum = append(*datum, pumpTest.NewBloodGlucoseTargetStart(pointer.FromString("mmol/L"), pump.BloodGlucoseTargetStartStartMinimum))
				Expect(datum.Last()).To(Equal((*datum)[0]))
			})

			It("returns the last element if the array has multiple elements", func() {
				*datum = append(*datum, pumpTest.NewBloodGlucoseTargetStart(pointer.FromString("mmol/L"), pump.BloodGlucoseTargetStartStartMinimum))
				*datum = append(*datum, pumpTest.NewBloodGlucoseTargetStart(pointer.FromString("mmol/L"), *datum.Last().Start+1))
				*datum = append(*datum, pumpTest.NewBloodGlucoseTargetStart(pointer.FromString("mmol/L"), *datum.Last().Start+1))
				Expect(datum.Last()).To(Equal((*datum)[2]))
			})
		})
	})

	Context("ParseBloodGlucoseTargetStartArrayMap", func() {
		// TODO
	})

	Context("NewBloodGlucoseTargetStartArrayMap", func() {
		It("is successful", func() {
			Expect(pump.NewBloodGlucoseTargetStartArrayMap()).To(Equal(&pump.BloodGlucoseTargetStartArrayMap{}))
		})
	})

	Context("BloodGlucoseTargetStartArrayMap", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *pump.BloodGlucoseTargetStartArrayMap, units *string), expectedErrors ...error) {
					datum := pump.NewBloodGlucoseTargetStartArrayMap()
					mutator(datum, units)
					dataTypesTest.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringAdapter(datum, units), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStartArrayMap, units *string) {},
				),
				Entry("empty",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStartArrayMap, units *string) {
						*datum = *pump.NewBloodGlucoseTargetStartArrayMap()
					},
				),
				Entry("empty name",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStartArrayMap, units *string) {
						datum.Set("", pumpTest.NewBloodGlucoseTargetStartArray(units))
					},
				),
				Entry("nil value",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStartArrayMap, units *string) { datum.Set("", nil) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/"),
				),
				Entry("single invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStartArrayMap, units *string) {
						invalid := pumpTest.NewBloodGlucoseTargetStartArray(units)
						(*invalid)[0].Start = nil
						datum.Set("one", invalid)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/one/0/start"),
				),
				Entry("single valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStartArrayMap, units *string) {
						datum.Set("one", pumpTest.NewBloodGlucoseTargetStartArray(units))
					},
				),
				Entry("multiple invalid",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStartArrayMap, units *string) {
						invalid := pumpTest.NewBloodGlucoseTargetStartArray(units)
						(*invalid)[0].Start = nil
						datum.Set("one", pumpTest.NewBloodGlucoseTargetStartArray(units))
						datum.Set("two", invalid)
						datum.Set("three", pumpTest.NewBloodGlucoseTargetStartArray(units))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/two/0/start"),
				),
				Entry("multiple valid",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStartArrayMap, units *string) {
						datum.Set("one", pumpTest.NewBloodGlucoseTargetStartArray(units))
						datum.Set("two", pumpTest.NewBloodGlucoseTargetStartArray(units))
						datum.Set("three", pumpTest.NewBloodGlucoseTargetStartArray(units))
					},
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStartArrayMap, units *string) {
						invalid := pumpTest.NewBloodGlucoseTargetStartArray(units)
						(*invalid)[0].Start = nil
						datum.Set("one", nil)
						datum.Set("two", invalid)
						datum.Set("three", pumpTest.NewBloodGlucoseTargetStartArray(units))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/one"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/two/0/start"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *pump.BloodGlucoseTargetStartArrayMap, units *string), expectator func(datum *pump.BloodGlucoseTargetStartArrayMap, expectedDatum *pump.BloodGlucoseTargetStartArrayMap, units *string)) {
					for _, origin := range structure.Origins() {
						datum := pumpTest.NewBloodGlucoseTargetStartArrayMap(units)
						mutator(datum, units)
						expectedDatum := pumpTest.CloneBloodGlucoseTargetStartArrayMap(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin), units)
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum, units)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStartArrayMap, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *pump.BloodGlucoseTargetStartArrayMap, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *pump.BloodGlucoseTargetStartArrayMap, units *string) {},
					nil,
				),
				Entry("does not modify the datum; start missing",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStartArrayMap, units *string) {
						for name := range *datum {
							(*(*datum)[name])[0].Start = nil
						}
					},
					nil,
				),
			)

			DescribeTable("normalizes the datum with structure external",
				func(units *string, mutator func(datum *pump.BloodGlucoseTargetStartArrayMap, units *string), expectator func(datum *pump.BloodGlucoseTargetStartArrayMap, expectedDatum *pump.BloodGlucoseTargetStartArrayMap, units *string)) {
					datum := pumpTest.NewBloodGlucoseTargetStartArrayMap(units)
					mutator(datum, units)
					expectedDatum := pumpTest.CloneBloodGlucoseTargetStartArrayMap(datum)
					normalizer := dataNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(structure.OriginExternal), units)
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum, units)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStartArrayMap, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *pump.BloodGlucoseTargetStartArrayMap, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *pump.BloodGlucoseTargetStartArrayMap, units *string) {},
					func(datum *pump.BloodGlucoseTargetStartArrayMap, expectedDatum *pump.BloodGlucoseTargetStartArrayMap, units *string) {
						for name := range *datum {
							for index := range *(*datum)[name] {
								dataBloodGlucoseTest.ExpectNormalizedTarget(&(*(*datum)[name])[index].Target, &(*(*expectedDatum)[name])[index].Target, units)
							}
						}
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *pump.BloodGlucoseTargetStartArrayMap, units *string) {},
					func(datum *pump.BloodGlucoseTargetStartArrayMap, expectedDatum *pump.BloodGlucoseTargetStartArrayMap, units *string) {
						for name := range *datum {
							for index := range *(*datum)[name] {
								dataBloodGlucoseTest.ExpectNormalizedTarget(&(*(*datum)[name])[index].Target, &(*(*expectedDatum)[name])[index].Target, units)
							}
						}
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *pump.BloodGlucoseTargetStartArrayMap, units *string), expectator func(datum *pump.BloodGlucoseTargetStartArrayMap, expectedDatum *pump.BloodGlucoseTargetStartArrayMap, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := pumpTest.NewBloodGlucoseTargetStartArrayMap(units)
						mutator(datum, units)
						expectedDatum := pumpTest.CloneBloodGlucoseTargetStartArrayMap(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin), units)
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum, units)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *pump.BloodGlucoseTargetStartArrayMap, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *pump.BloodGlucoseTargetStartArrayMap, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *pump.BloodGlucoseTargetStartArrayMap, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *pump.BloodGlucoseTargetStartArrayMap, units *string) {},
					nil,
				),
			)
		})
	})
})
