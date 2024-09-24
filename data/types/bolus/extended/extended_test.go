package extended_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/data/types/bolus/extended"
	dataTypesBolusExtendedTest "github.com/tidepool-org/platform/data/types/bolus/extended/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewMeta() interface{} {
	return &bolus.Meta{
		Type:    "bolus",
		SubType: "square",
	}
}

var _ = Describe("Extended", func() {
	It("SubType is expected", func() {
		Expect(extended.SubType).To(Equal("square"))
	})

	It("DurationMaximum is expected", func() {
		Expect(extended.DurationMaximum).To(Equal(86400000))
	})

	It("DurationMinimum is expected", func() {
		Expect(extended.DurationMinimum).To(Equal(0))
	})

	It("ExtendedMaximum is expected", func() {
		Expect(extended.ExtendedMaximum).To(Equal(250.0))
	})

	It("ExtendedMinimum is expected", func() {
		Expect(extended.ExtendedMinimum).To(Equal(0.0))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := extended.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("bolus"))
			Expect(datum.SubType).To(Equal("square"))
			Expect(datum.Duration).To(BeNil())
			Expect(datum.DurationExpected).To(BeNil())
			Expect(datum.Extended).To(BeNil())
			Expect(datum.ExtendedExpected).To(BeNil())
		})
	})

	Context("Extended", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *extended.Extended), expectedErrors ...error) {
					datum := dataTypesBolusExtendedTest.NewExtended()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *extended.Extended) {},
				),
				Entry("type missing",
					func(datum *extended.Extended) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &bolus.Meta{SubType: "square"}),
				),
				Entry("type invalid",
					func(datum *extended.Extended) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "bolus"), "/type", &bolus.Meta{Type: "invalidType", SubType: "square"}),
				),
				Entry("type bolus",
					func(datum *extended.Extended) { datum.Type = "bolus" },
				),
				Entry("sub type missing",
					func(datum *extended.Extended) { datum.SubType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &bolus.Meta{Type: "bolus"}),
				),
				Entry("sub type invalid",
					func(datum *extended.Extended) { datum.SubType = "invalidSubType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "square"), "/subType", &bolus.Meta{Type: "bolus", SubType: "invalidSubType"}),
				),
				Entry("sub type square",
					func(datum *extended.Extended) { datum.SubType = "square" },
				),
				Entry("duration missing; duration expected missing",
					func(datum *extended.Extended) {
						datum.Duration = nil
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (lower)",
					func(datum *extended.Extended) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, extended.DurationMaximum), "/expectedDuration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (lower)",
					func(datum *extended.Extended) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (upper)",
					func(datum *extended.Extended) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(extended.DurationMaximum)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (upper)",
					func(datum *extended.Extended) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(86400001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, extended.DurationMaximum), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected missing",
					func(datum *extended.Extended) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, extended.DurationMaximum), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (lower)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, extended.DurationMaximum), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, extended.DurationMaximum), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (lower)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, extended.DurationMaximum), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (upper)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(extended.DurationMaximum)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, extended.DurationMaximum), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (upper)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(86400001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, extended.DurationMaximum), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, extended.DurationMaximum), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected missing",
					func(datum *extended.Extended) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
				),
				Entry("duration in range (lower); duration expected out of range (lower)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, extended.DurationMaximum), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected in range (lower)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(0)
					},
				),
				Entry("duration in range (lower); duration expected in range (upper)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(extended.DurationMaximum)
					},
				),
				Entry("duration in range (lower); duration expected out of range (upper)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(86400001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, extended.DurationMaximum), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected missing",
					func(datum *extended.Extended) {
						datum.Duration = pointer.FromInt(extended.DurationMaximum)
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
				),
				Entry("duration in range (upper); duration expected out of range (lower)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.FromInt(extended.DurationMaximum)
						datum.DurationExpected = pointer.FromInt(604799999)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604799999, extended.DurationMaximum, extended.DurationMaximum), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected in range (lower)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.FromInt(extended.DurationMaximum)
						datum.DurationExpected = pointer.FromInt(extended.DurationMaximum)
					},
				),
				Entry("duration in range (upper); duration expected in range (upper)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.FromInt(extended.DurationMaximum)
						datum.DurationExpected = pointer.FromInt(extended.DurationMaximum)
					},
				),
				Entry("duration in range (upper); duration expected out of range (upper)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.FromInt(extended.DurationMaximum)
						datum.DurationExpected = pointer.FromInt(86400001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, extended.DurationMaximum, extended.DurationMaximum), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected missing",
					func(datum *extended.Extended) {
						datum.Duration = pointer.FromInt(86400001)
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, extended.DurationMaximum), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected out of range (lower)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.FromInt(86400001)
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, extended.DurationMaximum), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, extended.DurationMaximum), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected in range (lower)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.FromInt(86400001)
						datum.DurationExpected = pointer.FromInt(0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, extended.DurationMaximum), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected in range (upper)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.FromInt(86400001)
						datum.DurationExpected = pointer.FromInt(extended.DurationMaximum)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, extended.DurationMaximum), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected out of range (upper)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.FromInt(86400001)
						datum.DurationExpected = pointer.FromInt(86400001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, extended.DurationMaximum), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, extended.DurationMaximum), "/expectedDuration", NewMeta()),
				),
				Entry("extended missing; extended expected missing",
					func(datum *extended.Extended) {
						datum.DurationExpected = nil
						datum.Extended = nil
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
				),
				Entry("extended missing; extended expected out of range (lower)",
					func(datum *extended.Extended) {
						datum.Extended = nil
						datum.ExtendedExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, extended.ExtendedMinimum, extended.ExtendedMaximum), "/expectedExtended", NewMeta()),
				),
				Entry("extended missing; extended expected in range (lower)",
					func(datum *extended.Extended) {
						datum.Extended = nil
						datum.ExtendedExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
				),
				Entry("extended missing; extended expected in range (upper)",
					func(datum *extended.Extended) {
						datum.Extended = nil
						datum.ExtendedExpected = pointer.FromFloat64(extended.ExtendedMaximum)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
				),
				Entry("extended missing; extended expected out of range (upper)",
					func(datum *extended.Extended) {
						datum.Extended = nil
						datum.ExtendedExpected = pointer.FromFloat64(250.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, extended.ExtendedMinimum, extended.ExtendedMaximum), "/expectedExtended", NewMeta()),
				),
				Entry("extended out of range (lower); extended expected missing",
					func(datum *extended.Extended) {
						datum.DurationExpected = nil
						datum.Extended = pointer.FromFloat64(-0.1)
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, extended.ExtendedMinimum, extended.ExtendedMaximum), "/extended", NewMeta()),
				),
				Entry("extended out of range (lower); extended expected out of range (lower)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.FromFloat64(-0.1)
						datum.ExtendedExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, extended.ExtendedMinimum, extended.ExtendedMaximum), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, extended.ExtendedMinimum, extended.ExtendedMaximum), "/expectedExtended", NewMeta()),
				),
				Entry("extended out of range (lower); extended expected in range (lower)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.FromFloat64(-0.1)
						datum.ExtendedExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, extended.ExtendedMinimum, extended.ExtendedMaximum), "/extended", NewMeta()),
				),
				Entry("extended out of range (lower); extended expected in range (upper)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.FromFloat64(-0.1)
						datum.ExtendedExpected = pointer.FromFloat64(extended.ExtendedMaximum)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, extended.ExtendedMinimum, extended.ExtendedMaximum), "/extended", NewMeta()),
				),
				Entry("extended out of range (lower); extended expected out of range (upper)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.FromFloat64(-0.1)
						datum.ExtendedExpected = pointer.FromFloat64(250.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, extended.ExtendedMinimum, extended.ExtendedMaximum), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, extended.ExtendedMinimum, extended.ExtendedMaximum), "/expectedExtended", NewMeta()),
				),
				Entry("extended in range (lower); extended expected out of range (lower)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.FromFloat64(0.0)
						datum.ExtendedExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, extended.ExtendedMinimum, extended.ExtendedMaximum), "/expectedExtended", NewMeta()),
				),
				Entry("extended in range (lower); extended expected in range (lower)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.FromFloat64(0.0)
						datum.ExtendedExpected = pointer.FromFloat64(0.0)
					},
				),
				Entry("extended in range (lower); extended expected in range (upper)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.FromFloat64(0.0)
						datum.ExtendedExpected = pointer.FromFloat64(extended.ExtendedMaximum)
					},
				),
				Entry("extended in range (lower); extended expected out of range (upper)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.FromFloat64(0.0)
						datum.ExtendedExpected = pointer.FromFloat64(250.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, extended.ExtendedMinimum, extended.ExtendedMaximum), "/expectedExtended", NewMeta()),
				),
				Entry("extended in range (upper); extended expected missing",
					func(datum *extended.Extended) {
						datum.DurationExpected = nil
						datum.Extended = pointer.FromFloat64(extended.ExtendedMaximum)
						datum.ExtendedExpected = nil
					},
				),
				Entry("extended in range (upper); extended expected out of range (lower)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.FromFloat64(extended.ExtendedMaximum)
						datum.ExtendedExpected = pointer.FromFloat64(249.9)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(249.9, extended.ExtendedMaximum, extended.ExtendedMaximum), "/expectedExtended", NewMeta()),
				),
				Entry("extended in range (upper); extended expected in range (lower)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.FromFloat64(extended.ExtendedMaximum)
						datum.ExtendedExpected = pointer.FromFloat64(extended.ExtendedMaximum)
					},
				),
				Entry("extended in range (upper); extended expected in range (upper)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.FromFloat64(extended.ExtendedMaximum)
						datum.ExtendedExpected = pointer.FromFloat64(extended.ExtendedMaximum)
					},
				),
				Entry("extended in range (upper); extended expected out of range (upper)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.FromFloat64(extended.ExtendedMaximum)
						datum.ExtendedExpected = pointer.FromFloat64(250.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, extended.ExtendedMaximum, extended.ExtendedMaximum), "/expectedExtended", NewMeta()),
				),
				Entry("extended out of range (upper); extended expected missing",
					func(datum *extended.Extended) {
						datum.DurationExpected = nil
						datum.Extended = pointer.FromFloat64(250.1)
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, extended.ExtendedMinimum, extended.ExtendedMaximum), "/extended", NewMeta()),
				),
				Entry("extended out of range (upper); extended expected out of range (lower)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.FromFloat64(250.1)
						datum.ExtendedExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, extended.ExtendedMinimum, extended.ExtendedMaximum), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, extended.ExtendedMinimum, extended.ExtendedMaximum), "/expectedExtended", NewMeta()),
				),
				Entry("extended out of range (upper); extended expected in range (lower)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.FromFloat64(250.1)
						datum.ExtendedExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, extended.ExtendedMinimum, extended.ExtendedMaximum), "/extended", NewMeta()),
				),
				Entry("extended out of range (upper); extended expected in range (upper)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.FromFloat64(250.1)
						datum.ExtendedExpected = pointer.FromFloat64(extended.ExtendedMaximum)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, extended.ExtendedMinimum, extended.ExtendedMaximum), "/extended", NewMeta()),
				),
				Entry("extended out of range (upper); extended expected out of range (upper)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.FromFloat64(250.1)
						datum.ExtendedExpected = pointer.FromFloat64(250.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, extended.ExtendedMinimum, extended.ExtendedMaximum), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, extended.ExtendedMinimum, extended.ExtendedMaximum), "/expectedExtended", NewMeta()),
				),

				Entry("duration missing; extended expected missing",
					func(datum *extended.Extended) {
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
				),
				Entry("duration missing; extended expected exists",
					func(datum *extended.Extended) {
						datum.DurationExpected = nil
						datum.ExtendedExpected = pointer.FromFloat64(extended.ExtendedMaximum)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/expectedDuration", NewMeta()),
				),
				Entry("duration exists; extended expected missing",
					func(datum *extended.Extended) {
						datum.DurationExpected = pointer.FromInt(extended.DurationMaximum)
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/expectedDuration", NewMeta()),
				),
				Entry("duration exists; extended expected exists",
					func(datum *extended.Extended) {
						datum.DurationExpected = pointer.FromInt(extended.DurationMaximum)
						datum.ExtendedExpected = pointer.FromFloat64(extended.ExtendedMaximum)
					},
				),
				Entry("multiple errors",
					func(datum *extended.Extended) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(86400001)
						datum.Extended = nil
						datum.ExtendedExpected = pointer.FromFloat64(250.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "bolus"), "/type", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "square"), "/subType", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, extended.DurationMaximum), "/expectedDuration", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, extended.ExtendedMinimum, extended.ExtendedMaximum), "/expectedExtended", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *extended.Extended)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesBolusExtendedTest.NewExtended()
						mutator(datum)
						expectedDatum := dataTypesBolusExtendedTest.CloneExtended(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *extended.Extended) {},
				),
				Entry("does not modify the datum; type missing",
					func(datum *extended.Extended) { datum.Type = "" },
				),
				Entry("does not modify the datum; sub type missing",
					func(datum *extended.Extended) { datum.SubType = "" },
				),
				Entry("does not modify the datum; duration missing",
					func(datum *extended.Extended) { datum.Duration = nil },
				),
				Entry("does not modify the datum; duration expected missing",
					func(datum *extended.Extended) { datum.DurationExpected = nil },
				),
				Entry("does not modify the datum; extended missing",
					func(datum *extended.Extended) { datum.Extended = nil },
				),
				Entry("does not modify the datum; extended expected missing",
					func(datum *extended.Extended) { datum.ExtendedExpected = nil },
				),
			)
		})
	})
})
