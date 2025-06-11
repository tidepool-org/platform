package pump_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	pumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("CarbohydrateRatioStart", func() {
	It("CarbohydrateRatioStartAmountMaximum is expected", func() {
		Expect(pump.CarbohydrateRatioStartAmountMaximum).To(Equal(500.0))
	})

	It("CarbohydrateRatioStartAmountMinimum is expected", func() {
		Expect(pump.CarbohydrateRatioStartAmountMinimum).To(Equal(0.0))
	})

	It("CarbohydrateRatioStartStartMaximum is expected", func() {
		Expect(pump.CarbohydrateRatioStartStartMaximum).To(Equal(86400000))
	})

	It("CarbohydrateRatioStartStartMinimum is expected", func() {
		Expect(pump.CarbohydrateRatioStartStartMinimum).To(Equal(0))
	})

	Context("ParseCarbohydrateRatioStart", func() {
		// TODO
	})

	Context("NewCarbohydrateRatioStart", func() {
		It("is successful", func() {
			Expect(pump.NewCarbohydrateRatioStart()).To(Equal(&pump.CarbohydrateRatioStart{}))
		})
	})

	Context("CarbohydrateRatioStart", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.CarbohydrateRatioStart), expectedErrors ...error) {
					datum := pumpTest.NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum)
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithIntAdapter(datum, pointer.FromInt(pump.CarbohydrateRatioStartStartMinimum)), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.CarbohydrateRatioStart) {},
				),
				Entry("amount missing",
					func(datum *pump.CarbohydrateRatioStart) { datum.Amount = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("amount out of range (lower)",
					func(datum *pump.CarbohydrateRatioStart) { datum.Amount = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, pump.CarbohydrateRatioStartAmountMaximum), "/amount"),
				),
				Entry("amount in range (lower)",
					func(datum *pump.CarbohydrateRatioStart) { datum.Amount = pointer.FromFloat64(0.0) },
				),
				Entry("amount in range (upper)",
					func(datum *pump.CarbohydrateRatioStart) {
						datum.Amount = pointer.FromFloat64(pump.CarbohydrateRatioStartAmountMaximum)
					},
				),
				Entry("amount out of range (upper)",
					func(datum *pump.CarbohydrateRatioStart) {
						datum.Amount = pointer.FromFloat64(pump.CarbohydrateRatioStartAmountMaximum + 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(pump.CarbohydrateRatioStartAmountMaximum+0.1, 0, pump.CarbohydrateRatioStartAmountMaximum), "/amount"),
				),
				Entry("start missing",
					func(datum *pump.CarbohydrateRatioStart) { datum.Start = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
				),
				Entry("multiple errors",
					func(datum *pump.CarbohydrateRatioStart) {
						datum.Amount = nil
						datum.Start = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
				),
			)

			DescribeTable("validates the datum with minimum start",
				func(mutator func(datum *pump.CarbohydrateRatioStart), expectedErrors ...error) {
					datum := pumpTest.NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum)
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithIntAdapter(datum, pointer.FromInt(pump.CarbohydrateRatioStartStartMinimum)), structure.Origins(), expectedErrors...)
				},
				Entry("start out of range (lower)",
					func(datum *pump.CarbohydrateRatioStart) { datum.Start = pointer.FromInt(-1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo(-1, 0), "/start"),
				),
				Entry("start in range",
					func(datum *pump.CarbohydrateRatioStart) { datum.Start = pointer.FromInt(0) },
				),
				Entry("start out of range (upper)",
					func(datum *pump.CarbohydrateRatioStart) { datum.Start = pointer.FromInt(1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo(1, 0), "/start"),
				),
			)

			DescribeTable("validates the datum with non-minimum start",
				func(mutator func(datum *pump.CarbohydrateRatioStart), expectedErrors ...error) {
					datum := pumpTest.NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum + 1)
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithIntAdapter(datum, pointer.FromInt(pump.CarbohydrateRatioStartStartMinimum+1)), structure.Origins(), expectedErrors...)
				},
				Entry("start out of range (lower)",
					func(datum *pump.CarbohydrateRatioStart) { datum.Start = pointer.FromInt(0) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0, 1, 86400000), "/start"),
				),
				Entry("start in range (lower)",
					func(datum *pump.CarbohydrateRatioStart) { datum.Start = pointer.FromInt(1) },
				),
				Entry("start in range (upper)",
					func(datum *pump.CarbohydrateRatioStart) { datum.Start = pointer.FromInt(86400000) },
				),
				Entry("start out of range (upper)",
					func(datum *pump.CarbohydrateRatioStart) { datum.Start = pointer.FromInt(86400001) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(86400001, 1, 86400000), "/start"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.CarbohydrateRatioStart)) {
					for _, origin := range structure.Origins() {
						datum := pumpTest.NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum + 1)
						mutator(datum)
						expectedDatum := pumpTest.CloneCarbohydrateRatioStart(datum)
						normalizer := dataNormalizer.New(logTest.NewLogger())
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.CarbohydrateRatioStart) {},
				),
				Entry("does not modify the datum; amount missing",
					func(datum *pump.CarbohydrateRatioStart) { datum.Amount = nil },
				),
				Entry("does not modify the datum; start missing",
					func(datum *pump.CarbohydrateRatioStart) { datum.Start = nil },
				),
			)
		})
	})

	Context("ParseCarbohydrateRatioStartArray", func() {
		// TODO
	})

	Context("NewCarbohydrateRatioStartArray", func() {
		It("is successful", func() {
			Expect(pump.NewCarbohydrateRatioStartArray()).To(Equal(&pump.CarbohydrateRatioStartArray{}))
		})
	})

	Context("CarbohydrateRatioStartArray", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.CarbohydrateRatioStartArray), expectedErrors ...error) {
					datum := pump.NewCarbohydrateRatioStartArray()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.CarbohydrateRatioStartArray) {},
				),
				Entry("empty",
					func(datum *pump.CarbohydrateRatioStartArray) { *datum = *pump.NewCarbohydrateRatioStartArray() },
				),
				Entry("nil",
					func(datum *pump.CarbohydrateRatioStartArray) { *datum = append(*datum, nil) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
				),
				Entry("single invalid",
					func(datum *pump.CarbohydrateRatioStartArray) {
						invalid := pumpTest.NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum)
						invalid.Amount = nil
						*datum = append(*datum, invalid)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0/amount"),
				),
				Entry("single valid",
					func(datum *pump.CarbohydrateRatioStartArray) {
						*datum = append(*datum, pumpTest.NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum))
					},
				),
				Entry("multiple invalid",
					func(datum *pump.CarbohydrateRatioStartArray) {
						*datum = append(*datum, pumpTest.NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum))
						invalid := pumpTest.NewCarbohydrateRatioStart(*datum.Last().Start + 1)
						invalid.Amount = nil
						*datum = append(*datum, invalid)
						*datum = append(*datum, pumpTest.NewCarbohydrateRatioStart(*datum.Last().Start+1))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/amount"),
				),
				Entry("multiple valid",
					func(datum *pump.CarbohydrateRatioStartArray) {
						*datum = append(*datum, pumpTest.NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum))
						*datum = append(*datum, pumpTest.NewCarbohydrateRatioStart(*datum.Last().Start+1))
						*datum = append(*datum, pumpTest.NewCarbohydrateRatioStart(*datum.Last().Start+1))
					},
				),
				Entry("multiple errors",
					func(datum *pump.CarbohydrateRatioStartArray) {
						invalid := pumpTest.NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum)
						invalid.Amount = nil
						*datum = append(*datum, nil, invalid)
						*datum = append(*datum, nil, pumpTest.NewCarbohydrateRatioStart(*datum.Last().Start+1))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/amount"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/2"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.CarbohydrateRatioStartArray)) {
					for _, origin := range structure.Origins() {
						datum := pumpTest.NewCarbohydrateRatioStartArray()
						mutator(datum)
						expectedDatum := pumpTest.CloneCarbohydrateRatioStartArray(datum)
						normalizer := dataNormalizer.New(logTest.NewLogger())
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.CarbohydrateRatioStartArray) {},
				),
				Entry("does not modify the datum; amount missing",
					func(datum *pump.CarbohydrateRatioStartArray) { (*datum)[0].Amount = nil },
				),
				Entry("does not modify the datum; start missing",
					func(datum *pump.CarbohydrateRatioStartArray) { (*datum)[0].Start = nil },
				),
			)
		})

		Context("First", func() {
			var datum *pump.CarbohydrateRatioStartArray

			BeforeEach(func() {
				datum = pump.NewCarbohydrateRatioStartArray()
			})

			It("returns nil if array is empty", func() {
				Expect(datum.First()).To(BeNil())
			})

			It("returns the first element if the array has one element", func() {
				*datum = append(*datum, pumpTest.NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum))
				Expect(datum.First()).To(Equal((*datum)[0]))
			})

			It("returns the first element if the array has multiple elements", func() {
				*datum = append(*datum, pumpTest.NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum))
				*datum = append(*datum, pumpTest.NewCarbohydrateRatioStart(*datum.Last().Start+1))
				*datum = append(*datum, pumpTest.NewCarbohydrateRatioStart(*datum.Last().Start+1))
				Expect(datum.First()).To(Equal((*datum)[0]))
			})
		})

		Context("Last", func() {
			var datum *pump.CarbohydrateRatioStartArray

			BeforeEach(func() {
				datum = pump.NewCarbohydrateRatioStartArray()
			})

			It("returns nil if array is empty", func() {
				Expect(datum.Last()).To(BeNil())
			})

			It("returns the last element if the array has one element", func() {
				*datum = append(*datum, pumpTest.NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum))
				Expect(datum.Last()).To(Equal((*datum)[0]))
			})

			It("returns the last element if the array has multiple elements", func() {
				*datum = append(*datum, pumpTest.NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum))
				*datum = append(*datum, pumpTest.NewCarbohydrateRatioStart(*datum.Last().Start+1))
				*datum = append(*datum, pumpTest.NewCarbohydrateRatioStart(*datum.Last().Start+1))
				Expect(datum.Last()).To(Equal((*datum)[2]))
			})
		})
	})

	Context("ParseCarbohydrateRatioStartArrayMap", func() {
		// TODO
	})

	Context("NewCarbohydrateRatioStartArrayMap", func() {
		It("is successful", func() {
			Expect(pump.NewCarbohydrateRatioStartArrayMap()).To(Equal(&pump.CarbohydrateRatioStartArrayMap{}))
		})
	})

	Context("CarbohydrateRatioStartArrayMap", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.CarbohydrateRatioStartArrayMap), expectedErrors ...error) {
					datum := pump.NewCarbohydrateRatioStartArrayMap()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.CarbohydrateRatioStartArrayMap) {},
				),
				Entry("empty",
					func(datum *pump.CarbohydrateRatioStartArrayMap) { *datum = *pump.NewCarbohydrateRatioStartArrayMap() },
				),
				Entry("empty name",
					func(datum *pump.CarbohydrateRatioStartArrayMap) {
						datum.Set("", pumpTest.NewCarbohydrateRatioStartArray())
					},
				),
				Entry("nil value",
					func(datum *pump.CarbohydrateRatioStartArrayMap) { datum.Set("", nil) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/"),
				),
				Entry("single invalid",
					func(datum *pump.CarbohydrateRatioStartArrayMap) {
						invalid := pumpTest.NewCarbohydrateRatioStartArray()
						(*invalid)[0].Start = nil
						datum.Set("one", invalid)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/one/0/start"),
				),
				Entry("single valid",
					func(datum *pump.CarbohydrateRatioStartArrayMap) {
						datum.Set("one", pumpTest.NewCarbohydrateRatioStartArray())
					},
				),
				Entry("multiple invalid",
					func(datum *pump.CarbohydrateRatioStartArrayMap) {
						invalid := pumpTest.NewCarbohydrateRatioStartArray()
						(*invalid)[0].Start = nil
						datum.Set("one", pumpTest.NewCarbohydrateRatioStartArray())
						datum.Set("two", invalid)
						datum.Set("three", pumpTest.NewCarbohydrateRatioStartArray())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/two/0/start"),
				),
				Entry("multiple valid",
					func(datum *pump.CarbohydrateRatioStartArrayMap) {
						datum.Set("one", pumpTest.NewCarbohydrateRatioStartArray())
						datum.Set("two", pumpTest.NewCarbohydrateRatioStartArray())
						datum.Set("three", pumpTest.NewCarbohydrateRatioStartArray())
					},
				),
				Entry("multiple errors",
					func(datum *pump.CarbohydrateRatioStartArrayMap) {
						invalid := pumpTest.NewCarbohydrateRatioStartArray()
						(*invalid)[0].Start = nil
						datum.Set("one", nil)
						datum.Set("two", invalid)
						datum.Set("three", pumpTest.NewCarbohydrateRatioStartArray())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/one"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/two/0/start"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.CarbohydrateRatioStartArrayMap)) {
					for _, origin := range structure.Origins() {
						datum := pumpTest.NewCarbohydrateRatioStartArrayMap()
						mutator(datum)
						expectedDatum := pumpTest.CloneCarbohydrateRatioStartArrayMap(datum)
						normalizer := dataNormalizer.New(logTest.NewLogger())
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.CarbohydrateRatioStartArrayMap) {},
				),
			)
		})
	})
})
