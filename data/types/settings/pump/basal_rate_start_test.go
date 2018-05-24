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

func NewBasalRateStart(startMinimum int) *pump.BasalRateStart {
	datum := pump.NewBasalRateStart()
	datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(pump.BasalRateStartRateMinimum, pump.BasalRateStartRateMaximum))
	if startMinimum == pump.BasalRateStartStartMinimum {
		datum.Start = pointer.FromInt(pump.BasalRateStartStartMinimum)
	} else {
		datum.Start = pointer.FromInt(test.RandomIntFromRange(startMinimum, pump.BasalRateStartStartMaximum))
	}
	return datum
}

func CloneBasalRateStart(datum *pump.BasalRateStart) *pump.BasalRateStart {
	if datum == nil {
		return nil
	}
	clone := pump.NewBasalRateStart()
	clone.Rate = test.CloneFloat64(datum.Rate)
	clone.Start = test.CloneInt(datum.Start)
	return clone
}

func NewBasalRateStartArray() *pump.BasalRateStartArray {
	datum := pump.NewBasalRateStartArray()
	*datum = append(*datum, NewBasalRateStart(pump.BasalRateStartStartMinimum))
	*datum = append(*datum, NewBasalRateStart(*datum.Last().Start+1))
	*datum = append(*datum, NewBasalRateStart(*datum.Last().Start+1))
	return datum
}

func CloneBasalRateStartArray(datumArray *pump.BasalRateStartArray) *pump.BasalRateStartArray {
	if datumArray == nil {
		return nil
	}
	clone := pump.NewBasalRateStartArray()
	for _, datum := range *datumArray {
		*clone = append(*clone, CloneBasalRateStart(datum))
	}
	return clone
}

func NewBasalRateStartArrayMap() *pump.BasalRateStartArrayMap {
	datum := pump.NewBasalRateStartArrayMap()
	datum.Set(testDataTypesBasal.NewScheduleName(), NewBasalRateStartArray())
	return datum
}

