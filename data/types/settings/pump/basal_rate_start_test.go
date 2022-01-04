package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	dataTypesSettingsPumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("BasalRateStart", func() {
	It("BasalRateStartRateMaximum is expected", func() {
		Expect(pump.BasalRateStartRateMaximum).To(Equal(100.0))
	})

	It("BasalRateStartRateMinimum is expected", func() {
		Expect(pump.BasalRateStartRateMinimum).To(Equal(0.0))
	})

	It("BasalRateStartStartMaximum is expected", func() {
		Expect(pump.BasalRateStartStartMaximum).To(Equal(86400000))
	})

	It("BasalRateStartStartMinimum is expected", func() {
		Expect(pump.BasalRateStartStartMinimum).To(Equal(0))
	})

	Context("ParseBasalRateStart", func() {
		// TODO
	})

	Context("NewBasalRateStart", func() {
		It("is successful", func() {
			Expect(pump.NewBasalRateStart()).To(Equal(&pump.BasalRateStart{}))
		})
	})

	Context("BasalRateStart", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.BasalRateStart), expectedErrors ...error) {
					datum := dataTypesSettingsPumpTest.NewBasalRateStart(pump.BasalRateStartStartMinimum)
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithIntAdapter(datum, pointer.FromInt(pump.BasalRateStartStartMinimum)), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.BasalRateStart) {},
				),
				Entry("rate missing",
					func(datum *pump.BasalRateStart) { datum.Rate = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("rate out of range (lower)",
					func(datum *pump.BasalRateStart) { datum.Rate = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 100), "/rate"),
				),
				Entry("rate in range (lower)",
					func(datum *pump.BasalRateStart) { datum.Rate = pointer.FromFloat64(0.0) },
				),
				Entry("rate in range (upper)",
					func(datum *pump.BasalRateStart) { datum.Rate = pointer.FromFloat64(100.0) },
				),
				Entry("rate out of range (upper)",
					func(datum *pump.BasalRateStart) { datum.Rate = pointer.FromFloat64(100.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0, 100), "/rate"),
				),
				Entry("start missing",
					func(datum *pump.BasalRateStart) { datum.Start = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
				),
				Entry("multiple errors",
					func(datum *pump.BasalRateStart) {
						datum.Rate = nil
						datum.Start = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
				),
			)

			DescribeTable("validates the datum with minimum start",
				func(mutator func(datum *pump.BasalRateStart), expectedErrors ...error) {
					datum := dataTypesSettingsPumpTest.NewBasalRateStart(pump.BasalRateStartStartMinimum)
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithIntAdapter(datum, pointer.FromInt(pump.BasalRateStartStartMinimum)), structure.Origins(), expectedErrors...)
				},
				Entry("start out of range (lower)",
					func(datum *pump.BasalRateStart) { datum.Start = pointer.FromInt(-1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo(-1, 0), "/start"),
				),
				Entry("start in range",
					func(datum *pump.BasalRateStart) { datum.Start = pointer.FromInt(0) },
				),
				Entry("start out of range (upper)",
					func(datum *pump.BasalRateStart) { datum.Start = pointer.FromInt(1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo(1, 0), "/start"),
				),
			)

			DescribeTable("validates the datum with non-minimum start",
				func(mutator func(datum *pump.BasalRateStart), expectedErrors ...error) {
					datum := dataTypesSettingsPumpTest.NewBasalRateStart(pump.BasalRateStartStartMinimum + 1)
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithIntAdapter(datum, pointer.FromInt(pump.BasalRateStartStartMinimum+1)), structure.Origins(), expectedErrors...)
				},
				Entry("start out of range (lower)",
					func(datum *pump.BasalRateStart) { datum.Start = pointer.FromInt(0) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0, 1, 86400000), "/start"),
				),
				Entry("start in range (lower)",
					func(datum *pump.BasalRateStart) { datum.Start = pointer.FromInt(1) },
				),
				Entry("start in range (upper)",
					func(datum *pump.BasalRateStart) { datum.Start = pointer.FromInt(86400000) },
				),
				Entry("start out of range (upper)",
					func(datum *pump.BasalRateStart) { datum.Start = pointer.FromInt(86400001) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(86400001, 1, 86400000), "/start"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.BasalRateStart)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesSettingsPumpTest.NewBasalRateStart(pump.BasalRateStartStartMinimum + 1)
						mutator(datum)
						expectedDatum := dataTypesSettingsPumpTest.CloneBasalRateStart(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.BasalRateStart) {},
				),
				Entry("does not modify the datum; rate missing",
					func(datum *pump.BasalRateStart) { datum.Rate = nil },
				),
				Entry("does not modify the datum; start missing",
					func(datum *pump.BasalRateStart) { datum.Start = nil },
				),
			)
		})
	})

	Context("ParseBasalRateStartArray", func() {
		// TODO
	})

	Context("NewBasalRateStartArray", func() {
		It("is successful", func() {
			Expect(pump.NewBasalRateStartArray()).To(Equal(&pump.BasalRateStartArray{}))
		})
	})

	Context("BasalRateStartArray", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.BasalRateStartArray), expectedErrors ...error) {
					datum := pump.NewBasalRateStartArray()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.BasalRateStartArray) {},
				),
				Entry("empty",
					func(datum *pump.BasalRateStartArray) { *datum = *pump.NewBasalRateStartArray() },
				),
				Entry("nil",
					func(datum *pump.BasalRateStartArray) { *datum = append(*datum, nil) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
				),
				Entry("single invalid",
					func(datum *pump.BasalRateStartArray) {
						invalid := dataTypesSettingsPumpTest.NewBasalRateStart(pump.BasalRateStartStartMinimum)
						invalid.Rate = nil
						*datum = append(*datum, invalid)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0/rate"),
				),
				Entry("single valid",
					func(datum *pump.BasalRateStartArray) {
						*datum = append(*datum, dataTypesSettingsPumpTest.NewBasalRateStart(pump.BasalRateStartStartMinimum))
					},
				),
				Entry("multiple invalid",
					func(datum *pump.BasalRateStartArray) {
						*datum = append(*datum, dataTypesSettingsPumpTest.NewBasalRateStart(pump.BasalRateStartStartMinimum))
						invalid := dataTypesSettingsPumpTest.NewBasalRateStart(*datum.Last().Start + 1)
						invalid.Rate = nil
						*datum = append(*datum, invalid)
						*datum = append(*datum, dataTypesSettingsPumpTest.NewBasalRateStart(*datum.Last().Start+1))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/rate"),
				),
				Entry("multiple valid",
					func(datum *pump.BasalRateStartArray) {
						*datum = append(*datum, dataTypesSettingsPumpTest.NewBasalRateStart(pump.BasalRateStartStartMinimum))
						*datum = append(*datum, dataTypesSettingsPumpTest.NewBasalRateStart(*datum.Last().Start+1))
						*datum = append(*datum, dataTypesSettingsPumpTest.NewBasalRateStart(*datum.Last().Start+1))
					},
				),
				Entry("multiple errors",
					func(datum *pump.BasalRateStartArray) {
						invalid := dataTypesSettingsPumpTest.NewBasalRateStart(pump.BasalRateStartStartMinimum)
						invalid.Rate = nil
						*datum = append(*datum, nil, invalid)
						*datum = append(*datum, nil, dataTypesSettingsPumpTest.NewBasalRateStart(*datum.Last().Start+1))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/rate"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/2"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.BasalRateStartArray)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesSettingsPumpTest.NewBasalRateStartArray()
						mutator(datum)
						expectedDatum := dataTypesSettingsPumpTest.CloneBasalRateStartArray(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.BasalRateStartArray) {},
				),
				Entry("does not modify the datum; rate missing",
					func(datum *pump.BasalRateStartArray) { (*datum)[0].Rate = nil },
				),
				Entry("does not modify the datum; start missing",
					func(datum *pump.BasalRateStartArray) { (*datum)[0].Start = nil },
				),
			)
		})

		Context("First", func() {
			var datum *pump.BasalRateStartArray

			BeforeEach(func() {
				datum = pump.NewBasalRateStartArray()
			})

			It("returns nil if array is empty", func() {
				Expect(datum.First()).To(BeNil())
			})

			It("returns the first element if the array has one element", func() {
				*datum = append(*datum, dataTypesSettingsPumpTest.NewBasalRateStart(pump.BasalRateStartStartMinimum))
				Expect(datum.First()).To(Equal((*datum)[0]))
			})

			It("returns the first element if the array has multiple elements", func() {
				*datum = append(*datum, dataTypesSettingsPumpTest.NewBasalRateStart(pump.BasalRateStartStartMinimum))
				*datum = append(*datum, dataTypesSettingsPumpTest.NewBasalRateStart(*datum.Last().Start+1))
				*datum = append(*datum, dataTypesSettingsPumpTest.NewBasalRateStart(*datum.Last().Start+1))
				Expect(datum.First()).To(Equal((*datum)[0]))
			})
		})

		Context("Last", func() {
			var datum *pump.BasalRateStartArray

			BeforeEach(func() {
				datum = pump.NewBasalRateStartArray()
			})

			It("returns nil if array is empty", func() {
				Expect(datum.Last()).To(BeNil())
			})

			It("returns the last element if the array has one element", func() {
				*datum = append(*datum, dataTypesSettingsPumpTest.NewBasalRateStart(pump.BasalRateStartStartMinimum))
				Expect(datum.Last()).To(Equal((*datum)[0]))
			})

			It("returns the last element if the array has multiple elements", func() {
				*datum = append(*datum, dataTypesSettingsPumpTest.NewBasalRateStart(pump.BasalRateStartStartMinimum))
				*datum = append(*datum, dataTypesSettingsPumpTest.NewBasalRateStart(*datum.Last().Start+1))
				*datum = append(*datum, dataTypesSettingsPumpTest.NewBasalRateStart(*datum.Last().Start+1))
				Expect(datum.Last()).To(Equal((*datum)[2]))
			})
		})
	})

	Context("ParseBasalRateStartArrayMap", func() {
		// TODO
	})

	Context("NewBasalRateStartArrayMap", func() {
		It("is successful", func() {
			Expect(pump.NewBasalRateStartArrayMap()).To(Equal(&pump.BasalRateStartArrayMap{}))
		})
	})

	Context("BasalRateStartArrayMap", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.BasalRateStartArrayMap), expectedErrors ...error) {
					datum := pump.NewBasalRateStartArrayMap()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.BasalRateStartArrayMap) {},
				),
				Entry("empty",
					func(datum *pump.BasalRateStartArrayMap) { *datum = *pump.NewBasalRateStartArrayMap() },
				),
				Entry("empty name",
					func(datum *pump.BasalRateStartArrayMap) {
						datum.Set("", dataTypesSettingsPumpTest.NewBasalRateStartArray())
					},
				),
				Entry("nil value",
					func(datum *pump.BasalRateStartArrayMap) { datum.Set("", nil) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/"),
				),
				Entry("single invalid",
					func(datum *pump.BasalRateStartArrayMap) {
						invalid := dataTypesSettingsPumpTest.NewBasalRateStartArray()
						(*invalid)[0].Start = nil
						datum.Set("one", invalid)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/one/0/start"),
				),
				Entry("single valid",
					func(datum *pump.BasalRateStartArrayMap) {
						datum.Set("one", dataTypesSettingsPumpTest.NewBasalRateStartArray())
					},
				),
				Entry("multiple invalid",
					func(datum *pump.BasalRateStartArrayMap) {
						invalid := dataTypesSettingsPumpTest.NewBasalRateStartArray()
						(*invalid)[0].Start = nil
						datum.Set("one", dataTypesSettingsPumpTest.NewBasalRateStartArray())
						datum.Set("two", invalid)
						datum.Set("three", dataTypesSettingsPumpTest.NewBasalRateStartArray())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/two/0/start"),
				),
				Entry("multiple valid",
					func(datum *pump.BasalRateStartArrayMap) {
						datum.Set("one", dataTypesSettingsPumpTest.NewBasalRateStartArray())
						datum.Set("two", dataTypesSettingsPumpTest.NewBasalRateStartArray())
						datum.Set("three", dataTypesSettingsPumpTest.NewBasalRateStartArray())
					},
				),
				Entry("multiple errors",
					func(datum *pump.BasalRateStartArrayMap) {
						invalid := dataTypesSettingsPumpTest.NewBasalRateStartArray()
						(*invalid)[0].Start = nil
						datum.Set("one", nil)
						datum.Set("two", invalid)
						datum.Set("three", dataTypesSettingsPumpTest.NewBasalRateStartArray())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/one"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/two/0/start"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.BasalRateStartArrayMap)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesSettingsPumpTest.NewBasalRateStartArrayMap()
						mutator(datum)
						expectedDatum := dataTypesSettingsPumpTest.CloneBasalRateStartArrayMap(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.BasalRateStartArrayMap) {},
				),
			)
		})
	})
})
