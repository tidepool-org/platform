package combination_test

import (
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/data/types/bolus/combination"
	dataTypesBolusCombinationTest "github.com/tidepool-org/platform/data/types/bolus/combination/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func NewMeta() interface{} {
	return &bolus.Meta{
		Type:    "bolus",
		SubType: "dual/square",
	}
}

var _ = Describe("Combination", func() {
	It("SubType is expected", func() {
		Expect(combination.SubType).To(Equal("dual/square"))
	})

	It("DurationMaximum is expected", func() {
		Expect(combination.DurationMaximum).To(Equal(86400000))
	})

	It("DurationMinimum is expected", func() {
		Expect(combination.DurationMinimum).To(Equal(0))
	})

	It("ExtendedMaximum is expected", func() {
		Expect(combination.ExtendedMaximum).To(Equal(250.0))
	})

	It("ExtendedMinimum is expected", func() {
		Expect(combination.ExtendedMinimum).To(Equal(0.0))
	})

	It("NormalMaximum is expected", func() {
		Expect(combination.NormalMaximum).To(Equal(250.0))
	})

	It("NormalMinimum is expected", func() {
		Expect(combination.NormalMinimum).To(Equal(0.0))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := combination.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("bolus"))
			Expect(datum.SubType).To(Equal("dual/square"))
			Expect(datum.Duration).To(BeNil())
			Expect(datum.DurationExpected).To(BeNil())
			Expect(datum.Extended).To(BeNil())
			Expect(datum.ExtendedExpected).To(BeNil())
			Expect(datum.Normal).To(BeNil())
			Expect(datum.NormalExpected).To(BeNil())
		})
	})

	Context("Combination", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *combination.Combination), expectedErrors ...error) {
					datum := dataTypesBolusCombinationTest.NewCombination()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *combination.Combination) {},
				),
				Entry("type missing",
					func(datum *combination.Combination) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &bolus.Meta{SubType: combination.SubType}),
				),
				Entry("type invalid",
					func(datum *combination.Combination) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "bolus"), "/type", &bolus.Meta{Type: "invalidType", SubType: combination.SubType}),
				),
				Entry("type bolus",
					func(datum *combination.Combination) { datum.Type = "bolus" },
				),
				Entry("sub type missing",
					func(datum *combination.Combination) { datum.SubType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &bolus.Meta{Type: "bolus"}),
				),
				Entry("sub type invalid",
					func(datum *combination.Combination) { datum.SubType = "invalidSubType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", combination.SubType), "/subType", &bolus.Meta{Type: "bolus", SubType: "invalidSubType"}),
				),
				Entry("sub type dual/square",
					func(datum *combination.Combination) { datum.SubType = combination.SubType },
				),
				Entry("normal missing; normal expected missing",
					func(datum *combination.Combination) {
						datum.Normal = nil
						datum.NormalExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
				),
				Entry("normal missing; normal expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, combination.NormalMinimum, combination.NormalMaximum), "/expectedNormal", NewMeta()),
				),
				Entry("normal missing; normal expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
				),
				Entry("normal missing; normal expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
				),
				Entry("normal missing; normal expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(250.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, combination.NormalMinimum, combination.NormalMaximum), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected missing",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, combination.NormalMinimum, combination.NormalMaximum), "/normal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, combination.NormalMinimum, combination.NormalMaximum), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, combination.NormalMinimum, combination.NormalMaximum), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, combination.NormalMinimum, combination.NormalMaximum), "/normal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, combination.NormalMinimum, combination.NormalMaximum), "/normal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = pointer.FromFloat64(250.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, combination.NormalMinimum, combination.NormalMaximum), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, combination.NormalMinimum, combination.NormalMaximum), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (lower); normal expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(0.0)
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, combination.NormalMinimum, combination.NormalMaximum), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (lower); normal expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(0.0)
						datum.NormalExpected = pointer.FromFloat64(0.0)
					},
				),
				Entry("normal in range (lower); normal expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(0.0)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
					},
				),
				Entry("normal in range (lower); normal expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(0.0)
						datum.NormalExpected = pointer.FromFloat64(250.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, combination.NormalMinimum, combination.NormalMaximum), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (upper); normal expected missing",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = nil
					},
				),
				Entry("normal in range (upper); normal expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(249.9)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(249.9, combination.NormalMaximum, combination.NormalMaximum), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (upper); normal expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
					},
				),
				Entry("normal in range (upper); normal expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
					},
				),
				Entry("normal in range (upper); normal expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(250.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, combination.NormalMaximum, combination.NormalMaximum), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected missing",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(250.1)
						datum.NormalExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, combination.NormalMinimum, combination.NormalMaximum), "/normal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(250.1)
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, combination.NormalMinimum, combination.NormalMaximum), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, combination.NormalMinimum, combination.NormalMaximum), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(250.1)
						datum.NormalExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, combination.NormalMinimum, combination.NormalMaximum), "/normal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(250.1)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, combination.NormalMinimum, combination.NormalMaximum), "/normal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(250.1)
						datum.NormalExpected = pointer.FromFloat64(250.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, combination.NormalMinimum, combination.NormalMaximum), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, combination.NormalMinimum, combination.NormalMaximum), "/expectedNormal", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, combination.DurationMaximum), "/expectedDuration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(combination.DurationMaximum)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(86400001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, combination.DurationMaximum), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected missing",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, combination.DurationMaximum), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, combination.DurationMaximum), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, combination.DurationMaximum), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, combination.DurationMaximum), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(combination.DurationMaximum)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, combination.DurationMaximum), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(86400001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, combination.DurationMaximum), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, combination.DurationMaximum), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected missing",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
				),
				Entry("duration in range (lower); duration expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, combination.DurationMaximum), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(0)
					},
				),
				Entry("duration in range (lower); duration expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(combination.DurationMaximum)
					},
				),
				Entry("duration in range (lower); duration expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(86400001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, combination.DurationMaximum), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected missing",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Duration = pointer.FromInt(combination.DurationMaximum)
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
				),
				Entry("duration in range (upper); duration expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Duration = pointer.FromInt(combination.DurationMaximum)
						datum.DurationExpected = pointer.FromInt(604799999)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604799999, combination.DurationMaximum, combination.DurationMaximum), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Duration = pointer.FromInt(combination.DurationMaximum)
						datum.DurationExpected = pointer.FromInt(combination.DurationMaximum)
					},
				),
				Entry("duration in range (upper); duration expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Duration = pointer.FromInt(combination.DurationMaximum)
						datum.DurationExpected = pointer.FromInt(combination.DurationMaximum)
					},
				),
				Entry("duration in range (upper); duration expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Duration = pointer.FromInt(combination.DurationMaximum)
						datum.DurationExpected = pointer.FromInt(86400001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, combination.DurationMaximum, combination.DurationMaximum), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected missing",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Duration = pointer.FromInt(86400001)
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, combination.DurationMaximum), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Duration = pointer.FromInt(86400001)
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, combination.DurationMaximum), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, combination.DurationMaximum), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Duration = pointer.FromInt(86400001)
						datum.DurationExpected = pointer.FromInt(0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, combination.DurationMaximum), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Duration = pointer.FromInt(86400001)
						datum.DurationExpected = pointer.FromInt(combination.DurationMaximum)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, combination.DurationMaximum), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Duration = pointer.FromInt(86400001)
						datum.DurationExpected = pointer.FromInt(86400001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, combination.DurationMaximum), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, combination.DurationMaximum), "/expectedDuration", NewMeta()),
				),
				Entry("extended missing; extended expected missing",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.DurationExpected = nil
						datum.Extended = nil
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
				),
				Entry("extended missing; extended expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Extended = nil
						datum.ExtendedExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, combination.ExtendedMinimum, combination.ExtendedMaximum), "/expectedExtended", NewMeta()),
				),
				Entry("extended missing; extended expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Extended = nil
						datum.ExtendedExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
				),
				Entry("extended missing; extended expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Extended = nil
						datum.ExtendedExpected = pointer.FromFloat64(combination.ExtendedMaximum)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
				),
				Entry("extended missing; extended expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Extended = nil
						datum.ExtendedExpected = pointer.FromFloat64(250.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, combination.ExtendedMinimum, combination.ExtendedMaximum), "/expectedExtended", NewMeta()),
				),
				Entry("extended out of range (lower); extended expected missing",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.DurationExpected = nil
						datum.Extended = pointer.FromFloat64(-0.1)
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, combination.ExtendedMinimum, combination.ExtendedMaximum), "/extended", NewMeta()),
				),
				Entry("extended out of range (lower); extended expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Extended = pointer.FromFloat64(-0.1)
						datum.ExtendedExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, combination.ExtendedMinimum, combination.ExtendedMaximum), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, combination.ExtendedMinimum, combination.ExtendedMaximum), "/expectedExtended", NewMeta()),
				),
				Entry("extended out of range (lower); extended expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Extended = pointer.FromFloat64(-0.1)
						datum.ExtendedExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, combination.ExtendedMinimum, combination.ExtendedMaximum), "/extended", NewMeta()),
				),
				Entry("extended out of range (lower); extended expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Extended = pointer.FromFloat64(-0.1)
						datum.ExtendedExpected = pointer.FromFloat64(combination.ExtendedMaximum)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, combination.ExtendedMinimum, combination.ExtendedMaximum), "/extended", NewMeta()),
				),
				Entry("extended out of range (lower); extended expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Extended = pointer.FromFloat64(-0.1)
						datum.ExtendedExpected = pointer.FromFloat64(250.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, combination.ExtendedMinimum, combination.ExtendedMaximum), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, combination.ExtendedMinimum, combination.ExtendedMaximum), "/expectedExtended", NewMeta()),
				),
				Entry("extended in range (lower); extended expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Extended = pointer.FromFloat64(0.0)
						datum.ExtendedExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, combination.ExtendedMinimum, combination.ExtendedMaximum), "/expectedExtended", NewMeta()),
				),
				Entry("extended in range (lower); extended expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Extended = pointer.FromFloat64(0.0)
						datum.ExtendedExpected = pointer.FromFloat64(0.0)
					},
				),
				Entry("extended in range (lower); extended expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Extended = pointer.FromFloat64(0.0)
						datum.ExtendedExpected = pointer.FromFloat64(combination.ExtendedMaximum)
					},
				),
				Entry("extended in range (lower); extended expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Extended = pointer.FromFloat64(0.0)
						datum.ExtendedExpected = pointer.FromFloat64(250.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, combination.ExtendedMinimum, combination.ExtendedMaximum), "/expectedExtended", NewMeta()),
				),
				Entry("extended in range (upper); extended expected missing",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.DurationExpected = nil
						datum.Extended = pointer.FromFloat64(combination.ExtendedMaximum)
						datum.ExtendedExpected = nil
					},
				),
				Entry("extended in range (upper); extended expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Extended = pointer.FromFloat64(combination.ExtendedMaximum)
						datum.ExtendedExpected = pointer.FromFloat64(249.9)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(249.9, combination.ExtendedMaximum, combination.ExtendedMaximum), "/expectedExtended", NewMeta()),
				),
				Entry("extended in range (upper); extended expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Extended = pointer.FromFloat64(combination.ExtendedMaximum)
						datum.ExtendedExpected = pointer.FromFloat64(combination.ExtendedMaximum)
					},
				),
				Entry("extended in range (upper); extended expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Extended = pointer.FromFloat64(combination.ExtendedMaximum)
						datum.ExtendedExpected = pointer.FromFloat64(combination.ExtendedMaximum)
					},
				),
				Entry("extended in range (upper); extended expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Extended = pointer.FromFloat64(combination.ExtendedMaximum)
						datum.ExtendedExpected = pointer.FromFloat64(250.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, combination.ExtendedMaximum, combination.ExtendedMaximum), "/expectedExtended", NewMeta()),
				),
				Entry("extended out of range (upper); extended expected missing",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.DurationExpected = nil
						datum.Extended = pointer.FromFloat64(250.1)
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, combination.ExtendedMinimum, combination.ExtendedMaximum), "/extended", NewMeta()),
				),
				Entry("extended out of range (upper); extended expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Extended = pointer.FromFloat64(250.1)
						datum.ExtendedExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, combination.ExtendedMinimum, combination.ExtendedMaximum), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, combination.ExtendedMinimum, combination.ExtendedMaximum), "/expectedExtended", NewMeta()),
				),
				Entry("extended out of range (upper); extended expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Extended = pointer.FromFloat64(250.1)
						datum.ExtendedExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, combination.ExtendedMinimum, combination.ExtendedMaximum), "/extended", NewMeta()),
				),
				Entry("extended out of range (upper); extended expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Extended = pointer.FromFloat64(250.1)
						datum.ExtendedExpected = pointer.FromFloat64(combination.ExtendedMaximum)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, combination.ExtendedMinimum, combination.ExtendedMaximum), "/extended", NewMeta()),
				),
				Entry("extended out of range (upper); extended expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.Extended = pointer.FromFloat64(250.1)
						datum.ExtendedExpected = pointer.FromFloat64(250.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, combination.ExtendedMinimum, combination.ExtendedMaximum), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, combination.ExtendedMinimum, combination.ExtendedMaximum), "/expectedExtended", NewMeta()),
				),
				Entry("duration missing; extended expected missing",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
				),
				Entry("duration missing; extended expected exists",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.DurationExpected = nil
						datum.ExtendedExpected = pointer.FromFloat64(combination.ExtendedMaximum)
					},
				),
				Entry("duration exists; extended expected missing",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.DurationExpected = pointer.FromInt(combination.DurationMaximum)
						datum.ExtendedExpected = nil
					},
				),
				Entry("duration exists; extended expected exists",
					func(datum *combination.Combination) {
						datum.Normal = pointer.FromFloat64(combination.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(combination.NormalMaximum)
						datum.DurationExpected = pointer.FromInt(combination.DurationMaximum)
						datum.ExtendedExpected = pointer.FromFloat64(combination.ExtendedMaximum)
					},
				),
				Entry("multiple errors",
					func(datum *combination.Combination) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(86400001)
						datum.Extended = nil
						datum.ExtendedExpected = pointer.FromFloat64(250.1)
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(250.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "bolus"), "/type", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", combination.SubType), "/subType", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, combination.DurationMaximum), "/expectedDuration", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, combination.ExtendedMinimum, combination.ExtendedMaximum), "/expectedExtended", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(250.1, combination.NormalMinimum, combination.NormalMaximum), "/expectedNormal", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *combination.Combination)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesBolusCombinationTest.NewCombination()
						mutator(datum)
						expectedDatum := dataTypesBolusCombinationTest.CloneCombination(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *combination.Combination) {},
				),
				Entry("does not modify the datum; type missing",
					func(datum *combination.Combination) { datum.Type = "" },
				),
				Entry("does not modify the datum; sub type missing",
					func(datum *combination.Combination) { datum.SubType = "" },
				),
				Entry("does not modify the datum; duration missing",
					func(datum *combination.Combination) { datum.Duration = nil },
				),
				Entry("does not modify the datum; duration expected missing",
					func(datum *combination.Combination) { datum.DurationExpected = nil },
				),
				Entry("does not modify the datum; extended missing",
					func(datum *combination.Combination) { datum.Extended = nil },
				),
				Entry("does not modify the datum; extended expected missing",
					func(datum *combination.Combination) { datum.ExtendedExpected = nil },
				),
				Entry("does not modify the datum; normal missing",
					func(datum *combination.Combination) { datum.Normal = nil },
				),
				Entry("does not modify the datum; normal expected missing",
					func(datum *combination.Combination) { datum.NormalExpected = nil },
				),
			)
		})
	})
})
