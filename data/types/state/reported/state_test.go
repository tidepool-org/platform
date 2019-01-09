package reported_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/state/reported"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewState(state string) *reported.State {
	datum := reported.NewState()
	datum.Severity = pointer.FromInt(test.RandomIntFromRange(reported.StateSeverityMinimum, reported.StateSeverityMaximum))
	datum.State = pointer.FromString(state)
	if datum.State != nil && *datum.State == reported.StateStateOther {
		datum.StateOther = pointer.FromString(test.RandomStringFromRange(1, 100))
	}
	return datum
}

func CloneState(datum *reported.State) *reported.State {
	if datum == nil {
		return nil
	}
	clone := reported.NewState()
	clone.Severity = pointer.CloneInt(datum.Severity)
	clone.State = pointer.CloneString(datum.State)
	clone.StateOther = pointer.CloneString(datum.StateOther)
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
	It("StateSeverityMaximum is expected", func() {
		Expect(reported.StateSeverityMaximum).To(Equal(10))
	})

	It("StateSeverityMinimum is expected", func() {
		Expect(reported.StateSeverityMinimum).To(Equal(0))
	})

	It("StateStateAlcohol is expected", func() {
		Expect(reported.StateStateAlcohol).To(Equal("alcohol"))
	})

	It("StateStateCycle is expected", func() {
		Expect(reported.StateStateCycle).To(Equal("cycle"))
	})

	It("StateStateHyperglycemiaSymptoms is expected", func() {
		Expect(reported.StateStateHyperglycemiaSymptoms).To(Equal("hyperglycemiaSymptoms"))
	})

	It("StateStateHypoglycemiaSymptoms is expected", func() {
		Expect(reported.StateStateHypoglycemiaSymptoms).To(Equal("hypoglycemiaSymptoms"))
	})

	It("StateStateIllness is expected", func() {
		Expect(reported.StateStateIllness).To(Equal("illness"))
	})

	It("StateStateOther is expected", func() {
		Expect(reported.StateStateOther).To(Equal("other"))
	})

	It("StateStateOtherLengthMaximum is expected", func() {
		Expect(reported.StateStateOtherLengthMaximum).To(Equal(100))
	})

	It("StateStateStress is expected", func() {
		Expect(reported.StateStateStress).To(Equal("stress"))
	})

	It("StateStates returns expected", func() {
		Expect(reported.StateStates()).To(Equal([]string{"alcohol", "cycle", "hyperglycemiaSymptoms", "hypoglycemiaSymptoms", "illness", "other", "stress"}))
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
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *reported.State) {},
				),
				Entry("severity missing",
					func(datum *reported.State) { datum.Severity = nil },
				),
				Entry("severity out of range (lower)",
					func(datum *reported.State) { datum.Severity = pointer.FromInt(-1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 10), "/severity"),
				),
				Entry("severity in range (lower)",
					func(datum *reported.State) { datum.Severity = pointer.FromInt(0) },
				),
				Entry("severity in range (upper)",
					func(datum *reported.State) { datum.Severity = pointer.FromInt(10) },
				),
				Entry("severity out of range (upper)",
					func(datum *reported.State) { datum.Severity = pointer.FromInt(11) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(11, 0, 10), "/severity"),
				),
				Entry("state missing; state other missing",
					func(datum *reported.State) {
						datum.State = nil
						datum.StateOther = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/state"),
				),
				Entry("state missing; state other exists",
					func(datum *reported.State) {
						datum.State = nil
						datum.StateOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/state"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/stateOther"),
				),
				Entry("state invalid; state other missing",
					func(datum *reported.State) {
						datum.State = pointer.FromString("invalid")
						datum.StateOther = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"alcohol", "cycle", "hyperglycemiaSymptoms", "hypoglycemiaSymptoms", "illness", "other", "stress"}), "/state"),
				),
				Entry("state invalid; state other exists",
					func(datum *reported.State) {
						datum.State = pointer.FromString("invalid")
						datum.StateOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"alcohol", "cycle", "hyperglycemiaSymptoms", "hypoglycemiaSymptoms", "illness", "other", "stress"}), "/state"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/stateOther"),
				),
				Entry("state alcohol; state other missing",
					func(datum *reported.State) {
						datum.State = pointer.FromString("alcohol")
						datum.StateOther = nil
					},
				),
				Entry("state alcohol; state other exists",
					func(datum *reported.State) {
						datum.State = pointer.FromString("alcohol")
						datum.StateOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/stateOther"),
				),
				Entry("state cycle; state other missing",
					func(datum *reported.State) {
						datum.State = pointer.FromString("cycle")
						datum.StateOther = nil
					},
				),
				Entry("state cycle; state other exists",
					func(datum *reported.State) {
						datum.State = pointer.FromString("cycle")
						datum.StateOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/stateOther"),
				),
				Entry("state hyperglycemiaSymptoms; state other missing",
					func(datum *reported.State) {
						datum.State = pointer.FromString("hyperglycemiaSymptoms")
						datum.StateOther = nil
					},
				),
				Entry("state hyperglycemiaSymptoms; state other exists",
					func(datum *reported.State) {
						datum.State = pointer.FromString("hyperglycemiaSymptoms")
						datum.StateOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/stateOther"),
				),
				Entry("state hypoglycemiaSymptoms; state other missing",
					func(datum *reported.State) {
						datum.State = pointer.FromString("hypoglycemiaSymptoms")
						datum.StateOther = nil
					},
				),
				Entry("state hypoglycemiaSymptoms; state other exists",
					func(datum *reported.State) {
						datum.State = pointer.FromString("hypoglycemiaSymptoms")
						datum.StateOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/stateOther"),
				),
				Entry("state illness; state other missing",
					func(datum *reported.State) {
						datum.State = pointer.FromString("illness")
						datum.StateOther = nil
					},
				),
				Entry("state illness; state other exists",
					func(datum *reported.State) {
						datum.State = pointer.FromString("illness")
						datum.StateOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/stateOther"),
				),
				Entry("state other; state other missing",
					func(datum *reported.State) {
						datum.State = pointer.FromString("other")
						datum.StateOther = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/stateOther"),
				),
				Entry("state other; state other empty",
					func(datum *reported.State) {
						datum.State = pointer.FromString("other")
						datum.StateOther = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/stateOther"),
				),
				Entry("state other; state other length in range (upper)",
					func(datum *reported.State) {
						datum.State = pointer.FromString("other")
						datum.StateOther = pointer.FromString(test.RandomStringFromRange(100, 100))
					},
				),
				Entry("state other; state other length out of range (upper)",
					func(datum *reported.State) {
						datum.State = pointer.FromString("other")
						datum.StateOther = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/stateOther"),
				),
				Entry("state stress; state other missing",
					func(datum *reported.State) {
						datum.State = pointer.FromString("stress")
						datum.StateOther = nil
					},
				),
				Entry("state stress; state other exists",
					func(datum *reported.State) {
						datum.State = pointer.FromString("stress")
						datum.StateOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/stateOther"),
				),
				Entry("multiple errors",
					func(datum *reported.State) {
						datum.Severity = pointer.FromInt(-1)
						datum.State = pointer.FromString("invalid")
						datum.StateOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 10), "/severity"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"alcohol", "cycle", "hyperglycemiaSymptoms", "hypoglycemiaSymptoms", "illness", "other", "stress"}), "/state"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/stateOther"),
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
				Entry("does not modify the datum; severity missing",
					func(datum *reported.State) { datum.Severity = nil },
				),
				Entry("does not modify the datum; state missing",
					func(datum *reported.State) { datum.State = nil },
				),
				Entry("does not modify the datum; state alcohol",
					func(datum *reported.State) { datum.State = pointer.FromString("alcohol") },
				),
				Entry("does not modify the datum; state cycle",
					func(datum *reported.State) { datum.State = pointer.FromString("cycle") },
				),
				Entry("does not modify the datum; state hyperglycemiaSymptoms",
					func(datum *reported.State) { datum.State = pointer.FromString("hyperglycemiaSymptoms") },
				),
				Entry("does not modify the datum; state hypoglycemiaSymptoms",
					func(datum *reported.State) { datum.State = pointer.FromString("hypoglycemiaSymptoms") },
				),
				Entry("does not modify the datum; state illness",
					func(datum *reported.State) { datum.State = pointer.FromString("illness") },
				),
				Entry("does not modify the datum; state other",
					func(datum *reported.State) { datum.State = pointer.FromString("other") },
				),
				Entry("does not modify the datum; state stress",
					func(datum *reported.State) { datum.State = pointer.FromString("stress") },
				),
				Entry("does not modify the datum; state other missing",
					func(datum *reported.State) { datum.StateOther = nil },
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
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *reported.StateArray) {},
				),
				Entry("empty",
					func(datum *reported.StateArray) { *datum = *NewStateArray() },
				),
				Entry("nil",
					func(datum *reported.StateArray) { *datum = *NewStateArray(nil) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
				),
				Entry("single invalid",
					func(datum *reported.StateArray) { *datum = *NewStateArray(NewState("invalid")) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"alcohol", "cycle", "hyperglycemiaSymptoms", "hypoglycemiaSymptoms", "illness", "other", "stress"}), "/0/state"),
				),
				Entry("single valid",
					func(datum *reported.StateArray) { *datum = *NewStateArray(NewState("alcohol")) },
				),
				Entry("single valid with state other",
					func(datum *reported.StateArray) { *datum = *NewStateArray(NewState("other")) },
				),
				Entry("multiple invalid",
					func(datum *reported.StateArray) {
						*datum = *NewStateArray(NewState("cycle"), NewState("invalid"), NewState("alcohol"), NewState("stress"))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"alcohol", "cycle", "hyperglycemiaSymptoms", "hypoglycemiaSymptoms", "illness", "other", "stress"}), "/1/state"),
				),
				Entry("multiple valid",
					func(datum *reported.StateArray) {
						*datum = *NewStateArray(NewState("cycle"), NewState("illness"), NewState("alcohol"), NewState("stress"))
					},
				),
				Entry("multiple valid with state other",
					func(datum *reported.StateArray) {
						*datum = *NewStateArray(NewState("cycle"), NewState("illness"), NewState("other"), NewState("stress"))
					},
				),
				Entry("multiple errors",
					func(datum *reported.StateArray) {
						*datum = *NewStateArray(NewState("cycle"), nil, NewState("invalid"), NewState("stress"))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"alcohol", "cycle", "hyperglycemiaSymptoms", "hypoglycemiaSymptoms", "illness", "other", "stress"}), "/2/state"),
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
				Entry("does not modify the datum; single valid with state other",
					func(datum *reported.StateArray) { *datum = *NewStateArray(NewState("other")) },
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
				Entry("does not modify the datum; multiple valid with state other",
					func(datum *reported.StateArray) {
						*datum = *NewStateArray(NewState("cycle"), NewState("illness"), NewState("other"), NewState("stress"))
					},
				),
			)
		})
	})
})
