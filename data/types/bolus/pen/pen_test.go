package pen_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/data/types/bolus/pen"
	dataTypesBolusPenTest "github.com/tidepool-org/platform/data/types/bolus/pen/test"
	"github.com/tidepool-org/platform/data/types/common"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewMeta() interface{} {
	return &bolus.Meta{
		Type:    "bolus",
		SubType: "pen",
	}
}

var _ = Describe("Pen", func() {
	It("SubType is expected", func() {
		Expect(pen.SubType).To(Equal("pen"))
	})

	It("NormalMaximum is expected", func() {
		Expect(pen.NormalMaximum).To(Equal(100.0))
	})

	It("NormalMinimum is expected", func() {
		Expect(pen.NormalMinimum).To(Equal(0.0))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := pen.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("bolus"))
			Expect(datum.SubType).To(Equal("pen"))
			Expect(datum.Normal).To(BeNil())
		})
	})

	Context("Pen", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pen.Pen), expectedErrors ...error) {
					datum := dataTypesBolusPenTest.NewPen()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pen.Pen) {},
				),
				Entry("type missing",
					func(datum *pen.Pen) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &bolus.Meta{SubType: "pen"}),
				),
				Entry("type invalid",
					func(datum *pen.Pen) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "bolus"), "/type", &bolus.Meta{Type: "invalidType", SubType: "pen"}),
				),
				Entry("type bolus",
					func(datum *pen.Pen) { datum.Type = "bolus" },
				),
				Entry("sub type missing",
					func(datum *pen.Pen) { datum.SubType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &bolus.Meta{Type: "bolus"}),
				),
				Entry("sub type invalid",
					func(datum *pen.Pen) { datum.SubType = "invalidSubType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "pen"), "/subType", &bolus.Meta{Type: "bolus", SubType: "invalidSubType"}),
				),
				Entry("sub type normal",
					func(datum *pen.Pen) { datum.SubType = "pen" },
				),
				Entry("normal missing",
					func(datum *pen.Pen) {
						datum.Normal = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
				),
				Entry("normal out of range (lower)",
					func(datum *pen.Pen) {
						datum.Normal = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal in range (lower)",
					func(datum *pen.Pen) {
						datum.Normal = pointer.FromFloat64(0.0)
					},
				),
				Entry("normal in range (upper)",
					func(datum *pen.Pen) {
						datum.Normal = pointer.FromFloat64(100.0)
					},
				),
				Entry("normal out of range (upper)",
					func(datum *pen.Pen) {
						datum.Normal = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (upper)",
					func(datum *pen.Pen) {
						datum.Normal = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("multiple errors",
					func(datum *pen.Pen) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
						datum.Normal = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "bolus"), "/type", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "pen"), "/subType", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pen.Pen)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesBolusPenTest.NewPen()
						mutator(datum)
						expectedDatum := dataTypesBolusPenTest.ClonePen(datum)
						if *datum.Prescriptor.Prescriptor == common.ManualPrescriptor {
							expectedDatum.InsulinOnBoard = nil
						}
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pen.Pen) {},
				),
				Entry("does not modify the datum; type missing",
					func(datum *pen.Pen) { datum.Type = "" },
				),
				Entry("does not modify the datum; sub type missing",
					func(datum *pen.Pen) { datum.SubType = "" },
				),
				Entry("does not modify the datum; normal missing",
					func(datum *pen.Pen) { datum.Normal = nil },
				),
			)
		})
	})
})
