package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	testDataTypesBasal "github.com/tidepool-org/platform/data/types/basal/test"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewCarbohydrateRatioStart(startMinimum int) *pump.CarbohydrateRatioStart {
	datum := pump.NewCarbohydrateRatioStart()
	datum.Amount = pointer.Float64(test.RandomFloat64FromRange(pump.CarbohydrateRatioStartAmountMinimum, pump.CarbohydrateRatioStartAmountMaximum))
	if startMinimum == pump.CarbohydrateRatioStartStartMinimum {
		datum.Start = pointer.Int(pump.CarbohydrateRatioStartStartMinimum)
	} else {
		datum.Start = pointer.Int(test.RandomIntFromRange(startMinimum, pump.CarbohydrateRatioStartStartMaximum))
	}
	return datum
}

func CloneCarbohydrateRatioStart(datum *pump.CarbohydrateRatioStart) *pump.CarbohydrateRatioStart {
	if datum == nil {
		return nil
	}
	clone := pump.NewCarbohydrateRatioStart()
	clone.Amount = test.CloneFloat64(datum.Amount)
	clone.Start = test.CloneInt(datum.Start)
	return clone
}

func NewCarbohydrateRatioStartArray() *pump.CarbohydrateRatioStartArray {
	datum := pump.NewCarbohydrateRatioStartArray()
	*datum = append(*datum, NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum))
	*datum = append(*datum, NewCarbohydrateRatioStart(*datum.Last().Start+1))
	*datum = append(*datum, NewCarbohydrateRatioStart(*datum.Last().Start+1))
	return datum
}

func CloneCarbohydrateRatioStartArray(datumArray *pump.CarbohydrateRatioStartArray) *pump.CarbohydrateRatioStartArray {
	if datumArray == nil {
		return nil
	}
	clone := pump.NewCarbohydrateRatioStartArray()
	for _, datum := range *datumArray {
		*clone = append(*clone, CloneCarbohydrateRatioStart(datum))
	}
	return clone
}

func NewCarbohydrateRatioStartArrayMap() *pump.CarbohydrateRatioStartArrayMap {
	datum := pump.NewCarbohydrateRatioStartArrayMap()
	datum.Set(testDataTypesBasal.NewScheduleName(), NewCarbohydrateRatioStartArray())
	return datum
}

func CloneCarbohydrateRatioStartArrayMap(datumArrayMap *pump.CarbohydrateRatioStartArrayMap) *pump.CarbohydrateRatioStartArrayMap {
	if datumArrayMap == nil {
		return nil
	}
	clone := pump.NewCarbohydrateRatioStartArrayMap()
	for datumName, datumArray := range *datumArrayMap {
		clone.Set(datumName, CloneCarbohydrateRatioStartArray(datumArray))
	}
	return clone
}

