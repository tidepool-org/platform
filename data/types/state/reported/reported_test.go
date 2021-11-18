package reported_test

import (
	"math/rand"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/state/reported"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &types.Meta{
		Type: "reportedState",
	}
}

func NewReported() *reported.Reported {
	datum := reported.New()
	datum.Base = *dataTypesTest.RandomBase()
	datum.Type = "reportedState"
	datum.States = NewStateArray()
	for index := rand.Intn(len(reported.StateStates())); index >= 0; index-- {
		*datum.States = append(*datum.States, NewState(test.RandomStringFromArray(reported.StateStates())))
	}
	return datum
}

func CloneReported(datum *reported.Reported) *reported.Reported {
	if datum == nil {
		return nil
	}
	clone := reported.New()
	clone.Base = *dataTypesTest.CloneBase(&datum.Base)
	clone.States = CloneStateArray(datum.States)
	return clone
}

var _ = Describe("Reported", func() {
	It("Type is expected", func() {
		Expect(reported.Type).To(Equal("reportedState"))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := reported.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("reportedState"))
			Expect(datum.States).To(BeNil())
		})
	})

	Context("Reported", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *reported.Reported), expectedErrors ...error) {
					datum := NewReported()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *reported.Reported) {},
				),
				Entry("type missing",
					func(datum *reported.Reported) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
				),
				Entry("type invalid",
					func(datum *reported.Reported) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "reportedState"), "/type", &types.Meta{Type: "invalidType"}),
				),
				Entry("type reportedState",
					func(datum *reported.Reported) { datum.Type = "reportedState" },
				),
				Entry("states missing",
					func(datum *reported.Reported) { datum.States = nil },
				),
				Entry("states empty",
					func(datum *reported.Reported) { datum.States = NewStateArray() },
				),
				Entry("states single invalid",
					func(datum *reported.Reported) { datum.States = NewStateArray(NewState("invalidState")) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalidState", []string{"alcohol", "cycle", "hyperglycemiaSymptoms", "hypoglycemiaSymptoms", "illness", "other", "stress"}), "/states/0/state", NewMeta()),
				),
				Entry("states single valid",
					func(datum *reported.Reported) { datum.States = NewStateArray(NewState("alcohol")) },
				),
				Entry("states single valid with state other",
					func(datum *reported.Reported) { datum.States = NewStateArray(NewState("other")) },
				),
				Entry("states multiple invalid single",
					func(datum *reported.Reported) {
						datum.States = NewStateArray(NewState("alcohol"), NewState("invalidState"), NewState("hyperglycemiaSymptoms"))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalidState", []string{"alcohol", "cycle", "hyperglycemiaSymptoms", "hypoglycemiaSymptoms", "illness", "other", "stress"}), "/states/1/state", NewMeta()),
				),
				Entry("states multiple invalid multiple",
					func(datum *reported.Reported) {
						datum.States = NewStateArray(NewState("invalidStateOne"), NewState("cycle"), NewState("invalidStateTwo"))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalidStateOne", []string{"alcohol", "cycle", "hyperglycemiaSymptoms", "hypoglycemiaSymptoms", "illness", "other", "stress"}), "/states/0/state", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalidStateTwo", []string{"alcohol", "cycle", "hyperglycemiaSymptoms", "hypoglycemiaSymptoms", "illness", "other", "stress"}), "/states/2/state", NewMeta()),
				),
				Entry("states multiple valid",
					func(datum *reported.Reported) {
						datum.States = NewStateArray(NewState("alcohol"), NewState("cycle"), NewState("hyperglycemiaSymptoms"))
					},
				),
				Entry("states multiple valid with state other",
					func(datum *reported.Reported) {
						datum.States = NewStateArray(NewState("alcohol"), NewState("other"), NewState("hyperglycemiaSymptoms"))
					},
				),
				Entry("states multiple valid repeats",
					func(datum *reported.Reported) { datum.States = NewStateArray(NewState("alcohol"), NewState("alcohol")) },
				),
				Entry("multiple errors",
					func(datum *reported.Reported) {
						datum.Type = "invalidType"
						datum.States = NewStateArray(NewState("invalidState"))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "reportedState"), "/type", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalidState", []string{"alcohol", "cycle", "hyperglycemiaSymptoms", "hypoglycemiaSymptoms", "illness", "other", "stress"}), "/states/0/state", &types.Meta{Type: "invalidType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *reported.Reported)) {
					for _, origin := range structure.Origins() {
						datum := NewReported()
						mutator(datum)
						expectedDatum := CloneReported(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *reported.Reported) {},
				),
				Entry("does not modify the datum; states empty",
					func(datum *reported.Reported) { datum.States = NewStateArray() },
				),
				Entry("does not modify the datum; states single",
					func(datum *reported.Reported) { datum.States = NewStateArray(NewState("alcohol")) },
				),
				Entry("does not modify the datum; states multiple",
					func(datum *reported.Reported) {
						datum.States = NewStateArray(NewState("alcohol"), NewState("other"), NewState("hyperglycemiaSymptoms"))
					},
				),
			)
		})
	})
})