func CloneBasalRateStartArrayMap(datumArrayMap *pump.BasalRateStartArrayMap) *pump.BasalRateStartArrayMap {
	if datumArrayMap == nil {
		return nil
	}
	clone := pump.NewBasalRateStartArrayMap()
	for datumName, datumArray := range *datumArrayMap {
		clone.Set(datumName, CloneBasalRateStartArray(datumArray))
	}
	return clone
}

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
					datum := NewBasalRateStart(pump.BasalRateStartStartMinimum)
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithIntAdapter(datum, pointer.FromInt(pump.BasalRateStartStartMinimum)), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.BasalRateStart) {},
				),
				Entry("rate missing",
					func(datum *pump.BasalRateStart) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("rate out of range (lower)",
					func(datum *pump.BasalRateStart) { datum.Rate = pointer.FromFloat64(-0.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 100), "/rate"),
				),
				Entry("rate in range (lower)",
					func(datum *pump.BasalRateStart) { datum.Rate = pointer.FromFloat64(0.0) },
				),
				Entry("rate in range (upper)",
					func(datum *pump.BasalRateStart) { datum.Rate = pointer.FromFloat64(100.0) },
				),
				Entry("rate out of range (upper)",
					func(datum *pump.BasalRateStart) { datum.Rate = pointer.FromFloat64(100.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0, 100), "/rate"),
				),
				Entry("start missing",
					func(datum *pump.BasalRateStart) { datum.Start = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
				),
				Entry("multiple errors",
					func(datum *pump.BasalRateStart) {
						datum.Rate = nil
						datum.Start = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
				),
			)

			DescribeTable("validates the datum with minimum start",
				func(mutator func(datum *pump.BasalRateStart), expectedErrors ...error) {
					datum := NewBasalRateStart(pump.BasalRateStartStartMinimum)
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithIntAdapter(datum, pointer.FromInt(pump.BasalRateStartStartMinimum)), structure.Origins(), expectedErrors...)
				},
				Entry("start out of range (lower)",
					func(datum *pump.BasalRateStart) { datum.Start = pointer.FromInt(-1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo(-1, 0), "/start"),
				),
				Entry("start in range",
					func(datum *pump.BasalRateStart) { datum.Start = pointer.FromInt(0) },
				),
				Entry("start out of range (upper)",
					func(datum *pump.BasalRateStart) { datum.Start = pointer.FromInt(1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo(1, 0), "/start"),
				),
			)

			DescribeTable("validates the datum with non-minimum start",
				func(mutator func(datum *pump.BasalRateStart), expectedErrors ...error) {
					datum := NewBasalRateStart(pump.BasalRateStartStartMinimum + 1)
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithIntAdapter(datum, pointer.FromInt(pump.BasalRateStartStartMinimum+1)), structure.Origins(), expectedErrors...)
				},
				Entry("start out of range (lower)",
					func(datum *pump.BasalRateStart) { datum.Start = pointer.FromInt(0) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(0, 1, 86400000), "/start"),
				),
				Entry("start in range (lower)",
					func(datum *pump.BasalRateStart) { datum.Start = pointer.FromInt(1) },
				),
				Entry("start in range (upper)",
					func(datum *pump.BasalRateStart) { datum.Start = pointer.FromInt(86400000) },
				),
				Entry("start out of range (upper)",
					func(datum *pump.BasalRateStart) { datum.Start = pointer.FromInt(86400001) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(86400001, 1, 86400000), "/start"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.BasalRateStart)) {
					for _, origin := range structure.Origins() {
						datum := NewBasalRateStart(pump.BasalRateStartStartMinimum + 1)
						mutator(datum)
						expectedDatum := CloneBasalRateStart(datum)
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
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.BasalRateStartArray) {},
				),
				Entry("empty",
					func(datum *pump.BasalRateStartArray) { *datum = *pump.NewBasalRateStartArray() },
				),
				Entry("nil",
					func(datum *pump.BasalRateStartArray) { *datum = append(*datum, nil) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
				),
				Entry("single invalid",
					func(datum *pump.BasalRateStartArray) {
						invalid := NewBasalRateStart(pump.BasalRateStartStartMinimum)
						invalid.Rate = nil
						*datum = append(*datum, invalid)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0/rate"),
				),
				Entry("single valid",
					func(datum *pump.BasalRateStartArray) {
						*datum = append(*datum, NewBasalRateStart(pump.BasalRateStartStartMinimum))
					},
				),
				Entry("multiple invalid",
					func(datum *pump.BasalRateStartArray) {
						*datum = append(*datum, NewBasalRateStart(pump.BasalRateStartStartMinimum))
						invalid := NewBasalRateStart(*datum.Last().Start + 1)
						invalid.Rate = nil
						*datum = append(*datum, invalid)
						*datum = append(*datum, NewBasalRateStart(*datum.Last().Start+1))
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/rate"),
				),
				Entry("multiple valid",
					func(datum *pump.BasalRateStartArray) {
						*datum = append(*datum, NewBasalRateStart(pump.BasalRateStartStartMinimum))
						*datum = append(*datum, NewBasalRateStart(*datum.Last().Start+1))
						*datum = append(*datum, NewBasalRateStart(*datum.Last().Start+1))
					},
				),
				Entry("multiple errors",
					func(datum *pump.BasalRateStartArray) {
						invalid := NewBasalRateStart(pump.BasalRateStartStartMinimum)
						invalid.Rate = nil
						*datum = append(*datum, nil, invalid)
						*datum = append(*datum, nil, NewBasalRateStart(*datum.Last().Start+1))
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/rate"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/2"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.BasalRateStartArray)) {
					for _, origin := range structure.Origins() {
						datum := NewBasalRateStartArray()
						mutator(datum)
						expectedDatum := CloneBasalRateStartArray(datum)
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
				*datum = append(*datum, NewBasalRateStart(pump.BasalRateStartStartMinimum))
				Expect(datum.First()).To(Equal((*datum)[0]))
			})

			It("returns the first element if the array has multiple elements", func() {
				*datum = append(*datum, NewBasalRateStart(pump.BasalRateStartStartMinimum))
				*datum = append(*datum, NewBasalRateStart(*datum.Last().Start+1))
				*datum = append(*datum, NewBasalRateStart(*datum.Last().Start+1))
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
				*datum = append(*datum, NewBasalRateStart(pump.BasalRateStartStartMinimum))
				Expect(datum.Last()).To(Equal((*datum)[0]))
			})

			It("returns the last element if the array has multiple elements", func() {
				*datum = append(*datum, NewBasalRateStart(pump.BasalRateStartStartMinimum))
				*datum = append(*datum, NewBasalRateStart(*datum.Last().Start+1))
				*datum = append(*datum, NewBasalRateStart(*datum.Last().Start+1))
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
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.BasalRateStartArrayMap) {},
				),
				Entry("empty",
					func(datum *pump.BasalRateStartArrayMap) { *datum = *pump.NewBasalRateStartArrayMap() },
				),
				Entry("empty name",
					func(datum *pump.BasalRateStartArrayMap) { datum.Set("", NewBasalRateStartArray()) },
				),
				Entry("nil value",
					func(datum *pump.BasalRateStartArrayMap) { datum.Set("", nil) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/"),
				),
				Entry("single invalid",
					func(datum *pump.BasalRateStartArrayMap) {
						invalid := NewBasalRateStartArray()
						(*invalid)[0].Start = nil
						datum.Set("one", invalid)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/one/0/start"),
				),
				Entry("single valid",
					func(datum *pump.BasalRateStartArrayMap) {
						datum.Set("one", NewBasalRateStartArray())
					},
				),
				Entry("multiple invalid",
					func(datum *pump.BasalRateStartArrayMap) {
						invalid := NewBasalRateStartArray()
						(*invalid)[0].Start = nil
						datum.Set("one", NewBasalRateStartArray())
						datum.Set("two", invalid)
						datum.Set("three", NewBasalRateStartArray())
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/two/0/start"),
				),
				Entry("multiple valid",
					func(datum *pump.BasalRateStartArrayMap) {
						datum.Set("one", NewBasalRateStartArray())
						datum.Set("two", NewBasalRateStartArray())
						datum.Set("three", NewBasalRateStartArray())
					},
				),
				Entry("multiple errors",
					func(datum *pump.BasalRateStartArrayMap) {
						invalid := NewBasalRateStartArray()
						(*invalid)[0].Start = nil
						datum.Set("one", nil)
						datum.Set("two", invalid)
						datum.Set("three", NewBasalRateStartArray())
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/one"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/two/0/start"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.BasalRateStartArrayMap)) {
					for _, origin := range structure.Origins() {
						datum := NewBasalRateStartArrayMap()
						mutator(datum)
						expectedDatum := CloneBasalRateStartArrayMap(datum)
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
