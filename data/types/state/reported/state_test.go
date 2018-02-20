package reported_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/state/reported"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewState(state string) *reported.State {
	datum := reported.NewState()
	datum.State = pointer.String(state)
	return datum
}

func CloneState(datum *reported.State) *reported.State {
	if datum == nil {
		return nil
	}
	clone := reported.NewState()
	clone.State = test.CloneString(datum.State)
	return clone
}

func NewStateArray(states ...*reported.State) *reported.StateArray {
	datum := reported.NewStateArray()
	*datum = append(*datum, states...)
	return datum
}

func CloneStateArray(datum *reported.StateArray) *reported.StateArray {
	if datum == nil {
		return nil
	}
	clone := reported.NewStateArray()
	for _, value := range *datum {
		*clone = append(*clone, CloneState(value))
	}
	return clone
}

var _ = Describe("State", func() {
	It("StateAlcohol is expected", func() {
		Expect(reported.StateAlcohol).To(Equal("alcohol"))
	})

	It("StateCycle is expected", func() {
		Expect(reported.StateCycle).To(Equal("cycle"))
	})

	It("StateHyperglycemiaSymptoms is expected", func() {
		Expect(reported.StateHyperglycemiaSymptoms).To(Equal("hyperglycemiaSymptoms"))
	})

	It("StateHypoglycemiaSymptoms is expected", func() {
		Expect(reported.StateHypoglycemiaSymptoms).To(Equal("hypoglycemiaSymptoms"))
	})

	It("StateIllness is expected", func() {
		Expect(reported.StateIllness).To(Equal("illness"))
	})

	It("StateStress is expected", func() {
		Expect(reported.StateStress).To(Equal("stress"))
	})

	It("States returns expected", func() {
		Expect(reported.States()).To(Equal([]string{"alcohol", "cycle", "hyperglycemiaSymptoms", "hypoglycemiaSymptoms", "illness", "stress"}))
	})

	Context("ParseState", func() {
		// TODO
	})

	Context("NewState", func() {
		It("is successful", func() {
			Expect(reported.NewState()).To(Equal(&reported.State{}))
		})
	})

	Context("State", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *reported.State), expectedErrors ...error) {
					datum := NewState("stress")
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *reported.State) {},
				),
				Entry("state missing",
					func(datum *reported.State) { datum.State = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/state"),
				),
				Entry("state invalid",
					func(datum *reported.State) { *datum = *NewState("invalid") },
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"alcohol", "cycle", "hyperglycemiaSymptoms", "hypoglycemiaSymptoms", "illness", "stress"}), "/state"),
				),
				Entry("state alcohol",
					func(datum *reported.State) { *datum = *NewState("alcohol") },
				),
				Entry("state cycle",
					func(datum *reported.State) { *datum = *NewState("cycle") },
				),
				Entry("state hyperglycemiaSymptoms",
					func(datum *reported.State) { *datum = *NewState("hyperglycemiaSymptoms") },
				),
				Entry("state hypoglycemiaSymptoms",
					func(datum *reported.State) { *datum = *NewState("hypoglycemiaSymptoms") },
				),
				Entry("state illness",
					func(datum *reported.State) { *datum = *NewState("illness") },
				),
				Entry("state stress",
					func(datum *reported.State) { *datum = *NewState("stress") },
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *reported.State)) {
					for _, origin := range structure.Origins() {
						datum := NewState("state")
						mutator(datum)
						expectedDatum := CloneState(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *reported.State) {},
				),
				Entry("does not modify the datum; alcohol",
					func(datum *reported.State) { *datum = *NewState("alcohol") },
				),
				Entry("does not modify the datum; cycle",
					func(datum *reported.State) { *datum = *NewState("cycle") },
				),
				Entry("does not modify the datum; hyperglycemiaSymptoms",
					func(datum *reported.State) { *datum = *NewState("hyperglycemiaSymptoms") },
				),
				Entry("does not modify the datum; hypoglycemiaSymptoms",
					func(datum *reported.State) { *datum = *NewState("hypoglycemiaSymptoms") },
				),
				Entry("does not modify the datum; illness",
					func(datum *reported.State) { *datum = *NewState("illness") },
				),
				Entry("does not modify the datum; stress",
					func(datum *reported.State) { *datum = *NewState("stress") },
				),
			)
		})
	})

	Context("ParseStateArray", func() {
		// TODO
	})

	Context("NewStateArray", func() {
		It("is successful", func() {
			Expect(reported.NewStateArray()).To(Equal(&reported.StateArray{}))
		})
	})

	Context("StateArray", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *reported.StateArray), expectedErrors ...error) {
					datum := NewStateArray(NewState("alcohol"), NewState("stress"))
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *reported.StateArray) {},
				),
				Entry("empty",
					func(datum *reported.StateArray) { *datum = *NewStateArray() },
				),
				Entry("nil",
					func(datum *reported.StateArray) { *datum = *NewStateArray(nil) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
				),
				Entry("single invalid",
					func(datum *reported.StateArray) { *datum = *NewStateArray(NewState("invalid")) },
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"alcohol", "cycle", "hyperglycemiaSymptoms", "hypoglycemiaSymptoms", "illness", "stress"}), "/0/state"),
				),
				Entry("single valid",
					func(datum *reported.StateArray) { *datum = *NewStateArray(NewState("alcohol")) },
				),
				Entry("multiple invalid",
					func(datum *reported.StateArray) {
						*datum = *NewStateArray(NewState("cycle"), NewState("invalid"), NewState("alcohol"), NewState("stress"))
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"alcohol", "cycle", "hyperglycemiaSymptoms", "hypoglycemiaSymptoms", "illness", "stress"}), "/1/state"),
				),
				Entry("multiple valid",
					func(datum *reported.StateArray) {
						*datum = *NewStateArray(NewState("cycle"), NewState("illness"), NewState("alcohol"), NewState("stress"))
					},
				),
				Entry("multiple errors",
					func(datum *reported.StateArray) {
						*datum = *NewStateArray(NewState("cycle"), nil, NewState("invalid"), NewState("stress"))
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"alcohol", "cycle", "hyperglycemiaSymptoms", "hypoglycemiaSymptoms", "illness", "stress"}), "/2/state"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *reported.StateArray)) {
					for _, origin := range structure.Origins() {
						datum := NewStateArray(NewState("alcohol"), NewState("stress"))
						mutator(datum)
						expectedDatum := CloneStateArray(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *reported.StateArray) {},
				),
				Entry("does not modify the datum; empty",
					func(datum *reported.StateArray) { *datum = *NewStateArray() },
				),
				Entry("does not modify the datum; nil",
					func(datum *reported.StateArray) { *datum = *NewStateArray(nil) },
				),
				Entry("does not modify the datum; single invalid",
					func(datum *reported.StateArray) { *datum = *NewStateArray(NewState("invalid")) },
				),
				Entry("does not modify the datum; single valid",
					func(datum *reported.StateArray) { *datum = *NewStateArray(NewState("alcohol")) },
				),
				Entry("does not modify the datum; multiple invalid",
					func(datum *reported.StateArray) {
						*datum = *NewStateArray(NewState("cycle"), NewState("invalid"), NewState("alcohol"), NewState("stress"))
					},
				),
				Entry("does not modify the datum; multiple valid",
					func(datum *reported.StateArray) {
						*datum = *NewStateArray(NewState("cycle"), NewState("illness"), NewState("alcohol"), NewState("stress"))
					},
				),
			)
		})
	})
})
