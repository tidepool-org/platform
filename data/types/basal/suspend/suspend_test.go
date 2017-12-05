package suspend_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/factory"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/data/types/basal/suspend"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/service"
)

func NewMeta() interface{} {
	return &basal.Meta{
		Type:         "basal",
		DeliveryType: "suspend",
	}
}

func NewTestSuspend(sourceTime interface{}, sourceDuration interface{}, sourceExpectedDuration interface{}, sourceSuppressed *basal.Suppressed) *suspend.Suspend {
	testSuspend := suspend.Init()
	testSuspend.DeviceID = pointer.String(id.New())
	if value, ok := sourceTime.(string); ok {
		testSuspend.Time = pointer.String(value)
	}
	if value, ok := sourceDuration.(int); ok {
		testSuspend.Duration = pointer.Int(value)
	}
	if value, ok := sourceExpectedDuration.(int); ok {
		testSuspend.ExpectedDuration = pointer.Int(value)
	}
	testSuspend.Suppressed = sourceSuppressed
	return testSuspend
}

func NewTestSuppressed(sourceType interface{}, sourceDeliveryType interface{}, sourceRate interface{}, sourceScheduleName interface{}, sourceSuppressed *basal.Suppressed) *basal.Suppressed {
	testSuppressed := &basal.Suppressed{}
	if value, ok := sourceType.(string); ok {
		testSuppressed.Type = pointer.String(value)
	}
	if value, ok := sourceDeliveryType.(string); ok {
		testSuppressed.DeliveryType = pointer.String(value)
	}
	if value, ok := sourceRate.(float64); ok {
		testSuppressed.Rate = pointer.Float64(value)
	}
	if value, ok := sourceScheduleName.(string); ok {
		testSuppressed.ScheduleName = pointer.String(value)
	}
	testSuppressed.Suppressed = sourceSuppressed
	return testSuppressed
}

