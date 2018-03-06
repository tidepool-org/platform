package extended_test

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
	"github.com/tidepool-org/platform/data/types/bolus/extended"
	testDataTypesBolusExtended "github.com/tidepool-org/platform/data/types/bolus/extended/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewMeta() interface{} {
	return &bolus.Meta{
		Type:    "bolus",
		SubType: "square",
	}
}

func NewTestExtended(sourceTime interface{}, sourceDuration interface{}, sourceDurationExpected interface{}, sourceExtended interface{}, sourceExtendedExpected interface{}) *extended.Extended {
	datum := extended.Init()
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
	return datum
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
		Expect(extended.ExtendedMaximum).To(Equal(100.0))
	})

	It("ExtendedMinimum is expected", func() {
		Expect(extended.ExtendedMinimum).To(Equal(0.0))
	})

	Context("NewDatum", func() {
		It("returns the expected datum", func() {
			Expect(extended.NewDatum()).To(Equal(&extended.Extended{}))
		})
	})

	Context("New", func() {
		It("returns the expected datum", func() {
			Expect(extended.New()).To(Equal(&extended.Extended{}))
		})
	})

	Context("Init", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := extended.Init()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("bolus"))
			Expect(datum.SubType).To(Equal("square"))
			Expect(datum.Duration).To(BeNil())
			Expect(datum.DurationExpected).To(BeNil())
			Expect(datum.Extended).To(BeNil())
			Expect(datum.ExtendedExpected).To(BeNil())
		})
	})

	Context("with new datum", func() {
		var datum *extended.Extended

		BeforeEach(func() {
			datum = testDataTypesBolusExtended.NewExtended()
		})

		Context("Init", func() {
			It("initializes the datum", func() {
				datum.Init()
				Expect(datum.Type).To(Equal("bolus"))
				Expect(datum.SubType).To(Equal("square"))
				Expect(datum.Duration).To(BeNil())
				Expect(datum.DurationExpected).To(BeNil())
				Expect(datum.Extended).To(BeNil())
				Expect(datum.ExtendedExpected).To(BeNil())
			})
		})
	})

	Context("Extended", func() {
		Context("Parse", func() {
			var datum *extended.Extended

			BeforeEach(func() {
				datum = extended.Init()
				Expect(datum).ToNot(BeNil())
			})

			DescribeTable("parses the datum",
				func(sourceObject *map[string]interface{}, expectedDatum *extended.Extended, expectedErrors []*service.Error) {
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
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("parses object that is nil",
					nil,
					NewTestExtended(nil, nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that is empty",
					&map[string]interface{}{},
					NewTestExtended(nil, nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has valid time",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00"},
					NewTestExtended("2016-09-06T13:45:58-07:00", nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid time",
					&map[string]interface{}{"time": 0},
					NewTestExtended(nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
					}),
				Entry("parses object that has valid duration",
					&map[string]interface{}{"duration": 1000000},
					NewTestExtended(nil, 1000000, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid duration",
					&map[string]interface{}{"duration": "invalid"},
					NewTestExtended(nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/duration", NewMeta()),
					}),
				Entry("parses object that has valid duration expected",
					&map[string]interface{}{"expectedDuration": 2000000},
					NewTestExtended(nil, nil, 2000000, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid duration expected",
					&map[string]interface{}{"expectedDuration": "invalid"},
					NewTestExtended(nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/expectedDuration", NewMeta()),
					}),
				Entry("parses object that has valid extended",
					&map[string]interface{}{"extended": 3.6},
					NewTestExtended(nil, nil, nil, 3.6, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid extended",
					&map[string]interface{}{"extended": "invalid"},
					NewTestExtended(nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/extended", NewMeta()),
					}),
				Entry("parses object that has valid extended expected",
					&map[string]interface{}{"expectedExtended": 7.2},
					NewTestExtended(nil, nil, nil, nil, 7.2),
					[]*service.Error{}),
				Entry("parses object that has invalid extended expected",
					&map[string]interface{}{"expectedExtended": "invalid"},
					NewTestExtended(nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/expectedExtended", NewMeta()),
					}),
				Entry("parses object that has multiple valid fields",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00", "duration": 1000000, "expectedDuration": 2000000, "extended": 3.6, "expectedExtended": 7.2},
					NewTestExtended("2016-09-06T13:45:58-07:00", 1000000, 2000000, 3.6, 7.2),
					[]*service.Error{}),
				Entry("parses object that has multiple invalid fields",
					&map[string]interface{}{"time": 0, "duration": "invalid", "expectedDuration": "invalid", "extended": "invalid", "expectedExtended": "invalid"},
					NewTestExtended(nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/duration", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/expectedDuration", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/extended", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/expectedExtended", NewMeta()),
					}),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *extended.Extended), expectedErrors ...error) {
					datum := testDataTypesBolusExtended.NewExtended()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *extended.Extended) {},
				),
				Entry("type missing",
					func(datum *extended.Extended) { datum.Type = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &bolus.Meta{SubType: "square"}),
				),
				Entry("type invalid",
					func(datum *extended.Extended) { datum.Type = "invalidType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "bolus"), "/type", &bolus.Meta{Type: "invalidType", SubType: "square"}),
				),
				Entry("type bolus",
					func(datum *extended.Extended) { datum.Type = "bolus" },
				),
				Entry("sub type missing",
					func(datum *extended.Extended) { datum.SubType = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &bolus.Meta{Type: "bolus"}),
				),
				Entry("sub type invalid",
					func(datum *extended.Extended) { datum.SubType = "invalidSubType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "square"), "/subType", &bolus.Meta{Type: "bolus", SubType: "invalidSubType"}),
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
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (lower)",
					func(datum *extended.Extended) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (lower)",
					func(datum *extended.Extended) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (upper)",
					func(datum *extended.Extended) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(86400000)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (upper)",
					func(datum *extended.Extended) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(86400001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected missing",
					func(datum *extended.Extended) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (lower)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (lower)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (upper)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(86400000)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (upper)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(86400001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected missing",
					func(datum *extended.Extended) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
				),
				Entry("duration in range (lower); duration expected out of range (lower)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected in range (lower)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(0)
					},
				),
				Entry("duration in range (lower); duration expected in range (upper)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(86400000)
					},
				),
				Entry("duration in range (lower); duration expected out of range (upper)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(86400001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected missing",
					func(datum *extended.Extended) {
						datum.Duration = pointer.Int(86400000)
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
				),
				Entry("duration in range (upper); duration expected out of range (lower)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.Int(86400000)
						datum.DurationExpected = pointer.Int(604799999)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604799999, 86400000, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected in range (lower)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.Int(86400000)
						datum.DurationExpected = pointer.Int(86400000)
					},
				),
				Entry("duration in range (upper); duration expected in range (upper)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.Int(86400000)
						datum.DurationExpected = pointer.Int(86400000)
					},
				),
				Entry("duration in range (upper); duration expected out of range (upper)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.Int(86400000)
						datum.DurationExpected = pointer.Int(86400001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 86400000, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected missing",
					func(datum *extended.Extended) {
						datum.Duration = pointer.Int(86400001)
						datum.DurationExpected = nil
						datum.ExtendedExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected out of range (lower)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.Int(86400001)
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected in range (lower)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.Int(86400001)
						datum.DurationExpected = pointer.Int(0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected in range (upper)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.Int(86400001)
						datum.DurationExpected = pointer.Int(86400000)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected out of range (upper)",
					func(datum *extended.Extended) {
						datum.Duration = pointer.Int(86400001)
						datum.DurationExpected = pointer.Int(86400001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", NewMeta()),
				),
				Entry("extended missing; extended expected missing",
					func(datum *extended.Extended) {
						datum.DurationExpected = nil
						datum.Extended = nil
						datum.ExtendedExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
				),
				Entry("extended missing; extended expected out of range (lower)",
					func(datum *extended.Extended) {
						datum.Extended = nil
						datum.ExtendedExpected = pointer.Float64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended missing; extended expected in range (lower)",
					func(datum *extended.Extended) {
						datum.Extended = nil
						datum.ExtendedExpected = pointer.Float64(0.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
				),
				Entry("extended missing; extended expected in range (upper)",
					func(datum *extended.Extended) {
						datum.Extended = nil
						datum.ExtendedExpected = pointer.Float64(100.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
				),
				Entry("extended missing; extended expected out of range (upper)",
					func(datum *extended.Extended) {
						datum.Extended = nil
						datum.ExtendedExpected = pointer.Float64(100.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended out of range (lower); extended expected missing",
					func(datum *extended.Extended) {
						datum.DurationExpected = nil
						datum.Extended = pointer.Float64(-0.1)
						datum.ExtendedExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/extended", NewMeta()),
				),
				Entry("extended out of range (lower); extended expected out of range (lower)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.Float64(-0.1)
						datum.ExtendedExpected = pointer.Float64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended out of range (lower); extended expected in range (lower)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.Float64(-0.1)
						datum.ExtendedExpected = pointer.Float64(0.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/extended", NewMeta()),
				),
				Entry("extended out of range (lower); extended expected in range (upper)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.Float64(-0.1)
						datum.ExtendedExpected = pointer.Float64(100.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/extended", NewMeta()),
				),
				Entry("extended out of range (lower); extended expected out of range (upper)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.Float64(-0.1)
						datum.ExtendedExpected = pointer.Float64(100.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended in range (lower); extended expected missing",
					func(datum *extended.Extended) {
						datum.DurationExpected = nil
						datum.Extended = pointer.Float64(0.0)
						datum.ExtendedExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/expectedExtended", NewMeta()),
				),
				Entry("extended in range (lower); extended expected out of range (lower)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.Float64(0.0)
						datum.ExtendedExpected = pointer.Float64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended in range (lower); extended expected in range (lower)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.Float64(0.0)
						datum.ExtendedExpected = pointer.Float64(0.0)
					},
				),
				Entry("extended in range (lower); extended expected in range (upper)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.Float64(0.0)
						datum.ExtendedExpected = pointer.Float64(100.0)
					},
				),
				Entry("extended in range (lower); extended expected out of range (upper)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.Float64(0.0)
						datum.ExtendedExpected = pointer.Float64(100.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended in range (upper); extended expected missing",
					func(datum *extended.Extended) {
						datum.DurationExpected = nil
						datum.Extended = pointer.Float64(100.0)
						datum.ExtendedExpected = nil
					},
				),
				Entry("extended in range (upper); extended expected out of range (lower)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.Float64(100.0)
						datum.ExtendedExpected = pointer.Float64(99.9)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(99.9, 100.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended in range (upper); extended expected in range (lower)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.Float64(100.0)
						datum.ExtendedExpected = pointer.Float64(100.0)
					},
				),
				Entry("extended in range (upper); extended expected in range (upper)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.Float64(100.0)
						datum.ExtendedExpected = pointer.Float64(100.0)
					},
				),
				Entry("extended in range (upper); extended expected out of range (upper)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.Float64(100.0)
						datum.ExtendedExpected = pointer.Float64(100.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 100.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended out of range (upper); extended expected missing",
					func(datum *extended.Extended) {
						datum.DurationExpected = nil
						datum.Extended = pointer.Float64(100.1)
						datum.ExtendedExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/extended", NewMeta()),
				),
				Entry("extended out of range (upper); extended expected out of range (lower)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.Float64(100.1)
						datum.ExtendedExpected = pointer.Float64(-0.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
				),
				Entry("extended out of range (upper); extended expected in range (lower)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.Float64(100.1)
						datum.ExtendedExpected = pointer.Float64(0.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/extended", NewMeta()),
				),
				Entry("extended out of range (upper); extended expected in range (upper)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.Float64(100.1)
						datum.ExtendedExpected = pointer.Float64(100.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/extended", NewMeta()),
				),
				Entry("extended out of range (upper); extended expected out of range (upper)",
					func(datum *extended.Extended) {
						datum.Extended = pointer.Float64(100.1)
						datum.ExtendedExpected = pointer.Float64(100.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/extended", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedExtended", NewMeta()),
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
						datum.ExtendedExpected = pointer.Float64(100.0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/expectedDuration", NewMeta()),
				),
				Entry("duration exists; extended expected missing",
					func(datum *extended.Extended) {
						datum.DurationExpected = pointer.Int(86400000)
						datum.ExtendedExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/expectedDuration", NewMeta()),
				),
				Entry("duration exists; extended expected exists",
					func(datum *extended.Extended) {
						datum.DurationExpected = pointer.Int(86400000)
						datum.ExtendedExpected = pointer.Float64(100.0)
					},
				),
				Entry("multiple errors",
					func(datum *extended.Extended) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(86400001)
						datum.Extended = nil
						datum.ExtendedExpected = pointer.Float64(100.1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "bolus"), "/type", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "square"), "/subType", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/extended", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0, 100), "/expectedExtended", &bolus.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *extended.Extended)) {
					for _, origin := range structure.Origins() {
						datum := testDataTypesBolusExtended.NewExtended()
						mutator(datum)
						expectedDatum := testDataTypesBolusExtended.CloneExtended(datum)
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
