package normal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/data/types/bolus/normal"
	dataTypesBolusNormalTest "github.com/tidepool-org/platform/data/types/bolus/normal/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewMeta() interface{} {
	return &bolus.Meta{
		Type:    "bolus",
		SubType: "normal",
	}
}

var _ = Describe("Normal", func() {
	It("SubType is expected", func() {
		Expect(normal.SubType).To(Equal("normal"))
	})

	It("NormalMaximum is expected", func() {
		Expect(normal.NormalMaximum).To(Equal(100.0))
	})

	It("NormalMinimum is expected", func() {
		Expect(normal.NormalMinimum).To(Equal(0.0))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := normal.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("bolus"))
			Expect(datum.SubType).To(Equal("normal"))
			Expect(datum.Normal).To(BeNil())
			Expect(datum.NormalExpected).To(BeNil())
		})
	})

	Context("Normal", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *normal.Normal), expectedErrors ...error) {
					datum := dataTypesBolusNormalTest.NewNormal()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *normal.Normal) {},
				),
				Entry("type missing",
					func(datum *normal.Normal) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &bolus.Meta{SubType: "normal"}),
				),
				Entry("type invalid",
					func(datum *normal.Normal) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "bolus"), "/type", &bolus.Meta{Type: "invalidType", SubType: "normal"}),
				),
				Entry("type bolus",
					func(datum *normal.Normal) { datum.Type = "bolus" },
				),
				Entry("sub type missing",
					func(datum *normal.Normal) { datum.SubType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &bolus.Meta{Type: "bolus"}),
				),
				Entry("sub type invalid",
					func(datum *normal.Normal) { datum.SubType = "invalidSubType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "normal"), "/subType", &bolus.Meta{Type: "bolus", SubType: "invalidSubType"}),
				),
				Entry("sub type normal",
					func(datum *normal.Normal) { datum.SubType = "normal" },
				),
				Entry("normal missing; normal expected missing",
					func(datum *normal.Normal) {
						datum.Normal = nil
						datum.NormalExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
				),
				Entry("normal missing; normal expected out of range (lower)",
					func(datum *normal.Normal) {
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal missing; normal expected in range (lower)",
					func(datum *normal.Normal) {
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
				),
				Entry("normal missing; normal expected in range (upper)",
					func(datum *normal.Normal) {
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(100.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
				),
				Entry("normal missing; normal expected out of range (upper)",
					func(datum *normal.Normal) {
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected missing",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected out of range (lower)",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected in range (lower)",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected in range (upper)",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = pointer.FromFloat64(100.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected out of range (upper)",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				// Entry("normal in range (lower); normal expected missing",
				// 	func(datum *normal.Normal) {
				// 		datum.Normal = pointer.FromFloat64(0.0)
				// 		datum.NormalExpected = nil
				// 	},
				// 	errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/expectedNormal", NewMeta()),
				// ),
				Entry("normal in range (lower); normal expected out of range (lower)",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(0.0)
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (lower); normal expected in range (lower)",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(0.0)
						datum.NormalExpected = pointer.FromFloat64(0.0)
					},
				),
				Entry("normal in range (lower); normal expected in range (upper)",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(0.0)
						datum.NormalExpected = pointer.FromFloat64(100.0)
					},
				),
				Entry("normal in range (lower); normal expected out of range (upper)",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(0.0)
						datum.NormalExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (upper); normal expected missing",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(100.0)
						datum.NormalExpected = nil
					},
				),
				// Entry("normal in range (upper); normal expected out of range (lower)",
				// 	func(datum *normal.Normal) {
				// 		datum.Normal = pointer.FromFloat64(100.0)
				// 		datum.NormalExpected = pointer.FromFloat64(99.9)
				// 	},
				// 	errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(99.9, 100.0, 100.0), "/expectedNormal", NewMeta()),
				// ),
				Entry("normal in range (upper); normal expected in range (lower)",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(100.0)
						datum.NormalExpected = pointer.FromFloat64(100.0)
					},
				),
				Entry("normal in range (upper); normal expected in range (upper)",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(100.0)
						datum.NormalExpected = pointer.FromFloat64(100.0)
					},
				),
				// Entry("normal in range (upper); normal expected out of range (upper)",
				// 	func(datum *normal.Normal) {
				// 		datum.Normal = pointer.FromFloat64(100.0)
				// 		datum.NormalExpected = pointer.FromFloat64(100.1)
				// 	},
				// 	errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 100.0, 100.0), "/expectedNormal", NewMeta()),
				// ),
				Entry("normal out of range (upper); normal expected missing",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(100.1)
						datum.NormalExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected out of range (lower)",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(100.1)
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected in range (lower)",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(100.1)
						datum.NormalExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected in range (upper)",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(100.1)
						datum.NormalExpected = pointer.FromFloat64(100.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected out of range (upper)",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(100.1)
						datum.NormalExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("multiple errors",
					func(datum *normal.Normal) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "bolus"), "/type", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "normal"), "/subType", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *normal.Normal)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesBolusNormalTest.NewNormal()
						mutator(datum)
						expectedDatum := dataTypesBolusNormalTest.CloneNormal(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *normal.Normal) {},
				),
				Entry("does not modify the datum; type missing",
					func(datum *normal.Normal) { datum.Type = "" },
				),
				Entry("does not modify the datum; sub type missing",
					func(datum *normal.Normal) { datum.SubType = "" },
				),
				Entry("does not modify the datum; normal missing",
					func(datum *normal.Normal) { datum.Normal = nil },
				),
				Entry("does not modify the datum; normal expected missing",
					func(datum *normal.Normal) { datum.NormalExpected = nil },
				),
			)
		})
	})
})
