package combination_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypesBolus "github.com/tidepool-org/platform/data/types/bolus"
	dataTypesBolusCombination "github.com/tidepool-org/platform/data/types/bolus/combination"
	dataTypesBolusCombinationTest "github.com/tidepool-org/platform/data/types/bolus/combination/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &dataTypesBolus.Meta{
		Type:    dataTypesBolus.Type,
		SubType: dataTypesBolusCombination.SubType,
	}
}

var _ = Describe("Combination", func() {
	It("SubType is expected", func() {
		Expect(dataTypesBolusCombination.SubType).To(Equal("dual/square"))
	})

	It("DurationMaximum is expected", func() {
		Expect(dataTypesBolusCombination.DurationMaximum).To(Equal(86400000))
	})

	It("DurationMinimum is expected", func() {
		Expect(dataTypesBolusCombination.DurationMinimum).To(Equal(0))
	})

	It("ExtendedMaximum is expected", func() {
		Expect(dataTypesBolusCombination.ExtendedMaximum).To(Equal(100.0))
	})

	It("ExtendedMinimum is expected", func() {
		Expect(dataTypesBolusCombination.ExtendedMinimum).To(Equal(0.0))
	})

	It("NormalMaximum is expected", func() {
		Expect(dataTypesBolusCombination.NormalMaximum).To(Equal(100.0))
	})

	It("NormalMinimum is expected", func() {
		Expect(dataTypesBolusCombination.NormalMinimum).To(Equal(0.0))
	})

	Context("Combination", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesBolusCombination.Combination)) {
				datum := dataTypesBolusCombinationTest.RandomCombination()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesBolusCombinationTest.NewObjectFromCombination(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesBolusCombinationTest.NewObjectFromCombination(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesBolusCombination.Combination) {},
			),
			Entry("empty",
				func(datum *dataTypesBolusCombination.Combination) {
					*datum = *dataTypesBolusCombination.New()
				},
			),
			Entry("all",
				func(datum *dataTypesBolusCombination.Combination) {
					datum.Duration = pointer.FromInt(test.RandomIntFromRange(dataTypesBolusCombination.DurationMinimum, dataTypesBolusCombination.DurationMaximum))
					datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusCombination.DurationMaximum))
					datum.Extended = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesBolusCombination.ExtendedMinimum, dataTypesBolusCombination.ExtendedMaximum))
					datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusCombination.ExtendedMaximum))
					datum.Normal = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesBolusCombination.NormalMinimum, dataTypesBolusCombination.NormalMaximum))
					datum.NormalExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Normal, dataTypesBolusCombination.NormalMaximum))
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *dataTypesBolusCombination.Combination), expectedErrors ...error) {
					expectedDatum := dataTypesBolusCombinationTest.RandomCombinationForParser()
					object := dataTypesBolusCombinationTest.NewObjectFromCombination(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesBolusCombination.New()
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *dataTypesBolusCombination.Combination) {},
				),
				Entry("multiple errors",
					func(object map[string]any, expectedDatum *dataTypesBolusCombination.Combination) {
						object["duration"] = true
						object["expectedDuration"] = true
						object["extended"] = true
						object["expectedExtended"] = true
						object["normal"] = true
						object["expectedNormal"] = true
						expectedDatum.Duration = nil
						expectedDatum.DurationExpected = nil
						expectedDatum.Extended = nil
						expectedDatum.ExtendedExpected = nil
						expectedDatum.Normal = nil
						expectedDatum.NormalExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotInt(true), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotInt(true), "/expectedDuration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotFloat64(true), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotFloat64(true), "/expectedExtended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotFloat64(true), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotFloat64(true), "/expectedNormal", NewMeta()),
				),
			)
		})

		Context("New", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesBolusCombination.New()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Type).To(Equal(dataTypesBolus.Type))
				Expect(datum.SubType).To(Equal(dataTypesBolusCombination.SubType))
				Expect(datum.Duration).To(BeNil())
				Expect(datum.DurationExpected).To(BeNil())
				Expect(datum.Extended).To(BeNil())
				Expect(datum.ExtendedExpected).To(BeNil())
				Expect(datum.Normal).To(BeNil())
				Expect(datum.NormalExpected).To(BeNil())
			})
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesBolusCombination.Combination), expectedErrors ...error) {
					datum := dataTypesBolusCombinationTest.RandomCombination()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesBolusCombination.Combination) {},
				),
				Entry("type missing",
					func(datum *dataTypesBolusCombination.Combination) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &dataTypesBolus.Meta{SubType: dataTypesBolusCombination.SubType}),
				),
				Entry("type invalid",
					func(datum *dataTypesBolusCombination.Combination) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", dataTypesBolus.Type), "/type", &dataTypesBolus.Meta{Type: "invalidType", SubType: dataTypesBolusCombination.SubType}),
				),
				Entry("type bolus",
					func(datum *dataTypesBolusCombination.Combination) { datum.Type = dataTypesBolus.Type },
				),
				Entry("sub type missing",
					func(datum *dataTypesBolusCombination.Combination) { datum.SubType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &dataTypesBolus.Meta{Type: dataTypesBolus.Type}),
				),
				Entry("sub type invalid",
					func(datum *dataTypesBolusCombination.Combination) { datum.SubType = "invalidSubType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", dataTypesBolusCombination.SubType), "/subType", &dataTypesBolus.Meta{Type: dataTypesBolus.Type, SubType: "invalidSubType"}),
				),
				Entry("sub type dual/square",
					func(datum *dataTypesBolusCombination.Combination) { datum.SubType = dataTypesBolusCombination.SubType },
				),
				Entry("duration missing; duration expected missing",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = nil
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(-1)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusCombination.NormalMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(0)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusCombination.NormalMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(86400000)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusCombination.NormalMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(86400001)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusCombination.NormalMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected missing",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(-1)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusCombination.NormalMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(0)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusCombination.NormalMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 0), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(86400000)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusCombination.NormalMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(86400001)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusCombination.NormalMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected missing",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
				),
				Entry("duration in range (lower); duration expected out of range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(-1)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusCombination.NormalMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected in range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(0)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusCombination.NormalMaximum))
					},
				),
				Entry("duration in range (lower); duration expected in range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(86400000)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusCombination.NormalMaximum))
					},
				),
				Entry("duration in range (lower); duration expected out of range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(86400001)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusCombination.NormalMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected missing",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = pointer.FromInt(86400000)
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
				),
				Entry("duration in range (upper); duration expected out of range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = pointer.FromInt(86400000)
						datum.DurationExpected = pointer.FromInt(-1)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusCombination.NormalMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected in range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = pointer.FromInt(86400000)
						datum.DurationExpected = pointer.FromInt(0)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusCombination.NormalMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400000, 0, 0), "/duration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected in range",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = pointer.FromInt(86400000)
						datum.DurationExpected = pointer.FromInt(86399999)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusCombination.NormalMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400000, 0, 86399999), "/duration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected in range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = pointer.FromInt(86400000)
						datum.DurationExpected = pointer.FromInt(86400000)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusCombination.NormalMaximum))
					},
				),
				Entry("duration in range (upper); duration expected out of range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = pointer.FromInt(86400000)
						datum.DurationExpected = pointer.FromInt(86400001)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusCombination.NormalMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected missing",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = pointer.FromInt(86400001)
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected out of range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = pointer.FromInt(86400001)
						datum.DurationExpected = pointer.FromInt(-1)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusCombination.NormalMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected in range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = pointer.FromInt(86400001)
						datum.DurationExpected = pointer.FromInt(0)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusCombination.NormalMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 0), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected in range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = pointer.FromInt(86400001)
						datum.DurationExpected = pointer.FromInt(86400000)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusCombination.NormalMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected out of range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Duration = pointer.FromInt(86400001)
						datum.DurationExpected = pointer.FromInt(86400001)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusCombination.NormalMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("extended missing; extended expected missing",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = nil
						datum.Extended = nil
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
				),
				Entry("extended missing; extended expected out of range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusCombination.DurationMaximum))
						datum.Extended = nil
						datum.ExtendedExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended missing; extended expected in range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusCombination.DurationMaximum))
						datum.Extended = nil
						datum.ExtendedExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
				),
				Entry("extended missing; extended expected in range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusCombination.DurationMaximum))
						datum.Extended = nil
						datum.ExtendedExpected = pointer.FromFloat64(100.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
				),
				Entry("extended missing; extended expected out of range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusCombination.DurationMaximum))
						datum.Extended = nil
						datum.ExtendedExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended out of range (lower); extended expected missing",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = nil
						datum.Extended = pointer.FromFloat64(-0.1)
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/extended", NewMeta()),
				),
				Entry("extended out of range (lower); extended expected out of range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusCombination.DurationMaximum))
						datum.Extended = pointer.FromFloat64(-0.1)
						datum.ExtendedExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended out of range (lower); extended expected in range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusCombination.DurationMaximum))
						datum.Extended = pointer.FromFloat64(-0.1)
						datum.ExtendedExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 0.0), "/extended", NewMeta()),
				),
				Entry("extended out of range (lower); extended expected in range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusCombination.DurationMaximum))
						datum.Extended = pointer.FromFloat64(-0.1)
						datum.ExtendedExpected = pointer.FromFloat64(100.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/extended", NewMeta()),
				),
				Entry("extended out of range (lower); extended expected out of range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusCombination.DurationMaximum))
						datum.Extended = pointer.FromFloat64(-0.1)
						datum.ExtendedExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended in range (lower); extended expected missing",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = nil
						datum.Extended = pointer.FromFloat64(0.0)
						datum.ExtendedExpected = nil
					},
				),
				Entry("extended in range (lower); extended expected out of range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusCombination.DurationMaximum))
						datum.Extended = pointer.FromFloat64(0.0)
						datum.ExtendedExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended in range (lower); extended expected in range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusCombination.DurationMaximum))
						datum.Extended = pointer.FromFloat64(0.0)
						datum.ExtendedExpected = pointer.FromFloat64(0.0)
					},
				),
				Entry("extended in range (lower); extended expected in range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusCombination.DurationMaximum))
						datum.Extended = pointer.FromFloat64(0.0)
						datum.ExtendedExpected = pointer.FromFloat64(100.0)
					},
				),
				Entry("extended in range (lower); extended expected out of range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusCombination.DurationMaximum))
						datum.Extended = pointer.FromFloat64(0.0)
						datum.ExtendedExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended in range (upper); extended expected missing",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = nil
						datum.Extended = pointer.FromFloat64(100.0)
						datum.ExtendedExpected = nil
					},
				),
				Entry("extended in range (upper); extended expected out of range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusCombination.DurationMaximum))
						datum.Extended = pointer.FromFloat64(100.0)
						datum.ExtendedExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended in range (upper); extended expected in range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusCombination.DurationMaximum))
						datum.Extended = pointer.FromFloat64(100.0)
						datum.ExtendedExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100, 0.0, 0.0), "/extended", NewMeta()),
				),
				Entry("extended in range (upper); extended expected in range",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusCombination.DurationMaximum))
						datum.Extended = pointer.FromFloat64(100.0)
						datum.ExtendedExpected = pointer.FromFloat64(99.9)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100, 0.0, 99.9), "/extended", NewMeta()),
				),
				Entry("extended in range (upper); extended expected in range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusCombination.DurationMaximum))
						datum.Extended = pointer.FromFloat64(100.0)
						datum.ExtendedExpected = pointer.FromFloat64(100.0)
					},
				),
				Entry("extended in range (upper); extended expected out of range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusCombination.DurationMaximum))
						datum.Extended = pointer.FromFloat64(100.0)
						datum.ExtendedExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended out of range (upper); extended expected missing",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = nil
						datum.Extended = pointer.FromFloat64(100.1)
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/extended", NewMeta()),
				),
				Entry("extended out of range (upper); extended expected out of range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusCombination.DurationMaximum))
						datum.Extended = pointer.FromFloat64(100.1)
						datum.ExtendedExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended out of range (upper); extended expected in range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusCombination.DurationMaximum))
						datum.Extended = pointer.FromFloat64(100.1)
						datum.ExtendedExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 0.0), "/extended", NewMeta()),
				),
				Entry("extended out of range (upper); extended expected in range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusCombination.DurationMaximum))
						datum.Extended = pointer.FromFloat64(100.1)
						datum.ExtendedExpected = pointer.FromFloat64(100.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/extended", NewMeta()),
				),
				Entry("extended out of range (upper); extended expected out of range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusCombination.DurationMaximum))
						datum.Extended = pointer.FromFloat64(100.1)
						datum.ExtendedExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),

				Entry("duration missing; extended expected missing",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
				),
				Entry("duration missing; extended expected exists",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = nil
						datum.ExtendedExpected = pointer.FromFloat64(100.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/expectedDuration", NewMeta()),
				),
				Entry("duration exists; extended expected missing",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = pointer.FromInt(86400000)
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/expectedDuration", NewMeta()),
				),
				Entry("duration exists; extended expected exists",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.DurationExpected = pointer.FromInt(86400000)
						datum.ExtendedExpected = pointer.FromFloat64(100.0)
					},
				),
				Entry("normal missing; normal expected missing",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = nil
						datum.NormalExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
				),
				Entry("normal missing; normal expected out of range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal missing; normal expected in range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
				),
				Entry("normal missing; normal expected in range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(100.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
				),
				Entry("normal missing; normal expected out of range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected missing",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected out of range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected in range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 0.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected in range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = pointer.FromFloat64(100.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected out of range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (lower); normal expected missing",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = pointer.FromFloat64(0.0)
						datum.NormalExpected = nil
					},
				),
				Entry("normal in range (lower); normal expected out of range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = pointer.FromFloat64(0.0)
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (lower); normal expected in range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = pointer.FromFloat64(0.0)
						datum.NormalExpected = pointer.FromFloat64(0.0)
					},
				),
				Entry("normal in range (lower); normal expected in range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = pointer.FromFloat64(0.0)
						datum.NormalExpected = pointer.FromFloat64(100.0)
					},
				),
				Entry("normal in range (lower); normal expected out of range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = pointer.FromFloat64(0.0)
						datum.NormalExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (upper); normal expected missing",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = pointer.FromFloat64(100.0)
						datum.NormalExpected = nil
					},
				),
				Entry("normal in range (upper); normal expected out of range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = pointer.FromFloat64(100.0)
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (upper); normal expected in range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = pointer.FromFloat64(100.0)
						datum.NormalExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.0, 0.0, 0.0), "/normal", NewMeta()),
				),
				Entry("normal in range (upper); normal expected in range",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = pointer.FromFloat64(100.0)
						datum.NormalExpected = pointer.FromFloat64(99.9)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.0, 0.0, 99.9), "/normal", NewMeta()),
				),
				Entry("normal in range (upper); normal expected in range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = pointer.FromFloat64(100.0)
						datum.NormalExpected = pointer.FromFloat64(100.0)
					},
				),
				Entry("normal in range (upper); normal expected out of range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = pointer.FromFloat64(100.0)
						datum.NormalExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected missing",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = pointer.FromFloat64(100.1)
						datum.NormalExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected out of range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = pointer.FromFloat64(100.1)
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected in range (lower)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = pointer.FromFloat64(100.1)
						datum.NormalExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 0.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected in range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = pointer.FromFloat64(100.1)
						datum.NormalExpected = pointer.FromFloat64(100.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected out of range (upper)",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Normal = pointer.FromFloat64(100.1)
						datum.NormalExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("multiple errors",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(86400001)
						datum.Extended = nil
						datum.ExtendedExpected = pointer.FromFloat64(100.1)
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", dataTypesBolus.Type), "/type", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", dataTypesBolusCombination.SubType), "/subType", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0, 100), "/expectedExtended", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
				Entry("multiple errors",
					func(datum *dataTypesBolusCombination.Combination) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(86400001)
						datum.Extended = nil
						datum.ExtendedExpected = pointer.FromFloat64(100.1)
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", dataTypesBolus.Type), "/type", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", dataTypesBolusCombination.SubType), "/subType", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0, 100), "/expectedExtended", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *dataTypesBolusCombination.Combination)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesBolusCombinationTest.RandomCombination()
						mutator(datum)
						expectedDatum := dataTypesBolusCombinationTest.CloneCombination(datum)
						normalizer := dataNormalizer.New(logTest.NewLogger())
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *dataTypesBolusCombination.Combination) {},
				),
				Entry("does not modify the datum; type missing",
					func(datum *dataTypesBolusCombination.Combination) { datum.Type = "" },
				),
				Entry("does not modify the datum; sub type missing",
					func(datum *dataTypesBolusCombination.Combination) { datum.SubType = "" },
				),
				Entry("does not modify the datum; duration missing",
					func(datum *dataTypesBolusCombination.Combination) { datum.Duration = nil },
				),
				Entry("does not modify the datum; duration expected missing",
					func(datum *dataTypesBolusCombination.Combination) { datum.DurationExpected = nil },
				),
				Entry("does not modify the datum; extended missing",
					func(datum *dataTypesBolusCombination.Combination) { datum.Extended = nil },
				),
				Entry("does not modify the datum; extended expected missing",
					func(datum *dataTypesBolusCombination.Combination) { datum.ExtendedExpected = nil },
				),
				Entry("does not modify the datum; normal missing",
					func(datum *dataTypesBolusCombination.Combination) { datum.Normal = nil },
				),
				Entry("does not modify the datum; normal expected missing",
					func(datum *dataTypesBolusCombination.Combination) { datum.NormalExpected = nil },
				),
			)
		})
	})
})
