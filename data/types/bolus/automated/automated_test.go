package automated_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypesBolus "github.com/tidepool-org/platform/data/types/bolus"
	dataTypesBolusAutomated "github.com/tidepool-org/platform/data/types/bolus/automated"
	dataTypesBolusAutomatedTest "github.com/tidepool-org/platform/data/types/bolus/automated/test"
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
		Type:    "bolus",
		SubType: "automated",
	}
}

var _ = Describe("Automated", func() {
	It("SubType is expected", func() {
		Expect(dataTypesBolusAutomated.SubType).To(Equal("automated"))
	})

	It("NormalMaximum is expected", func() {
		Expect(dataTypesBolusAutomated.NormalMaximum).To(Equal(250.0))
	})

	It("NormalMinimum is expected", func() {
		Expect(dataTypesBolusAutomated.NormalMinimum).To(Equal(0.0))
	})

	Context("Automated", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesBolusAutomated.Automated)) {
				datum := dataTypesBolusAutomatedTest.RandomAutomated()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesBolusAutomatedTest.NewObjectFromAutomated(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesBolusAutomatedTest.NewObjectFromAutomated(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesBolusAutomated.Automated) {},
			),
			Entry("empty",
				func(datum *dataTypesBolusAutomated.Automated) {
					*datum = *dataTypesBolusAutomated.New()
				},
			),
			Entry("all",
				func(datum *dataTypesBolusAutomated.Automated) {
					datum.Normal = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesBolusAutomated.NormalMinimum, dataTypesBolusAutomated.NormalMaximum))
					datum.NormalExpected = pointer.FromFloat64(test.RandomFloat64FromRange(*datum.Normal, dataTypesBolusAutomated.NormalMaximum))
				},
			),
		)

		Context("New", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesBolusAutomated.New()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Type).To(Equal("bolus"))
				Expect(datum.SubType).To(Equal("automated"))
				Expect(datum.Normal).To(BeNil())
				Expect(datum.NormalExpected).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesBolusAutomated.Automated), expectedErrors ...error) {
					expectedDatum := dataTypesBolusAutomatedTest.RandomAutomatedForParser()
					object := dataTypesBolusAutomatedTest.NewObjectFromAutomated(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesBolusAutomated.New()
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesBolusAutomated.Automated) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesBolusAutomated.Automated) {
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
				func(mutator func(datum *dataTypesBolusAutomated.Automated), expectedErrors ...error) {
					datum := dataTypesBolusAutomatedTest.RandomAutomated()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesBolusAutomated.Automated) {},
				),
				Entry("type missing",
					func(datum *dataTypesBolusAutomated.Automated) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &dataTypesBolus.Meta{SubType: "automated"}),
				),
				Entry("type invalid",
					func(datum *dataTypesBolusAutomated.Automated) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "bolus"), "/type", &dataTypesBolus.Meta{Type: "invalidType", SubType: "automated"}),
				),
				Entry("type bolus",
					func(datum *dataTypesBolusAutomated.Automated) { datum.Type = "bolus" },
				),
				Entry("sub type missing",
					func(datum *dataTypesBolusAutomated.Automated) { datum.SubType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &dataTypesBolus.Meta{Type: "bolus"}),
				),
				Entry("sub type invalid",
					func(datum *dataTypesBolusAutomated.Automated) { datum.SubType = "invalidSubType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "automated"), "/subType", &dataTypesBolus.Meta{Type: "bolus", SubType: "invalidSubType"}),
				),
				Entry("sub type automated",
					func(datum *dataTypesBolusAutomated.Automated) { datum.SubType = "automated" },
				),
				Entry("normal missing; normal expected missing",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Normal = nil
						datum.NormalExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
				),
				Entry("normal missing; normal expected out of range (lower)",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, dataTypesBolusAutomated.NormalMaximum), "/expectedNormal", NewMeta()),
				),
				Entry("normal missing; normal expected in range (lower)",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
				),
				Entry("normal missing; normal expected in range (upper)",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(dataTypesBolusAutomated.NormalMaximum)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
				),
				Entry("normal missing; normal expected out of range (upper)",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(250.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(dataTypesBolusAutomated.NormalMaximum+0.1, 0.0, dataTypesBolusAutomated.NormalMaximum), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected missing",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, dataTypesBolusAutomated.NormalMaximum), "/normal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected out of range (lower)",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, dataTypesBolusAutomated.NormalMaximum), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, dataTypesBolusAutomated.NormalMaximum), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected in range (lower)",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, dataTypesBolusAutomated.NormalMaximum), "/normal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected in range (upper)",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = pointer.FromFloat64(dataTypesBolusAutomated.NormalMaximum)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, dataTypesBolusAutomated.NormalMaximum), "/normal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected out of range (upper)",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = pointer.FromFloat64(250.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, dataTypesBolusAutomated.NormalMaximum), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(dataTypesBolusAutomated.NormalMaximum+0.1, 0.0, dataTypesBolusAutomated.NormalMaximum), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (lower); normal expected missing",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Normal = pointer.FromFloat64(0.0)
						datum.NormalExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (lower); normal expected out of range (lower)",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Normal = pointer.FromFloat64(0.0)
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, dataTypesBolusAutomated.NormalMaximum), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (lower); normal expected in range (lower)",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Normal = pointer.FromFloat64(0.0)
						datum.NormalExpected = pointer.FromFloat64(0.0)
					},
				),
				Entry("normal in range (lower); normal expected in range (upper)",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Normal = pointer.FromFloat64(0.0)
						datum.NormalExpected = pointer.FromFloat64(dataTypesBolusAutomated.NormalMaximum)
					},
				),
				Entry("normal in range (lower); normal expected out of range (upper)",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Normal = pointer.FromFloat64(0.0)
						datum.NormalExpected = pointer.FromFloat64(250.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(dataTypesBolusAutomated.NormalMaximum+0.1, 0.0, dataTypesBolusAutomated.NormalMaximum), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (upper); normal expected missing",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Normal = pointer.FromFloat64(dataTypesBolusAutomated.NormalMaximum)
						datum.NormalExpected = nil
					},
				),
				Entry("normal in range (upper); normal expected out of range (lower)",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Normal = pointer.FromFloat64(dataTypesBolusAutomated.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(249.9)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(249.9, dataTypesBolusAutomated.NormalMaximum, dataTypesBolusAutomated.NormalMaximum), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (upper); normal expected in range (lower)",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Normal = pointer.FromFloat64(dataTypesBolusAutomated.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(dataTypesBolusAutomated.NormalMaximum)
					},
				),
				Entry("normal in range (upper); normal expected in range (upper)",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Normal = pointer.FromFloat64(dataTypesBolusAutomated.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(dataTypesBolusAutomated.NormalMaximum)
					},
				),
				Entry("normal in range (upper); normal expected out of range (upper)",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Normal = pointer.FromFloat64(dataTypesBolusAutomated.NormalMaximum)
						datum.NormalExpected = pointer.FromFloat64(250.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(dataTypesBolusAutomated.NormalMaximum+0.1, dataTypesBolusAutomated.NormalMaximum, dataTypesBolusAutomated.NormalMaximum), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected missing",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Normal = pointer.FromFloat64(250.1)
						datum.NormalExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(dataTypesBolusAutomated.NormalMaximum+0.1, 0.0, dataTypesBolusAutomated.NormalMaximum), "/normal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected out of range (lower)",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Normal = pointer.FromFloat64(250.1)
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(dataTypesBolusAutomated.NormalMaximum+0.1, 0.0, dataTypesBolusAutomated.NormalMaximum), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, dataTypesBolusAutomated.NormalMaximum), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected in range (lower)",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Normal = pointer.FromFloat64(250.1)
						datum.NormalExpected = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(dataTypesBolusAutomated.NormalMaximum+0.1, 0.0, dataTypesBolusAutomated.NormalMaximum), "/normal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected in range (upper)",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Normal = pointer.FromFloat64(250.1)
						datum.NormalExpected = pointer.FromFloat64(dataTypesBolusAutomated.NormalMaximum)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(dataTypesBolusAutomated.NormalMaximum+0.1, 0.0, dataTypesBolusAutomated.NormalMaximum), "/normal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected out of range (upper)",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Normal = pointer.FromFloat64(250.1)
						datum.NormalExpected = pointer.FromFloat64(250.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(dataTypesBolusAutomated.NormalMaximum+0.1, 0.0, dataTypesBolusAutomated.NormalMaximum), "/normal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(dataTypesBolusAutomated.NormalMaximum+0.1, 0.0, dataTypesBolusAutomated.NormalMaximum), "/expectedNormal", NewMeta()),
				),
				Entry("multiple errors",
					func(datum *dataTypesBolusAutomated.Automated) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(250.1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "bolus"), "/type", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "automated"), "/subType", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(dataTypesBolusAutomated.NormalMaximum+0.1, 0.0, dataTypesBolusAutomated.NormalMaximum), "/expectedNormal", &dataTypesBolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *dataTypesBolusAutomated.Automated)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesBolusAutomatedTest.RandomAutomated()
						mutator(datum)
						expectedDatum := dataTypesBolusAutomatedTest.CloneAutomated(datum)
						normalizer := dataNormalizer.New(logTest.NewLogger())
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *dataTypesBolusAutomated.Automated) {},
				),
				Entry("does not modify the datum; type missing",
					func(datum *dataTypesBolusAutomated.Automated) { datum.Type = "" },
				),
				Entry("does not modify the datum; sub type missing",
					func(datum *dataTypesBolusAutomated.Automated) { datum.SubType = "" },
				),
				Entry("does not modify the datum; normal missing",
					func(datum *dataTypesBolusAutomated.Automated) { datum.Normal = nil },
				),
				Entry("does not modify the datum; normal expected missing",
					func(datum *dataTypesBolusAutomated.Automated) { datum.NormalExpected = nil },
				),
			)
		})
	})
})
