package normal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/context"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/data/types/bolus/normal"
	testDataTypesBolusNormal "github.com/tidepool-org/platform/data/types/bolus/normal/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewMeta() interface{} {
	return &bolus.Meta{
		Type:    "bolus",
		SubType: "normal",
	}
}

func NewTestNormal(sourceTime interface{}, sourceNormal interface{}, sourceNormalExpected interface{}) *normal.Normal {
	datum := normal.New()
	datum.DeviceID = pointer.FromString(testData.NewDeviceID())
	if val, ok := sourceTime.(string); ok {
		datum.Time = &val
	}
	if val, ok := sourceNormal.(float64); ok {
		datum.Normal = &val
	}
	if val, ok := sourceNormalExpected.(float64); ok {
		datum.NormalExpected = &val
	}
	return datum
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
			var datum *normal.Normal

			BeforeEach(func() {
				datum = normal.New()
				Expect(datum).ToNot(BeNil())
			})

			DescribeTable("parses the datum",
				func(sourceObject *map[string]interface{}, expectedDatum *normal.Normal, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(null.NewLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testParser, err := parser.NewStandardObject(testContext, sourceObject, parser.AppendErrorNotParsed)
					Expect(err).ToNot(HaveOccurred())
					Expect(testParser).ToNot(BeNil())
					Expect(datum.Parse(testParser)).To(Succeed())
					Expect(datum.Time).To(Equal(expectedDatum.Time))
					Expect(datum.Normal).To(Equal(expectedDatum.Normal))
					Expect(datum.NormalExpected).To(Equal(expectedDatum.NormalExpected))
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("parses object that is nil",
					nil,
					NewTestNormal(nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that is empty",
					&map[string]interface{}{},
					NewTestNormal(nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has valid time",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00"},
					NewTestNormal("2016-09-06T13:45:58-07:00", nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid time",
					&map[string]interface{}{"time": 0.0},
					NewTestNormal(nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0.0), "/time", NewMeta()),
					}),
				Entry("parses object that has valid normal",
					&map[string]interface{}{"normal": 3.6},
					NewTestNormal(nil, 3.6, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid normal",
					&map[string]interface{}{"normal": "invalid"},
					NewTestNormal(nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/normal", NewMeta()),
					}),
				Entry("parses object that has valid normal expected",
					&map[string]interface{}{"expectedNormal": 7.2},
					NewTestNormal(nil, nil, 7.2),
					[]*service.Error{}),
				Entry("parses object that has invalid normal expected",
					&map[string]interface{}{"expectedNormal": "invalid"},
					NewTestNormal(nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/expectedNormal", NewMeta()),
					}),
				Entry("parses object that has multiple valid fields",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00", "normal": 3.6, "expectedNormal": 7.2},
					NewTestNormal("2016-09-06T13:45:58-07:00", 3.6, 7.2),
					[]*service.Error{}),
				Entry("parses object that has multiple invalid fields",
					&map[string]interface{}{"time": 0.0, "normal": "invalid", "expectedNormal": "invalid"},
					NewTestNormal(nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0.0), "/time", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/normal", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/expectedNormal", NewMeta()),
					}),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *normal.Normal), expectedErrors ...error) {
					datum := testDataTypesBolusNormal.NewNormal()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *normal.Normal) {},
				),
				Entry("type missing",
					func(datum *normal.Normal) { datum.Type = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &bolus.Meta{SubType: "normal"}),
				),
				Entry("type invalid",
					func(datum *normal.Normal) { datum.Type = "invalidType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "bolus"), "/type", &bolus.Meta{Type: "invalidType", SubType: "normal"}),
				),
				Entry("type bolus",
					func(datum *normal.Normal) { datum.Type = "bolus" },
				),
				Entry("sub type missing",
					func(datum *normal.Normal) { datum.SubType = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &bolus.Meta{Type: "bolus"}),
				),
				Entry("sub type invalid",
					func(datum *normal.Normal) { datum.SubType = "invalidSubType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "normal"), "/subType", &bolus.Meta{Type: "bolus", SubType: "invalidSubType"}),
				),
				Entry("sub type normal",
					func(datum *normal.Normal) { datum.SubType = "normal" },
				),
				Entry("normal missing; normal expected missing",
					func(datum *normal.Normal) {
						datum.Normal = nil
						datum.NormalExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
				),
				Entry("normal missing; normal expected out of range (lower)",
					func(datum *normal.Normal) {
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal missing; normal expected in range (lower)",
					func(datum *normal.Normal) {
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(0.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
				),
				Entry("normal missing; normal expected in range (upper)",
					func(datum *normal.Normal) {
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(100.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
				),
				Entry("normal missing; normal expected out of range (upper)",
					func(datum *normal.Normal) {
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(100.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected missing",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected out of range (lower)",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected in range (lower)",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = pointer.FromFloat64(0.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected in range (upper)",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = pointer.FromFloat64(100.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected out of range (upper)",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(-0.1)
						datum.NormalExpected = pointer.FromFloat64(100.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (lower); normal expected missing",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(0.0)
						datum.NormalExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (lower); normal expected out of range (lower)",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(0.0)
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
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
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (upper); normal expected missing",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(100.0)
						datum.NormalExpected = nil
					},
				),
				Entry("normal in range (upper); normal expected out of range (lower)",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(100.0)
						datum.NormalExpected = pointer.FromFloat64(99.9)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(99.9, 100.0, 100.0), "/expectedNormal", NewMeta()),
				),
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
				Entry("normal in range (upper); normal expected out of range (upper)",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(100.0)
						datum.NormalExpected = pointer.FromFloat64(100.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 100.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected missing",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(100.1)
						datum.NormalExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected out of range (lower)",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(100.1)
						datum.NormalExpected = pointer.FromFloat64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected in range (lower)",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(100.1)
						datum.NormalExpected = pointer.FromFloat64(0.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected in range (upper)",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(100.1)
						datum.NormalExpected = pointer.FromFloat64(100.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected out of range (upper)",
					func(datum *normal.Normal) {
						datum.Normal = pointer.FromFloat64(100.1)
						datum.NormalExpected = pointer.FromFloat64(100.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("multiple errors",
					func(datum *normal.Normal) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
						datum.Normal = nil
						datum.NormalExpected = pointer.FromFloat64(100.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "bolus"), "/type", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "normal"), "/subType", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *normal.Normal)) {
					for _, origin := range structure.Origins() {
						datum := testDataTypesBolusNormal.NewNormal()
						mutator(datum)
						expectedDatum := testDataTypesBolusNormal.CloneNormal(datum)
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
