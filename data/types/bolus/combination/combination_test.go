package combination_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/factory"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/data/types/bolus/combination"
	testDataTypesBolusCombination "github.com/tidepool-org/platform/data/types/bolus/combination/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &bolus.Meta{
		Type:    "bolus",
		SubType: "dual/square",
	}
}

func NewTestCombination(sourceTime interface{}, sourceDuration interface{}, sourceDurationExpected interface{}, sourceExtended interface{}, sourceExtendedExpected interface{}, sourceNormal interface{}, sourceNormalExpected interface{}) *combination.Combination {
	datum := combination.Init()
	datum.DeviceID = pointer.String(id.New())
	if val, ok := sourceTime.(string); ok {
		datum.Time = &val
	}
	if val, ok := sourceDuration.(int); ok {
		datum.Duration = &val
	}
	if val, ok := sourceDurationExpected.(int); ok {
		datum.DurationExpected = &val
	}
	if val, ok := sourceExtended.(float64); ok {
		datum.Extended = &val
	}
	if val, ok := sourceExtendedExpected.(float64); ok {
		datum.ExtendedExpected = &val
	}
	if val, ok := sourceNormal.(float64); ok {
		datum.Normal = &val
	}
	if val, ok := sourceNormalExpected.(float64); ok {
		datum.NormalExpected = &val
	}
	return datum
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
		Expect(combination.ExtendedMaximum).To(Equal(100.0))
	})

	It("ExtendedMinimum is expected", func() {
		Expect(combination.ExtendedMinimum).To(Equal(0.0))
	})

	It("NormalMaximum is expected", func() {
		Expect(combination.NormalMaximum).To(Equal(100.0))
	})

	It("NormalMinimum is expected", func() {
		Expect(combination.NormalMinimum).To(Equal(0.0))
	})

	Context("NewDatum", func() {
		It("returns the expected datum", func() {
			Expect(combination.NewDatum()).To(Equal(&combination.Combination{}))
		})
	})

	Context("New", func() {
		It("returns the expected datum", func() {
			Expect(combination.New()).To(Equal(&combination.Combination{}))
		})
	})

	Context("Init", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := combination.Init()
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

	Context("with new datum", func() {
		var datum *combination.Combination

		BeforeEach(func() {
			datum = testDataTypesBolusCombination.NewCombination()
		})

		Context("Init", func() {
			It("initializes the datum", func() {
				datum.Init()
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
	})

	Context("Combination", func() {
		Context("Parse", func() {
			var datum *combination.Combination

			BeforeEach(func() {
				datum = combination.Init()
				Expect(datum).ToNot(BeNil())
			})

			DescribeTable("parses the datum",
				func(sourceObject *map[string]interface{}, expectedDatum *combination.Combination, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(null.NewLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testFactory, err := factory.NewStandard()
					Expect(err).ToNot(HaveOccurred())
					Expect(testFactory).ToNot(BeNil())
					testParser, err := parser.NewStandardObject(testContext, testFactory, sourceObject, parser.AppendErrorNotParsed)
					Expect(err).ToNot(HaveOccurred())
					Expect(testParser).ToNot(BeNil())
					Expect(datum.Parse(testParser)).To(Succeed())
					Expect(datum.Time).To(Equal(expectedDatum.Time))
					Expect(datum.Duration).To(Equal(expectedDatum.Duration))
					Expect(datum.DurationExpected).To(Equal(expectedDatum.DurationExpected))
					Expect(datum.Extended).To(Equal(expectedDatum.Extended))
					Expect(datum.ExtendedExpected).To(Equal(expectedDatum.ExtendedExpected))
					Expect(datum.Normal).To(Equal(expectedDatum.Normal))
					Expect(datum.NormalExpected).To(Equal(expectedDatum.NormalExpected))
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("parses object that is nil",
					nil,
					NewTestCombination(nil, nil, nil, nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that is empty",
					&map[string]interface{}{},
					NewTestCombination(nil, nil, nil, nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has valid time",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00"},
					NewTestCombination("2016-09-06T13:45:58-07:00", nil, nil, nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid time",
					&map[string]interface{}{"time": 0},
					NewTestCombination(nil, nil, nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
					}),
				Entry("parses object that has valid duration",
					&map[string]interface{}{"duration": 1000000},
					NewTestCombination(nil, 1000000, nil, nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid duration",
					&map[string]interface{}{"duration": "invalid"},
					NewTestCombination(nil, nil, nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/duration", NewMeta()),
					}),
				Entry("parses object that has valid duration expected",
					&map[string]interface{}{"expectedDuration": 2000000},
					NewTestCombination(nil, nil, 2000000, nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid duration expected",
					&map[string]interface{}{"expectedDuration": "invalid"},
					NewTestCombination(nil, nil, nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/expectedDuration", NewMeta()),
					}),
				Entry("parses object that has valid extended",
					&map[string]interface{}{"extended": 3.6},
					NewTestCombination(nil, nil, nil, 3.6, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid extended",
					&map[string]interface{}{"extended": "invalid"},
					NewTestCombination(nil, nil, nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/extended", NewMeta()),
					}),
				Entry("parses object that has valid extended expected",
					&map[string]interface{}{"expectedExtended": 7.2},
					NewTestCombination(nil, nil, nil, nil, 7.2, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid extended expected",
					&map[string]interface{}{"expectedExtended": "invalid"},
					NewTestCombination(nil, nil, nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/expectedExtended", NewMeta()),
					}),
				Entry("parses object that has valid normal",
					&map[string]interface{}{"normal": 5.4},
					NewTestCombination(nil, nil, nil, nil, nil, 5.4, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid normal",
					&map[string]interface{}{"normal": "invalid"},
					NewTestCombination(nil, nil, nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/normal", NewMeta()),
					}),
				Entry("parses object that has valid normal expected",
					&map[string]interface{}{"expectedNormal": 9.1},
					NewTestCombination(nil, nil, nil, nil, nil, nil, 9.1),
					[]*service.Error{}),
				Entry("parses object that has invalid normal expected",
					&map[string]interface{}{"expectedNormal": "invalid"},
					NewTestCombination(nil, nil, nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/expectedNormal", NewMeta()),
					}),
				Entry("parses object that has multiple valid fields",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00", "duration": 1000000, "expectedDuration": 2000000, "extended": 3.6, "expectedExtended": 7.2, "normal": 5.4, "expectedNormal": 9.1},
					NewTestCombination("2016-09-06T13:45:58-07:00", 1000000, 2000000, 3.6, 7.2, 5.4, 9.1),
					[]*service.Error{}),
				Entry("parses object that has multiple invalid fields",
					&map[string]interface{}{"time": 0, "duration": "invalid", "expectedDuration": "invalid", "extended": "invalid", "expectedExtended": "invalid", "normal": "invalid", "expectedNormal": "invalid"},
					NewTestCombination(nil, nil, nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/duration", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/expectedDuration", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/extended", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/expectedExtended", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/normal", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/expectedNormal", NewMeta()),
					}),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *combination.Combination), expectedErrors ...error) {
					datum := testDataTypesBolusCombination.NewCombination()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *combination.Combination) {},
				),
				Entry("type missing",
					func(datum *combination.Combination) { datum.Type = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &bolus.Meta{SubType: "dual/square"}),
				),
				Entry("type invalid",
					func(datum *combination.Combination) { datum.Type = "invalidType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "bolus"), "/type", &bolus.Meta{Type: "invalidType", SubType: "dual/square"}),
				),
				Entry("type bolus",
					func(datum *combination.Combination) { datum.Type = "bolus" },
				),
				Entry("sub type missing",
					func(datum *combination.Combination) { datum.SubType = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &bolus.Meta{Type: "bolus"}),
				),
				Entry("sub type invalid",
					func(datum *combination.Combination) { datum.SubType = "invalidSubType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "dual/square"), "/subType", &bolus.Meta{Type: "bolus", SubType: "invalidSubType"}),
				),
				Entry("sub type dual/square",
					func(datum *combination.Combination) { datum.SubType = "dual/square" },
				),
				Entry("normal expected missing; duration missing; duration expected missing",
					func(datum *combination.Combination) {
						datum.Duration = nil
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("normal expected missing; duration missing; duration expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("normal expected missing; duration missing; duration expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("normal expected missing; duration missing; duration expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(86400000)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("normal expected missing; duration missing; duration expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(86400001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("normal expected missing; duration out of range (lower); duration expected missing",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/duration", NewMeta()),
				),
				Entry("normal expected missing; duration out of range (lower); duration expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("normal expected missing; duration out of range (lower); duration expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/duration", NewMeta()),
				),
				Entry("normal expected missing; duration out of range (lower); duration expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(86400000)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/duration", NewMeta()),
				),
				Entry("normal expected missing; duration out of range (lower); duration expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(86400001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("normal expected missing; duration in range (lower); duration expected missing",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
				),
				Entry("normal expected missing; duration in range (lower); duration expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("normal expected missing; duration in range (lower); duration expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(0)
					},
				),
				Entry("normal expected missing; duration in range (lower); duration expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(86400000)
					},
				),
				Entry("normal expected missing; duration in range (lower); duration expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(86400001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("normal expected missing; duration in range (upper); duration expected missing",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(86400000)
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
				),
				Entry("normal expected missing; duration in range (upper); duration expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(86400000)
						datum.DurationExpected = pointer.Int(86399999)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86399999, 86400000, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("normal expected missing; duration in range (upper); duration expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(86400000)
						datum.DurationExpected = pointer.Int(86400000)
					},
				),
				Entry("normal expected missing; duration in range (upper); duration expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(86400000)
						datum.DurationExpected = pointer.Int(86400000)
					},
				),
				Entry("normal expected missing; duration in range (upper); duration expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(86400000)
						datum.DurationExpected = pointer.Int(86400001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 86400000, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("normal expected missing; duration out of range (upper); duration expected missing",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(86400001)
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/duration", NewMeta()),
				),
				Entry("normal expected missing; duration out of range (upper); duration expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(86400001)
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("normal expected missing; duration out of range (upper); duration expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(86400001)
						datum.DurationExpected = pointer.Int(0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/duration", NewMeta()),
				),
				Entry("normal expected missing; duration out of range (upper); duration expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(86400001)
						datum.DurationExpected = pointer.Int(86400000)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/duration", NewMeta()),
				),
				Entry("normal expected missing; duration out of range (upper); duration expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(86400001)
						datum.DurationExpected = pointer.Int(86400001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("normal expected missing; extended missing; extended expected missing",
					func(datum *combination.Combination) {
						datum.DurationExpected = nil
						datum.Extended = nil
						datum.ExtendedExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
				),
				Entry("normal expected missing; extended missing; extended expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Extended = nil
						datum.ExtendedExpected = pointer.Float64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("normal expected missing; extended missing; extended expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Extended = nil
						datum.ExtendedExpected = pointer.Float64(0.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
				),
				Entry("normal expected missing; extended missing; extended expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Extended = nil
						datum.ExtendedExpected = pointer.Float64(100.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
				),
				Entry("normal expected missing; extended missing; extended expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Extended = nil
						datum.ExtendedExpected = pointer.Float64(100.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("normal expected missing; extended out of range (lower); extended expected missing",
					func(datum *combination.Combination) {
						datum.DurationExpected = nil
						datum.Extended = pointer.Float64(-0.1)
						datum.ExtendedExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/extended", NewMeta()),
				),
				Entry("normal expected missing; extended out of range (lower); extended expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Extended = pointer.Float64(-0.1)
						datum.ExtendedExpected = pointer.Float64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("normal expected missing; extended out of range (lower); extended expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Extended = pointer.Float64(-0.1)
						datum.ExtendedExpected = pointer.Float64(0.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/extended", NewMeta()),
				),
				Entry("normal expected missing; extended out of range (lower); extended expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Extended = pointer.Float64(-0.1)
						datum.ExtendedExpected = pointer.Float64(100.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/extended", NewMeta()),
				),
				Entry("normal expected missing; extended out of range (lower); extended expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Extended = pointer.Float64(-0.1)
						datum.ExtendedExpected = pointer.Float64(100.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("normal expected missing; extended in range (lower); extended expected missing",
					func(datum *combination.Combination) {
						datum.DurationExpected = nil
						datum.Extended = pointer.Float64(0.0)
						datum.ExtendedExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/expectedExtended", NewMeta()),
				),
				Entry("normal expected missing; extended in range (lower); extended expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Extended = pointer.Float64(0.0)
						datum.ExtendedExpected = pointer.Float64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("normal expected missing; extended in range (lower); extended expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Extended = pointer.Float64(0.0)
						datum.ExtendedExpected = pointer.Float64(0.0)
					},
				),
				Entry("normal expected missing; extended in range (lower); extended expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Extended = pointer.Float64(0.0)
						datum.ExtendedExpected = pointer.Float64(100.0)
					},
				),
				Entry("normal expected missing; extended in range (lower); extended expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Extended = pointer.Float64(0.0)
						datum.ExtendedExpected = pointer.Float64(100.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("normal expected missing; extended in range (upper); extended expected missing",
					func(datum *combination.Combination) {
						datum.DurationExpected = nil
						datum.Extended = pointer.Float64(100.0)
						datum.ExtendedExpected = nil
					},
				),
				Entry("normal expected missing; extended in range (upper); extended expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Extended = pointer.Float64(100.0)
						datum.ExtendedExpected = pointer.Float64(99.9)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(99.9, 100.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("normal expected missing; extended in range (upper); extended expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Extended = pointer.Float64(100.0)
						datum.ExtendedExpected = pointer.Float64(100.0)
					},
				),
				Entry("normal expected missing; extended in range (upper); extended expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Extended = pointer.Float64(100.0)
						datum.ExtendedExpected = pointer.Float64(100.0)
					},
				),
				Entry("normal expected missing; extended in range (upper); extended expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Extended = pointer.Float64(100.0)
						datum.ExtendedExpected = pointer.Float64(100.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 100.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("normal expected missing; extended out of range (upper); extended expected missing",
					func(datum *combination.Combination) {
						datum.DurationExpected = nil
						datum.Extended = pointer.Float64(100.1)
						datum.ExtendedExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/extended", NewMeta()),
				),
				Entry("normal expected missing; extended out of range (upper); extended expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Extended = pointer.Float64(100.1)
						datum.ExtendedExpected = pointer.Float64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("normal expected missing; extended out of range (upper); extended expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Extended = pointer.Float64(100.1)
						datum.ExtendedExpected = pointer.Float64(0.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/extended", NewMeta()),
				),
				Entry("normal expected missing; extended out of range (upper); extended expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Extended = pointer.Float64(100.1)
						datum.ExtendedExpected = pointer.Float64(100.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/extended", NewMeta()),
				),
				Entry("normal expected missing; extended out of range (upper); extended expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Extended = pointer.Float64(100.1)
						datum.ExtendedExpected = pointer.Float64(100.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("normal expected missing; duration missing; extended expected missing",
					func(datum *combination.Combination) {
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
						datum.NormalExpected = nil
					},
				),
				Entry("normal expected missing; duration missing; extended expected exists",
					func(datum *combination.Combination) {
						datum.DurationExpected = nil
						datum.NormalExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/expectedDuration", NewMeta()),
				),
				Entry("normal expected missing; duration exists; extended expected missing",
					func(datum *combination.Combination) {
						datum.ExtendedExpected = nil
						datum.NormalExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/expectedDuration", NewMeta()),
				),
				Entry("normal expected missing; duration exists; extended expected exists",
					func(datum *combination.Combination) {
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
				),
				Entry("normal expected exists; duration missing; duration expected missing",
					func(datum *combination.Combination) {
						datum.Duration = nil
						datum.DurationExpected = nil
						datum.Extended = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/expectedDuration", NewMeta()),
				),
				Entry("normal expected exists; duration missing; duration expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(-1)
						datum.Extended = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("normal expected exists; duration missing; duration expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("normal expected exists; duration missing; duration expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(86400000)
						datum.Extended = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("normal expected exists; duration missing; duration expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(86400001)
						datum.Extended = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("normal expected exists; duration out of range (lower); duration expected missing",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = nil
						datum.Extended = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo(-1, 0), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/expectedDuration", NewMeta()),
				),
				Entry("normal expected exists; duration out of range (lower); duration expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(-1)
						datum.Extended = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo(-1, 0), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("normal expected exists; duration out of range (lower); duration expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo(-1, 0), "/duration", NewMeta()),
				),
				Entry("normal expected exists; duration out of range (lower); duration expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(86400000)
						datum.Extended = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo(-1, 0), "/duration", NewMeta()),
				),
				Entry("normal expected exists; duration out of range (lower); duration expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(86400001)
						datum.Extended = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo(-1, 0), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("normal expected exists; duration in range; duration expected missing",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = nil
						datum.Extended = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/expectedDuration", NewMeta()),
				),
				Entry("normal expected exists; duration in range; duration expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(-1)
						datum.Extended = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("normal expected exists; duration in range; duration expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
				),
				Entry("normal expected exists; duration in range; duration expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(86400000)
						datum.Extended = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
				),
				Entry("normal expected exists; duration in range; duration expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(86400001)
						datum.Extended = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("normal expected exists; duration out of range (upper); duration expected missing",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(1)
						datum.DurationExpected = nil
						datum.Extended = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo(1, 0), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/expectedDuration", NewMeta()),
				),
				Entry("normal expected exists; duration out of range (upper); duration expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(1)
						datum.DurationExpected = pointer.Int(-1)
						datum.Extended = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo(1, 0), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("normal expected exists; duration out of range (upper); duration expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(1)
						datum.DurationExpected = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo(1, 0), "/duration", NewMeta()),
				),
				Entry("normal expected exists; duration out of range (upper); duration expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(1)
						datum.DurationExpected = pointer.Int(86400000)
						datum.Extended = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo(1, 0), "/duration", NewMeta()),
				),
				Entry("normal expected exists; duration out of range (upper); duration expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(1)
						datum.DurationExpected = pointer.Int(86400001)
						datum.Extended = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo(1, 0), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("normal expected exists; extended missing; extended expected missing",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = nil
						datum.ExtendedExpected = nil
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/expectedExtended", NewMeta()),
				),
				Entry("normal expected exists; extended missing; extended expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = nil
						datum.ExtendedExpected = pointer.Float64(-0.1)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("normal expected exists; extended missing; extended expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = nil
						datum.ExtendedExpected = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
				),
				Entry("normal expected exists; extended missing; extended expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = nil
						datum.ExtendedExpected = pointer.Float64(100.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
				),
				Entry("normal expected exists; extended missing; extended expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = nil
						datum.ExtendedExpected = pointer.Float64(100.1)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("normal expected exists; extended out of range (lower); extended expected missing",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(-0.1)
						datum.ExtendedExpected = nil
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo(-0.1, 0.0), "/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/expectedExtended", NewMeta()),
				),
				Entry("normal expected exists; extended out of range (lower); extended expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(-0.1)
						datum.ExtendedExpected = pointer.Float64(-0.1)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo(-0.1, 0.0), "/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("normal expected exists; extended out of range (lower); extended expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(-0.1)
						datum.ExtendedExpected = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo(-0.1, 0.0), "/extended", NewMeta()),
				),
				Entry("normal expected exists; extended out of range (lower); extended expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(-0.1)
						datum.ExtendedExpected = pointer.Float64(100.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo(-0.1, 0.0), "/extended", NewMeta()),
				),
				Entry("normal expected exists; extended out of range (lower); extended expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(-0.1)
						datum.ExtendedExpected = pointer.Float64(100.1)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo(-0.1, 0.0), "/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("normal expected exists; extended in range; extended expected missing",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.ExtendedExpected = nil
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/expectedExtended", NewMeta()),
				),
				Entry("normal expected exists; extended in range; extended expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.ExtendedExpected = pointer.Float64(-0.1)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("normal expected exists; extended in range; extended expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.ExtendedExpected = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
				),
				Entry("normal expected exists; extended in range; extended expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.ExtendedExpected = pointer.Float64(100.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
				),
				Entry("normal expected exists; extended in range; extended expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.ExtendedExpected = pointer.Float64(100.1)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("normal expected exists; extended out of range (upper); extended expected missing",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.1)
						datum.ExtendedExpected = nil
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo(0.1, 0.0), "/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/expectedExtended", NewMeta()),
				),
				Entry("normal expected exists; extended out of range (upper); extended expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.1)
						datum.ExtendedExpected = pointer.Float64(-0.1)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo(0.1, 0.0), "/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("normal expected exists; extended out of range (upper); extended expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.1)
						datum.ExtendedExpected = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo(0.1, 0.0), "/extended", NewMeta()),
				),
				Entry("normal expected exists; extended out of range (upper); extended expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.1)
						datum.ExtendedExpected = pointer.Float64(100.0)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo(0.1, 0.0), "/extended", NewMeta()),
				),
				Entry("normal expected exists; extended out of range (upper); extended expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.1)
						datum.ExtendedExpected = pointer.Float64(100.1)
						datum.NormalExpected = pointer.Float64(test.RandomFloat64FromRange(*datum.Normal, combination.NormalMaximum))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo(0.1, 0.0), "/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("normal missing; normal expected missing",
					func(datum *combination.Combination) {
						datum.Normal = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
				),
				Entry("normal missing; normal expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.Normal = nil
						datum.NormalExpected = pointer.Float64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal missing; normal expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.Normal = nil
						datum.NormalExpected = pointer.Float64(0.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
				),
				Entry("normal missing; normal expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.Normal = nil
						datum.NormalExpected = pointer.Float64(100.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
				),
				Entry("normal missing; normal expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.Normal = nil
						datum.NormalExpected = pointer.Float64(100.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected missing",
					func(datum *combination.Combination) {
						datum.Normal = pointer.Float64(-0.1)
						datum.NormalExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.Normal = pointer.Float64(-0.1)
						datum.NormalExpected = pointer.Float64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.Normal = pointer.Float64(-0.1)
						datum.NormalExpected = pointer.Float64(0.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.Normal = pointer.Float64(-0.1)
						datum.NormalExpected = pointer.Float64(100.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (lower); normal expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.Normal = pointer.Float64(-0.1)
						datum.NormalExpected = pointer.Float64(100.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (lower); normal expected missing",
					func(datum *combination.Combination) {
						datum.Normal = pointer.Float64(0.0)
						datum.NormalExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (lower); normal expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.Normal = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (lower); normal expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.Normal = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(0.0)
					},
				),
				Entry("normal in range (lower); normal expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.Normal = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(100.0)
					},
				),
				Entry("normal in range (lower); normal expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.Normal = pointer.Float64(0.0)
						datum.NormalExpected = pointer.Float64(100.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (upper); normal expected missing",
					func(datum *combination.Combination) {
						datum.Normal = pointer.Float64(100.0)
						datum.NormalExpected = nil
					},
				),
				Entry("normal in range (upper); normal expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.Normal = pointer.Float64(100.0)
						datum.NormalExpected = pointer.Float64(99.9)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(99.9, 100.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal in range (upper); normal expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.Normal = pointer.Float64(100.0)
						datum.NormalExpected = pointer.Float64(100.0)
					},
				),
				Entry("normal in range (upper); normal expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.Normal = pointer.Float64(100.0)
						datum.NormalExpected = pointer.Float64(100.0)
					},
				),
				Entry("normal in range (upper); normal expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.Normal = pointer.Float64(100.0)
						datum.NormalExpected = pointer.Float64(100.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 100.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected missing",
					func(datum *combination.Combination) {
						datum.Normal = pointer.Float64(100.1)
						datum.NormalExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected out of range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.Normal = pointer.Float64(100.1)
						datum.NormalExpected = pointer.Float64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected in range (lower)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.Normal = pointer.Float64(100.1)
						datum.NormalExpected = pointer.Float64(0.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected in range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.Normal = pointer.Float64(100.1)
						datum.NormalExpected = pointer.Float64(100.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta()),
				),
				Entry("normal out of range (upper); normal expected out of range (upper)",
					func(datum *combination.Combination) {
						datum.Duration = pointer.Int(0)
						datum.Extended = pointer.Float64(0.0)
						datum.Normal = pointer.Float64(100.1)
						datum.NormalExpected = pointer.Float64(100.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedNormal", NewMeta()),
				),
				Entry("multiple errors",
					func(datum *combination.Combination) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(86400001)
						datum.Extended = nil
						datum.ExtendedExpected = pointer.Float64(100.1)
						datum.Normal = nil
						datum.NormalExpected = pointer.Float64(100.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "bolus"), "/type", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "dual/square"), "/subType", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0, 100), "/expectedExtended", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0, 100), "/expectedNormal", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *combination.Combination)) {
					for _, origin := range structure.Origins() {
						datum := testDataTypesBolusCombination.NewCombination()
						mutator(datum)
						expectedDatum := testDataTypesBolusCombination.CloneCombination(datum)
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