var _ = Describe("CarbohydrateRatioStart", func() {
	It("CarbohydrateRatioStartAmountMaximum is expected", func() {
		Expect(pump.CarbohydrateRatioStartAmountMaximum).To(Equal(250.0))
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
					datum := NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum)
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithIntAdapter(datum, pointer.Int(pump.CarbohydrateRatioStartStartMinimum)), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.CarbohydrateRatioStart) {},
				),
				Entry("amount missing",
					func(datum *pump.CarbohydrateRatioStart) { datum.Amount = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("amount out of range (lower)",
					func(datum *pump.CarbohydrateRatioStart) { datum.Amount = pointer.Float64(-0.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 250), "/amount"),
				),
				Entry("amount in range (lower)",
					func(datum *pump.CarbohydrateRatioStart) { datum.Amount = pointer.Float64(0.0) },
				),
				Entry("amount in range (upper)",
					func(datum *pump.CarbohydrateRatioStart) { datum.Amount = pointer.Float64(250.0) },
				),
				Entry("amount out of range (upper)",
					func(datum *pump.CarbohydrateRatioStart) { datum.Amount = pointer.Float64(250.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(250.1, 0, 250), "/amount"),
				),
				Entry("start missing",
					func(datum *pump.CarbohydrateRatioStart) { datum.Start = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
				),
				Entry("multiple errors",
					func(datum *pump.CarbohydrateRatioStart) {
						datum.Amount = nil
						datum.Start = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
				),
			)

			DescribeTable("validates the datum with minimum start",
				func(mutator func(datum *pump.CarbohydrateRatioStart), expectedErrors ...error) {
					datum := NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum)
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithIntAdapter(datum, pointer.Int(pump.CarbohydrateRatioStartStartMinimum)), structure.Origins(), expectedErrors...)
				},
				Entry("start out of range (lower)",
					func(datum *pump.CarbohydrateRatioStart) { datum.Start = pointer.Int(-1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo(-1, 0), "/start"),
				),
				Entry("start in range",
					func(datum *pump.CarbohydrateRatioStart) { datum.Start = pointer.Int(0) },
				),
				Entry("start out of range (upper)",
					func(datum *pump.CarbohydrateRatioStart) { datum.Start = pointer.Int(1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo(1, 0), "/start"),
				),
			)

			DescribeTable("validates the datum with non-minimum start",
				func(mutator func(datum *pump.CarbohydrateRatioStart), expectedErrors ...error) {
					datum := NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum + 1)
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithIntAdapter(datum, pointer.Int(pump.CarbohydrateRatioStartStartMinimum+1)), structure.Origins(), expectedErrors...)
				},
				Entry("start out of range (lower)",
					func(datum *pump.CarbohydrateRatioStart) { datum.Start = pointer.Int(0) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(0, 1, 86400000), "/start"),
				),
				Entry("start in range (lower)",
					func(datum *pump.CarbohydrateRatioStart) { datum.Start = pointer.Int(1) },
				),
				Entry("start in range (upper)",
					func(datum *pump.CarbohydrateRatioStart) { datum.Start = pointer.Int(86400000) },
				),
				Entry("start out of range (upper)",
					func(datum *pump.CarbohydrateRatioStart) { datum.Start = pointer.Int(86400001) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(86400001, 1, 86400000), "/start"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.CarbohydrateRatioStart)) {
					for _, origin := range structure.Origins() {
						datum := NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum + 1)
						mutator(datum)
						expectedDatum := CloneCarbohydrateRatioStart(datum)
						normalizer := dataNormalizer.New()
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
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.CarbohydrateRatioStartArray) {},
				),
				Entry("empty",
					func(datum *pump.CarbohydrateRatioStartArray) { *datum = *pump.NewCarbohydrateRatioStartArray() },
				),
				Entry("nil",
					func(datum *pump.CarbohydrateRatioStartArray) { *datum = append(*datum, nil) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
				),
				Entry("single invalid",
					func(datum *pump.CarbohydrateRatioStartArray) {
						invalid := NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum)
						invalid.Amount = nil
						*datum = append(*datum, invalid)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0/amount"),
				),
				Entry("single valid",
					func(datum *pump.CarbohydrateRatioStartArray) {
						*datum = append(*datum, NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum))
					},
				),
				Entry("multiple invalid",
					func(datum *pump.CarbohydrateRatioStartArray) {
						*datum = append(*datum, NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum))
						invalid := NewCarbohydrateRatioStart(*datum.Last().Start + 1)
						invalid.Amount = nil
						*datum = append(*datum, invalid)
						*datum = append(*datum, NewCarbohydrateRatioStart(*datum.Last().Start+1))
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/amount"),
				),
				Entry("multiple valid",
					func(datum *pump.CarbohydrateRatioStartArray) {
						*datum = append(*datum, NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum))
						*datum = append(*datum, NewCarbohydrateRatioStart(*datum.Last().Start+1))
						*datum = append(*datum, NewCarbohydrateRatioStart(*datum.Last().Start+1))
					},
				),
				Entry("multiple errors",
					func(datum *pump.CarbohydrateRatioStartArray) {
						invalid := NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum)
						invalid.Amount = nil
						*datum = append(*datum, nil, invalid)
						*datum = append(*datum, nil, NewCarbohydrateRatioStart(*datum.Last().Start+1))
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/amount"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/2"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.CarbohydrateRatioStartArray)) {
					for _, origin := range structure.Origins() {
						datum := NewCarbohydrateRatioStartArray()
						mutator(datum)
						expectedDatum := CloneCarbohydrateRatioStartArray(datum)
						normalizer := dataNormalizer.New()
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
				*datum = append(*datum, NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum))
				Expect(datum.First()).To(Equal((*datum)[0]))
			})

			It("returns the first element if the array has multiple elements", func() {
				*datum = append(*datum, NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum))
				*datum = append(*datum, NewCarbohydrateRatioStart(*datum.Last().Start+1))
				*datum = append(*datum, NewCarbohydrateRatioStart(*datum.Last().Start+1))
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
				*datum = append(*datum, NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum))
				Expect(datum.Last()).To(Equal((*datum)[0]))
			})

			It("returns the last element if the array has multiple elements", func() {
				*datum = append(*datum, NewCarbohydrateRatioStart(pump.CarbohydrateRatioStartStartMinimum))
				*datum = append(*datum, NewCarbohydrateRatioStart(*datum.Last().Start+1))
				*datum = append(*datum, NewCarbohydrateRatioStart(*datum.Last().Start+1))
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
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.CarbohydrateRatioStartArrayMap) {},
				),
				Entry("empty",
					func(datum *pump.CarbohydrateRatioStartArrayMap) { *datum = *pump.NewCarbohydrateRatioStartArrayMap() },
				),
				Entry("empty name",
					func(datum *pump.CarbohydrateRatioStartArrayMap) { datum.Set("", NewCarbohydrateRatioStartArray()) },
				),
				Entry("nil value",
					func(datum *pump.CarbohydrateRatioStartArrayMap) { datum.Set("", nil) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/"),
				),
				Entry("single invalid",
					func(datum *pump.CarbohydrateRatioStartArrayMap) {
						invalid := NewCarbohydrateRatioStartArray()
						(*invalid)[0].Start = nil
						datum.Set("one", invalid)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/one/0/start"),
				),
				Entry("single valid",
					func(datum *pump.CarbohydrateRatioStartArrayMap) {
						datum.Set("one", NewCarbohydrateRatioStartArray())
					},
				),
				Entry("multiple invalid",
					func(datum *pump.CarbohydrateRatioStartArrayMap) {
						invalid := NewCarbohydrateRatioStartArray()
						(*invalid)[0].Start = nil
						datum.Set("one", NewCarbohydrateRatioStartArray())
						datum.Set("two", invalid)
						datum.Set("three", NewCarbohydrateRatioStartArray())
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/two/0/start"),
				),
				Entry("multiple valid",
					func(datum *pump.CarbohydrateRatioStartArrayMap) {
						datum.Set("one", NewCarbohydrateRatioStartArray())
						datum.Set("two", NewCarbohydrateRatioStartArray())
						datum.Set("three", NewCarbohydrateRatioStartArray())
					},
				),
				Entry("multiple errors",
					func(datum *pump.CarbohydrateRatioStartArrayMap) {
						invalid := NewCarbohydrateRatioStartArray()
						(*invalid)[0].Start = nil
						datum.Set("one", nil)
						datum.Set("two", invalid)
						datum.Set("three", NewCarbohydrateRatioStartArray())
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/one"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/two/0/start"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.CarbohydrateRatioStartArrayMap)) {
					for _, origin := range structure.Origins() {
						datum := NewCarbohydrateRatioStartArrayMap()
						mutator(datum)
						expectedDatum := CloneCarbohydrateRatioStartArrayMap(datum)
						normalizer := dataNormalizer.New()
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