var _ = Describe("Suspend", func() {
	Context("DeliveryType", func() {
		It("returns the expected type", func() {
			Expect(suspend.DeliveryType()).To(Equal("suspend"))
		})
	})

	Context("NewDatum", func() {
		It("returns the expected datum", func() {
			Expect(suspend.NewDatum()).To(Equal(&suspend.Suspend{}))
		})
	})

	Context("New", func() {
		It("returns the expected suspend", func() {
			Expect(suspend.New()).To(Equal(&suspend.Suspend{}))
		})
	})

	Context("Init", func() {
		It("returns the expected suspend with all values initialized", func() {
			testSuspend := suspend.Init()
			Expect(testSuspend).ToNot(BeNil())
			Expect(testSuspend.ID).ToNot(BeEmpty())
			Expect(testSuspend.Type).To(Equal("basal"))
			Expect(testSuspend.DeliveryType).To(Equal("suspend"))
			Expect(testSuspend.Duration).To(BeNil())
			Expect(testSuspend.ExpectedDuration).To(BeNil())
			Expect(testSuspend.Suppressed).To(BeNil())
		})
	})

	Context("with new suspend", func() {
		var testSuspend *suspend.Suspend

		BeforeEach(func() {
			testSuspend = suspend.New()
			Expect(testSuspend).ToNot(BeNil())
		})

		Context("Init", func() {
			It("initializes the suspend", func() {
				testSuspend.Init()
				Expect(testSuspend.ID).ToNot(BeEmpty())
				Expect(testSuspend.Type).To(Equal("basal"))
				Expect(testSuspend.DeliveryType).To(Equal("suspend"))
				Expect(testSuspend.Duration).To(BeNil())
				Expect(testSuspend.ExpectedDuration).To(BeNil())
				Expect(testSuspend.Suppressed).To(BeNil())
			})
		})

		Context("with initialized", func() {
			BeforeEach(func() {
				testSuspend.Init()
			})

			DescribeTable("Parse",
				func(sourceObject *map[string]interface{}, expectedSuspend *suspend.Suspend, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(null.NewLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testFactory, err := factory.NewStandard()
					Expect(err).ToNot(HaveOccurred())
					Expect(testFactory).ToNot(BeNil())
					testParser, err := parser.NewStandardObject(testContext, testFactory, sourceObject, parser.AppendErrorNotParsed)
					Expect(err).ToNot(HaveOccurred())
					Expect(testParser).ToNot(BeNil())
					Expect(testSuspend.Parse(testParser)).To(Succeed())
					Expect(testSuspend.Time).To(Equal(expectedSuspend.Time))
					Expect(testSuspend.Duration).To(Equal(expectedSuspend.Duration))
					Expect(testSuspend.ExpectedDuration).To(Equal(expectedSuspend.ExpectedDuration))
					Expect(testSuspend.Suppressed).To(Equal(expectedSuspend.Suppressed))
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("parses object that is nil",
					nil,
					NewTestSuspend(nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that is empty",
					&map[string]interface{}{},
					NewTestSuspend(nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has valid time",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00"},
					NewTestSuspend("2016-09-06T13:45:58-07:00", nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid time",
					&map[string]interface{}{"time": 0},
					NewTestSuspend(nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
					}),
				Entry("parses object that has valid duration",
					&map[string]interface{}{"duration": 3600000},
					NewTestSuspend(nil, 3600000, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid duration",
					&map[string]interface{}{"duration": "invalid"},
					NewTestSuspend(nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/duration", NewMeta()),
					}),
				Entry("parses object that has valid expected duration",
					&map[string]interface{}{"expectedDuration": 7200000},
					NewTestSuspend(nil, nil, 7200000, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid expected duration",
					&map[string]interface{}{"expectedDuration": "invalid"},
					NewTestSuspend(nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/expectedDuration", NewMeta()),
					}),
				Entry("parses object that has valid suppressed",
					&map[string]interface{}{"suppressed": map[string]interface{}{"type": "basal", "deliveryType": "scheduled", "rate": 1.0, "scheduleName": "Weekday"}},
					NewTestSuspend(nil, nil, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil)),
					[]*service.Error{}),
				Entry("parses object that has invalid suppressed",
					&map[string]interface{}{"suppressed": "invalid"},
					NewTestSuspend(nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotObject("invalid"), "/suppressed", NewMeta()),
					}),
				Entry("parses object that has multiple valid fields",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00", "duration": 3600000, "expectedDuration": 7200000},
					NewTestSuspend("2016-09-06T13:45:58-07:00", 3600000, 7200000, nil),
					[]*service.Error{}),
				Entry("parses object that has multiple invalid fields",
					&map[string]interface{}{"time": 0, "duration": "invalid", "expectedDuration": "invalid", "suppressed": "invalid"},
					NewTestSuspend(nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/duration", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/expectedDuration", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotObject("invalid"), "/suppressed", NewMeta()),
					}),
			)

			DescribeTable("Validate",
				func(sourceSuspend *suspend.Suspend, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(null.NewLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testValidator, err := validator.NewStandard(testContext)
					Expect(err).ToNot(HaveOccurred())
					Expect(testValidator).ToNot(BeNil())
					Expect(sourceSuspend.Validate(testValidator)).To(Succeed())
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("all valid",
					NewTestSuspend("2016-09-06T13:45:58-07:00", 3600000, 7200000, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil)),
					[]*service.Error{}),
				Entry("missing time",
					NewTestSuspend(nil, 3600000, 7200000, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotExists(), "/time", NewMeta()),
					}),
				Entry("missing duration",
					NewTestSuspend("2016-09-06T13:45:58-07:00", nil, 7200000, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotExists(), "/duration", NewMeta()),
					}),
				Entry("duration out of range (lower)",
					NewTestSuspend("2016-09-06T13:45:58-07:00", -1, 7200000, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
					}),
				Entry("duration in range (lower)",
					NewTestSuspend("2016-09-06T13:45:58-07:00", 0, 7200000, nil),
					[]*service.Error{}),
				Entry("duration in range (upper)",
					NewTestSuspend("2016-09-06T13:45:58-07:00", 604800000, nil, nil),
					[]*service.Error{}),
				Entry("duration out of range (upper)",
					NewTestSuspend("2016-09-06T13:45:58-07:00", 604800001, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
					}),
				Entry("expected duration out of range (lower)",
					NewTestSuspend("2016-09-06T13:45:58-07:00", 3600000, 3599999, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotInRange(3599999, 3600000, 604800000), "/expectedDuration", NewMeta()),
					}),
				Entry("expected duration in range (lower)",
					NewTestSuspend("2016-09-06T13:45:58-07:00", 3600000, 3600000, nil),
					[]*service.Error{}),
				Entry("expected duration in range (upper)",
					NewTestSuspend("2016-09-06T13:45:58-07:00", 3600000, 604800000, nil),
					[]*service.Error{}),
				Entry("expected duration out of range (upper)",
					NewTestSuspend("2016-09-06T13:45:58-07:00", 3600000, 604800001, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotInRange(604800001, 3600000, 604800000), "/expectedDuration", NewMeta()),
					}),
				Entry("missing duration; expected duration out of range (lower)",
					NewTestSuspend("2016-09-06T13:45:58-07:00", nil, -1, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotExists(), "/duration", NewMeta()),
						testData.ComposeError(service.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
					}),
				Entry("missing duration; expected duration in range (lower)",
					NewTestSuspend("2016-09-06T13:45:58-07:00", nil, 0, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotExists(), "/duration", NewMeta()),
					}),
				Entry("missing duration; expected duration in range (upper)",
					NewTestSuspend("2016-09-06T13:45:58-07:00", nil, 604800000, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotExists(), "/duration", NewMeta()),
					}),
				Entry("missing duration; expected duration out of range (upper)",
					NewTestSuspend("2016-09-06T13:45:58-07:00", nil, 604800001, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotExists(), "/duration", NewMeta()),
						testData.ComposeError(service.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
					}),
				Entry("multiple",
					NewTestSuspend(nil, nil, 604800001, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotExists(), "/time", NewMeta()),
						testData.ComposeError(service.ErrorValueNotExists(), "/duration", NewMeta()),
						testData.ComposeError(service.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
					}),
				Entry("suppressed scheduled",
					NewTestSuspend("2016-09-06T13:45:58-07:00", 3600000, 7200000, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil)),
					[]*service.Error{}),
				Entry("suppressed scheduled multiple",
					NewTestSuspend("2016-09-06T13:45:58-07:00", 3600000, 7200000, NewTestSuppressed("invalid", "scheduled", 100.1, "", NewTestSuppressed(nil, nil, nil, nil, nil))),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotEqualTo("invalid", "basal"), "/suppressed/type", NewMeta()),
						testData.ComposeError(service.ErrorValueNotInRange(100.1, 0.0, 100.0), "/suppressed/rate", NewMeta()),
						testData.ComposeError(service.ErrorValueEmpty(), "/suppressed/scheduleName", NewMeta()),
						testData.ComposeError(service.ErrorValueExists(), "/suppressed/suppressed", NewMeta()),
					}),
				Entry("suppressed temp with suppressed scheduled",
					NewTestSuspend("2016-09-06T13:45:58-07:00", 3600000, 7200000, NewTestSuppressed("basal", "temp", 2.0, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil))),
					[]*service.Error{}),
				Entry("suppressed temp with suppressed multiple",
					NewTestSuspend("2016-09-06T13:45:58-07:00", 3600000, 7200000, NewTestSuppressed("basal", "temp", 2.0, nil, NewTestSuppressed("invalid", "scheduled", 100.1, "", NewTestSuppressed(nil, nil, nil, nil, nil)))),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotEqualTo("invalid", "basal"), "/suppressed/suppressed/type", NewMeta()),
						testData.ComposeError(service.ErrorValueNotInRange(100.1, 0.0, 100.0), "/suppressed/suppressed/rate", NewMeta()),
						testData.ComposeError(service.ErrorValueEmpty(), "/suppressed/suppressed/scheduleName", NewMeta()),
						testData.ComposeError(service.ErrorValueExists(), "/suppressed/suppressed/suppressed", NewMeta()),
					}),
			)

			Context("Normalize", func() {
				It("succeeds", func() {
					testNormalizer := dataNormalizer.New()
					Expect(testNormalizer).ToNot(BeNil())
					testSuspend.Normalize(testNormalizer)
					Expect(testNormalizer.Error()).ToNot(HaveOccurred())
				})
			})
		})
	})
})
