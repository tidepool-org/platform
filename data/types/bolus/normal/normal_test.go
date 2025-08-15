package normal_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypesBolus "github.com/tidepool-org/platform/data/types/bolus"
	dataTypesBolusNormal "github.com/tidepool-org/platform/data/types/bolus/normal"
	dataTypesBolusNormalTest "github.com/tidepool-org/platform/data/types/bolus/normal/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() any {
	return &dataTypesBolus.Meta{
		Type:    dataTypesBolus.Type,
		SubType: dataTypesBolusNormal.SubType,
	}
}

var _ = Describe("Normal", func() {
	It("SubType is expected", func() {
		Expect(dataTypesBolusNormal.SubType).To(Equal("normal"))
	})

	It("NormalMaximum is expected", func() {
		Expect(dataTypesBolusNormal.NormalMaximum).To(Equal(100.0))
	})

	It("NormalMinimum is expected", func() {
		Expect(dataTypesBolusNormal.NormalMinimum).To(Equal(0.0))
	})

	Context("Normal", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesBolusNormal.Normal)) {
				datum := dataTypesBolusNormalTest.RandomNormal()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesBolusNormalTest.NewObjectFromNormal(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesBolusNormalTest.NewObjectFromNormal(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesBolusNormal.Normal) {},
			),
			Entry("empty",
				func(datum *dataTypesBolusNormal.Normal) {
					*datum = *dataTypesBolusNormal.New()
				},
			),
			Entry("all",
				func(datum *dataTypesBolusNormal.Normal) {
					datum.Normal = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesBolusNormal.NormalMinimum, dataTypesBolusNormal.NormalMaximum))
					datum.NormalExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Normal, dataTypesBolusNormal.NormalMaximum))
				},
			),
		)

		Context("New", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesBolusNormal.New()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Type).To(Equal(dataTypesBolus.Type))
				Expect(datum.SubType).To(Equal(dataTypesBolusNormal.SubType))
				Expect(datum.Normal).To(BeNil())
				Expect(datum.NormalExpected).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *dataTypesBolusNormal.Normal), expectedErrors ...error) {
					expectedDatum := dataTypesBolusNormalTest.RandomNormalForParser()
					object := dataTypesBolusNormalTest.NewObjectFromNormal(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesBolusNormal.New()
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *dataTypesBolusNormal.Normal) {},
				),
				Entry("multiple errors",
					func(object map[string]any, expectedDatum *dataTypesBolusNormal.Normal) {
						object["normal"] = true
						object["expectedNormal"] = true
						expectedDatum.Normal = nil
						expectedDatum.NormalExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotFloat64(true), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotFloat64(true), "/expectedNormal", NewMeta()),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesBolusNormal.Normal), expectedErrors ...error) {
					datum := dataTypesBolusNormalTest.RandomNormal()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesBolusNormal.Normal) {},
				),
				Entry("type missing",
					func(datum *dataTypesBolusNormal.Normal) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &dataTypesBolus.Meta{SubType: dataTypesBolusNormal.SubType}),
				),
				Entry("type invalid",
					func(datum *dataTypesBolusNormal.Normal) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", dataTypesBolus.Type), "/type", &dataTypesBolus.Meta{Type: "invalidType", SubType: dataTypesBolusNormal.SubType}),
				),
				Entry("type bolus",
					func(datum *dataTypesBolusNormal.Normal) { datum.Type = dataTypesBolus.Type },
				),
				Entry("sub type missing",
					func(datum *dataTypesBolusNormal.Normal) { datum.SubType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &dataTypesBolus.Meta{Type: dataTypesBolus.Type}),
				),
				Entry("sub type invalid",
					func(datum *dataTypesBolusNormal.Normal) { datum.SubType = "invalidSubType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", dataTypesBolusNormal.SubType), "/subType", &dataTypesBolus.Meta{Type: dataTypesBolus.Type, SubType: "invalidSubType"}),
				),
				Entry("sub type normal",
					func(datum *dataTypesBolusNormal.Normal) { datum.SubType = dataTypesBolusNormal.SubType },
				),
				Entry("normal missing; normal expected missing",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = nil
						datum.NormalExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
				),
				Entry("normal missing; normal expected out of range (lower)",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal missing; normal expected in range (lower)",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
				),
				Entry("normal missing; normal expected in range (upper)",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(100.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
				),
				Entry("normal missing; normal expected out of range (upper)",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected missing",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected out of range (lower)",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected in range (lower)",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 0.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected in range (upper)",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = pointer.FromFloat64(100.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected out of range (upper)",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (lower); normal expected missing",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = pointer.FromFloat64(0.0)
						datum.NormalExpected = nil
					},
				),
				Entry("normal in range (lower); normal expected out of range (lower)",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = pointer.FromFloat64(0.0)
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (lower); normal expected in range (lower)",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = pointer.FromFloat64(0.0)
						datum.NormalExpected = pointer.FromFloat64(0.0)
					},
				),
				Entry("normal in range (lower); normal expected in range (upper)",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = pointer.FromFloat64(0.0)
						datum.NormalExpected = pointer.FromFloat64(100.0)
					},
				),
				Entry("normal in range (lower); normal expected out of range (upper)",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = pointer.FromFloat64(0.0)
						datum.NormalExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (upper); normal expected missing",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = pointer.FromFloat64(100.0)
						datum.NormalExpected = nil
					},
				),
				Entry("normal in range (upper); normal expected out of range (lower)",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = pointer.FromFloat64(100.0)
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (upper); normal expected in range (lower)",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = pointer.FromFloat64(100.0)
						datum.NormalExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.0, 0.0, 0.0), "/normal", NewMeta()),
				),
				Entry("normal in range (upper); normal expected in range",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = pointer.FromFloat64(100.0)
						datum.NormalExpected = pointer.FromFloat64(99.9)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.0, 0.0, 99.9), "/normal", NewMeta()),
				),
				Entry("normal in range (upper); normal expected in range (upper)",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = pointer.FromFloat64(100.0)
						datum.NormalExpected = pointer.FromFloat64(100.0)
					},
				),
				Entry("normal in range (upper); normal expected out of range (upper)",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = pointer.FromFloat64(100.0)
						datum.NormalExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected missing",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = pointer.FromFloat64(100.1)
						datum.NormalExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected out of range (lower)",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = pointer.FromFloat64(100.1)
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected in range (lower)",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = pointer.FromFloat64(100.1)
						datum.NormalExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 0.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected in range (upper)",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = pointer.FromFloat64(100.1)
						datum.NormalExpected = pointer.FromFloat64(100.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected out of range (upper)",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Normal = pointer.FromFloat64(100.1)
						datum.NormalExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("multiple errors",
					func(datum *dataTypesBolusNormal.Normal) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", dataTypesBolus.Type), "/type", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", dataTypesBolusNormal.SubType), "/subType", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *dataTypesBolusNormal.Normal)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesBolusNormalTest.RandomNormal()
						mutator(datum)
						expectedDatum := dataTypesBolusNormalTest.CloneNormal(datum)
						normalizer := dataNormalizer.New(logTest.NewLogger())
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *dataTypesBolusNormal.Normal) {},
				),
				Entry("does not modify the datum; type missing",
					func(datum *dataTypesBolusNormal.Normal) { datum.Type = "" },
				),
				Entry("does not modify the datum; sub type missing",
					func(datum *dataTypesBolusNormal.Normal) { datum.SubType = "" },
				),
				Entry("does not modify the datum; normal missing",
					func(datum *dataTypesBolusNormal.Normal) { datum.Normal = nil },
				),
				Entry("does not modify the datum; normal expected missing",
					func(datum *dataTypesBolusNormal.Normal) { datum.NormalExpected = nil },
				),
			)
		})
	})
})
