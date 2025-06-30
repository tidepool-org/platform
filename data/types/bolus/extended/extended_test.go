package extended_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypesBolus "github.com/tidepool-org/platform/data/types/bolus"
	dataTypesBolusExtended "github.com/tidepool-org/platform/data/types/bolus/extended"
	dataTypesBolusExtendedTest "github.com/tidepool-org/platform/data/types/bolus/extended/test"
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
		SubType: dataTypesBolusExtended.SubType,
	}
}

var _ = Describe("Extended", func() {
	It("SubType is expected", func() {
		Expect(dataTypesBolusExtended.SubType).To(Equal("square"))
	})

	It("DurationMaximum is expected", func() {
		Expect(dataTypesBolusExtended.DurationMaximum).To(Equal(86400000))
	})

	It("DurationMinimum is expected", func() {
		Expect(dataTypesBolusExtended.DurationMinimum).To(Equal(0))
	})

	It("ExtendedMaximum is expected", func() {
		Expect(dataTypesBolusExtended.ExtendedMaximum).To(Equal(100.0))
	})

	It("ExtendedMinimum is expected", func() {
		Expect(dataTypesBolusExtended.ExtendedMinimum).To(Equal(0.0))
	})

	Context("Extended", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesBolusExtended.Extended)) {
				datum := dataTypesBolusExtendedTest.RandomExtended()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesBolusExtendedTest.NewObjectFromExtended(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesBolusExtendedTest.NewObjectFromExtended(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesBolusExtended.Extended) {},
			),
			Entry("empty",
				func(datum *dataTypesBolusExtended.Extended) {
					*datum = *dataTypesBolusExtended.New()
				},
			),
			Entry("all",
				func(datum *dataTypesBolusExtended.Extended) {
					datum.Duration = pointer.FromInt(test.RandomIntFromRange(dataTypesBolusExtended.DurationMinimum, dataTypesBolusExtended.DurationMaximum))
					datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusExtended.DurationMaximum))
					datum.Extended = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesBolusExtended.ExtendedMinimum, dataTypesBolusExtended.ExtendedMaximum))
					datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusExtended.ExtendedMaximum))
				},
			),
		)

		Context("New", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesBolusExtended.New()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Type).To(Equal(dataTypesBolus.Type))
				Expect(datum.SubType).To(Equal(dataTypesBolusExtended.SubType))
				Expect(datum.Duration).To(BeNil())
				Expect(datum.DurationExpected).To(BeNil())
				Expect(datum.Extended).To(BeNil())
				Expect(datum.ExtendedExpected).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *dataTypesBolusExtended.Extended), expectedErrors ...error) {
					expectedDatum := dataTypesBolusExtendedTest.RandomExtendedForParser()
					object := dataTypesBolusExtendedTest.NewObjectFromExtended(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesBolusExtended.New()
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *dataTypesBolusExtended.Extended) {},
				),
				Entry("multiple errors",
					func(object map[string]any, expectedDatum *dataTypesBolusExtended.Extended) {
						object["duration"] = true
						object["expectedDuration"] = true
						object["extended"] = true
						object["expectedExtended"] = true
						expectedDatum.Duration = nil
						expectedDatum.DurationExpected = nil
						expectedDatum.Extended = nil
						expectedDatum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotInt(true), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotInt(true), "/expectedDuration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotFloat64(true), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotFloat64(true), "/expectedExtended", NewMeta()),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesBolusExtended.Extended), expectedErrors ...error) {
					datum := dataTypesBolusExtendedTest.RandomExtended()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesBolusExtended.Extended) {},
				),
				Entry("type missing",
					func(datum *dataTypesBolusExtended.Extended) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &dataTypesBolus.Meta{SubType: dataTypesBolusExtended.SubType}),
				),
				Entry("type invalid",
					func(datum *dataTypesBolusExtended.Extended) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", dataTypesBolus.Type), "/type", &dataTypesBolus.Meta{Type: "invalidType", SubType: dataTypesBolusExtended.SubType}),
				),
				Entry("type bolus",
					func(datum *dataTypesBolusExtended.Extended) { datum.Type = dataTypesBolus.Type },
				),
				Entry("sub type missing",
					func(datum *dataTypesBolusExtended.Extended) { datum.SubType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &dataTypesBolus.Meta{Type: dataTypesBolus.Type}),
				),
				Entry("sub type invalid",
					func(datum *dataTypesBolusExtended.Extended) { datum.SubType = "invalidSubType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", dataTypesBolusExtended.SubType), "/subType", &dataTypesBolus.Meta{Type: dataTypesBolus.Type, SubType: "invalidSubType"}),
				),
				Entry("sub type square",
					func(datum *dataTypesBolusExtended.Extended) { datum.SubType = dataTypesBolusExtended.SubType },
				),
				Entry("duration missing; duration expected missing",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = nil
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (lower)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(-1)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusExtended.ExtendedMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (lower)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(0)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusExtended.ExtendedMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (upper)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(86400000)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusExtended.ExtendedMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (upper)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(86400001)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusExtended.ExtendedMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected missing",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (lower)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(-1)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusExtended.ExtendedMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (lower)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(0)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusExtended.ExtendedMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 0), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (upper)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(86400000)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusExtended.ExtendedMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (upper)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(86400001)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusExtended.ExtendedMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected missing",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
				),
				Entry("duration in range (lower); duration expected out of range (lower)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(-1)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusExtended.ExtendedMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected in range (lower)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(0)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusExtended.ExtendedMaximum))
					},
				),
				Entry("duration in range (lower); duration expected in range (upper)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(86400000)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusExtended.ExtendedMaximum))
					},
				),
				Entry("duration in range (lower); duration expected out of range (upper)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(86400001)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusExtended.ExtendedMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected missing",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = pointer.FromInt(86400000)
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
				),
				Entry("duration in range (upper); duration expected out of range (lower)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = pointer.FromInt(86400000)
						datum.DurationExpected = pointer.FromInt(-1)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusExtended.ExtendedMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected in range (lower)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = pointer.FromInt(86400000)
						datum.DurationExpected = pointer.FromInt(0)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusExtended.ExtendedMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400000, 0, 0), "/duration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected in range",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = pointer.FromInt(86400000)
						datum.DurationExpected = pointer.FromInt(86399999)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusExtended.ExtendedMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400000, 0, 86399999), "/duration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected in range (upper)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = pointer.FromInt(86400000)
						datum.DurationExpected = pointer.FromInt(86400000)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusExtended.ExtendedMaximum))
					},
				),
				Entry("duration in range (upper); duration expected out of range (upper)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = pointer.FromInt(86400000)
						datum.DurationExpected = pointer.FromInt(86400001)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusExtended.ExtendedMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected missing",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = pointer.FromInt(86400001)
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected out of range (lower)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = pointer.FromInt(86400001)
						datum.DurationExpected = pointer.FromInt(-1)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusExtended.ExtendedMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected in range (lower)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = pointer.FromInt(86400001)
						datum.DurationExpected = pointer.FromInt(0)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusExtended.ExtendedMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 0), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected in range (upper)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = pointer.FromInt(86400001)
						datum.DurationExpected = pointer.FromInt(86400000)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusExtended.ExtendedMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected out of range (upper)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Duration = pointer.FromInt(86400001)
						datum.DurationExpected = pointer.FromInt(86400001)
						datum.ExtendedExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Extended, dataTypesBolusExtended.ExtendedMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("extended missing; extended expected missing",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = nil
						datum.Extended = nil
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
				),
				Entry("extended missing; extended expected out of range (lower)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusExtended.DurationMaximum))
						datum.Extended = nil
						datum.ExtendedExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended missing; extended expected in range (lower)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusExtended.DurationMaximum))
						datum.Extended = nil
						datum.ExtendedExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
				),
				Entry("extended missing; extended expected in range (upper)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusExtended.DurationMaximum))
						datum.Extended = nil
						datum.ExtendedExpected = pointer.FromFloat64(100.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
				),
				Entry("extended missing; extended expected out of range (upper)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusExtended.DurationMaximum))
						datum.Extended = nil
						datum.ExtendedExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended out of range (lower); extended expected missing",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = nil
						datum.Extended = pointer.FromFloat64(-0.1)
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/extended", NewMeta()),
				),
				Entry("extended out of range (lower); extended expected out of range (lower)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusExtended.DurationMaximum))
						datum.Extended = pointer.FromFloat64(-0.1)
						datum.ExtendedExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended out of range (lower); extended expected in range (lower)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusExtended.DurationMaximum))
						datum.Extended = pointer.FromFloat64(-0.1)
						datum.ExtendedExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 0.0), "/extended", NewMeta()),
				),
				Entry("extended out of range (lower); extended expected in range (upper)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusExtended.DurationMaximum))
						datum.Extended = pointer.FromFloat64(-0.1)
						datum.ExtendedExpected = pointer.FromFloat64(100.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/extended", NewMeta()),
				),
				Entry("extended out of range (lower); extended expected out of range (upper)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusExtended.DurationMaximum))
						datum.Extended = pointer.FromFloat64(-0.1)
						datum.ExtendedExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended in range (lower); extended expected missing",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = nil
						datum.Extended = pointer.FromFloat64(0.0)
						datum.ExtendedExpected = nil
					},
				),
				Entry("extended in range (lower); extended expected out of range (lower)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusExtended.DurationMaximum))
						datum.Extended = pointer.FromFloat64(0.0)
						datum.ExtendedExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended in range (lower); extended expected in range (lower)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusExtended.DurationMaximum))
						datum.Extended = pointer.FromFloat64(0.0)
						datum.ExtendedExpected = pointer.FromFloat64(0.0)
					},
				),
				Entry("extended in range (lower); extended expected in range (upper)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusExtended.DurationMaximum))
						datum.Extended = pointer.FromFloat64(0.0)
						datum.ExtendedExpected = pointer.FromFloat64(100.0)
					},
				),
				Entry("extended in range (lower); extended expected out of range (upper)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusExtended.DurationMaximum))
						datum.Extended = pointer.FromFloat64(0.0)
						datum.ExtendedExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended in range (upper); extended expected missing",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = nil
						datum.Extended = pointer.FromFloat64(100.0)
						datum.ExtendedExpected = nil
					},
				),
				Entry("extended in range (upper); extended expected out of range (lower)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusExtended.DurationMaximum))
						datum.Extended = pointer.FromFloat64(100.0)
						datum.ExtendedExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended in range (upper); extended expected in range (lower)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusExtended.DurationMaximum))
						datum.Extended = pointer.FromFloat64(100.0)
						datum.ExtendedExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100, 0.0, 0.0), "/extended", NewMeta()),
				),
				Entry("extended in range (upper); extended expected in range",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusExtended.DurationMaximum))
						datum.Extended = pointer.FromFloat64(100.0)
						datum.ExtendedExpected = pointer.FromFloat64(99.9)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100, 0.0, 99.9), "/extended", NewMeta()),
				),
				Entry("extended in range (upper); extended expected in range (upper)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusExtended.DurationMaximum))
						datum.Extended = pointer.FromFloat64(100.0)
						datum.ExtendedExpected = pointer.FromFloat64(100.0)
					},
				),
				Entry("extended in range (upper); extended expected out of range (upper)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusExtended.DurationMaximum))
						datum.Extended = pointer.FromFloat64(100.0)
						datum.ExtendedExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended out of range (upper); extended expected missing",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = nil
						datum.Extended = pointer.FromFloat64(100.1)
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/extended", NewMeta()),
				),
				Entry("extended out of range (upper); extended expected out of range (lower)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusExtended.DurationMaximum))
						datum.Extended = pointer.FromFloat64(100.1)
						datum.ExtendedExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended out of range (upper); extended expected in range (lower)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusExtended.DurationMaximum))
						datum.Extended = pointer.FromFloat64(100.1)
						datum.ExtendedExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 0.0), "/extended", NewMeta()),
				),
				Entry("extended out of range (upper); extended expected in range (upper)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusExtended.DurationMaximum))
						datum.Extended = pointer.FromFloat64(100.1)
						datum.ExtendedExpected = pointer.FromFloat64(100.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/extended", NewMeta()),
				),
				Entry("extended out of range (upper); extended expected out of range (upper)",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBolusExtended.DurationMaximum))
						datum.Extended = pointer.FromFloat64(100.1)
						datum.ExtendedExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/extended", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),

				Entry("duration missing; extended expected missing",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
				),
				Entry("duration missing; extended expected exists",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = nil
						datum.ExtendedExpected = pointer.FromFloat64(100.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/expectedDuration", NewMeta()),
				),
				Entry("duration exists; extended expected missing",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = pointer.FromInt(86400000)
						datum.ExtendedExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/expectedDuration", NewMeta()),
				),
				Entry("duration exists; extended expected exists",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.DurationExpected = pointer.FromInt(86400000)
						datum.ExtendedExpected = pointer.FromFloat64(100.0)
					},
				),
				Entry("multiple errors",
					func(datum *dataTypesBolusExtended.Extended) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(86400001)
						datum.Extended = nil
						datum.ExtendedExpected = pointer.FromFloat64(100.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", dataTypesBolus.Type), "/type", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", dataTypesBolusExtended.SubType), "/subType", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0, 100), "/expectedExtended", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *dataTypesBolusExtended.Extended)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesBolusExtendedTest.RandomExtended()
						mutator(datum)
						expectedDatum := dataTypesBolusExtendedTest.CloneExtended(datum)
						normalizer := dataNormalizer.New(logTest.NewLogger())
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *dataTypesBolusExtended.Extended) {},
				),
				Entry("does not modify the datum; type missing",
					func(datum *dataTypesBolusExtended.Extended) { datum.Type = "" },
				),
				Entry("does not modify the datum; sub type missing",
					func(datum *dataTypesBolusExtended.Extended) { datum.SubType = "" },
				),
				Entry("does not modify the datum; duration missing",
					func(datum *dataTypesBolusExtended.Extended) { datum.Duration = nil },
				),
				Entry("does not modify the datum; duration expected missing",
					func(datum *dataTypesBolusExtended.Extended) { datum.DurationExpected = nil },
				),
				Entry("does not modify the datum; extended missing",
					func(datum *dataTypesBolusExtended.Extended) { datum.Extended = nil },
				),
				Entry("does not modify the datum; extended expected missing",
					func(datum *dataTypesBolusExtended.Extended) { datum.ExtendedExpected = nil },
				),
			)
		})
	})
})
