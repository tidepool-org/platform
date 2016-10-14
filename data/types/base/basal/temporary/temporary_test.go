package temporary_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/factory"
	"github.com/tidepool-org/platform/data/parser"
	"github.com/tidepool-org/platform/data/types/base/basal"
	"github.com/tidepool-org/platform/data/types/base/basal/temporary"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

func NewMeta() interface{} {
	return &basal.Meta{
		Type:         "basal",
		DeliveryType: "temp",
	}
}

func NewTestTemporary(sourceTime interface{}, sourceDuration interface{}, sourceExpectedDuration interface{}, sourceRate interface{}, sourcePercent interface{}) *temporary.Temporary {
	testTemporary := temporary.Init()
	testTemporary.DeviceID = app.StringAsPointer(app.NewID())
	if value, ok := sourceTime.(string); ok {
		testTemporary.Time = app.StringAsPointer(value)
	}
	if value, ok := sourceDuration.(int); ok {
		testTemporary.Duration = app.IntegerAsPointer(value)
	}
	if value, ok := sourceExpectedDuration.(int); ok {
		testTemporary.ExpectedDuration = app.IntegerAsPointer(value)
	}
	if value, ok := sourceRate.(float64); ok {
		testTemporary.Rate = app.FloatAsPointer(value)
	}
	if value, ok := sourcePercent.(float64); ok {
		testTemporary.Percent = app.FloatAsPointer(value)
	}
	return testTemporary
}

var _ = Describe("Temporary", func() {
	Context("DeliveryType", func() {
		It("returns the expected type", func() {
			Expect(temporary.DeliveryType()).To(Equal("temp"))
		})
	})

	Context("NewDatum", func() {
		It("returns the expected datum", func() {
			Expect(temporary.NewDatum()).To(Equal(&temporary.Temporary{}))
		})
	})

	Context("New", func() {
		It("returns the expected temporary", func() {
			Expect(temporary.New()).To(Equal(&temporary.Temporary{}))
		})
	})

	Context("Init", func() {
		It("returns the expected temporary with all values initialized", func() {
			testTemporary := temporary.Init()
			Expect(testTemporary).ToNot(BeNil())
			Expect(testTemporary.Type).To(Equal("basal"))
			Expect(testTemporary.DeliveryType).To(Equal("temp"))
			Expect(testTemporary.Duration).To(BeNil())
			Expect(testTemporary.ExpectedDuration).To(BeNil())
			Expect(testTemporary.Rate).To(BeNil())
			Expect(testTemporary.Percent).To(BeNil())
		})
	})

	Context("with new temporary", func() {
		var testTemporary *temporary.Temporary

		BeforeEach(func() {
			testTemporary = temporary.New()
			Expect(testTemporary).ToNot(BeNil())
		})

		Context("Init", func() {
			It("initializes the temporary", func() {
				testTemporary.Init()
				Expect(testTemporary.Type).To(Equal("basal"))
				Expect(testTemporary.DeliveryType).To(Equal("temp"))
				Expect(testTemporary.Duration).To(BeNil())
				Expect(testTemporary.ExpectedDuration).To(BeNil())
				Expect(testTemporary.Rate).To(BeNil())
				Expect(testTemporary.Percent).To(BeNil())
			})
		})

		Context("with initialized", func() {
			BeforeEach(func() {
				testTemporary.Init()
			})

			DescribeTable("Parse",
				func(sourceObject *map[string]interface{}, expectedTemporary *temporary.Temporary, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(log.NewNull())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testFactory, err := factory.NewStandard()
					Expect(err).ToNot(HaveOccurred())
					Expect(testFactory).ToNot(BeNil())
					testParser, err := parser.NewStandardObject(testContext, testFactory, sourceObject, parser.AppendErrorNotParsed)
					Expect(err).ToNot(HaveOccurred())
					Expect(testParser).ToNot(BeNil())
					Expect(testTemporary.Parse(testParser)).To(Succeed())
					Expect(testTemporary.Time).To(Equal(expectedTemporary.Time))
					Expect(testTemporary.Duration).To(Equal(expectedTemporary.Duration))
					Expect(testTemporary.ExpectedDuration).To(Equal(expectedTemporary.ExpectedDuration))
					Expect(testTemporary.Rate).To(Equal(expectedTemporary.Rate))
					Expect(testTemporary.Percent).To(Equal(expectedTemporary.Percent))
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("parses object that is nil", nil, NewTestTemporary(nil, nil, nil, nil, nil), []*service.Error{}),
				Entry("parses object that is empty", &map[string]interface{}{}, NewTestTemporary(nil, nil, nil, nil, nil), []*service.Error{}),
				Entry("parses object that has valid time", &map[string]interface{}{"time": "2016-09-06T13:45:58-07:00"}, NewTestTemporary("2016-09-06T13:45:58-07:00", nil, nil, nil, nil), []*service.Error{}),
				Entry("parses object that has invalid time", &map[string]interface{}{"time": 0}, NewTestTemporary(nil, nil, nil, nil, nil), []*service.Error{
					testing.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
				}),
				Entry("parses object that has valid duration", &map[string]interface{}{"duration": 3600000}, NewTestTemporary(nil, 3600000, nil, nil, nil), []*service.Error{}),
				Entry("parses object that has invalid duration", &map[string]interface{}{"duration": "invalid"}, NewTestTemporary(nil, nil, nil, nil, nil), []*service.Error{
					testing.ComposeError(service.ErrorTypeNotInteger("invalid"), "/duration", NewMeta()),
				}),
				Entry("parses object that has valid expected duration", &map[string]interface{}{"expectedDuration": 7200000}, NewTestTemporary(nil, nil, 7200000, nil, nil), []*service.Error{}),
				Entry("parses object that has invalid expected duration", &map[string]interface{}{"expectedDuration": "invalid"}, NewTestTemporary(nil, nil, nil, nil, nil), []*service.Error{
					testing.ComposeError(service.ErrorTypeNotInteger("invalid"), "/expectedDuration", NewMeta()),
				}),
				Entry("parses object that has valid rate", &map[string]interface{}{"rate": 1.0}, NewTestTemporary(nil, nil, nil, 1.0, nil), []*service.Error{}),
				Entry("parses object that has invalid rate", &map[string]interface{}{"rate": "invalid"}, NewTestTemporary(nil, nil, nil, nil, nil), []*service.Error{
					testing.ComposeError(service.ErrorTypeNotFloat("invalid"), "/rate", NewMeta()),
				}),
				Entry("parses object that has valid percent", &map[string]interface{}{"percent": 0.5}, NewTestTemporary(nil, nil, nil, nil, 0.5), []*service.Error{}),
				Entry("parses object that has invalid percent", &map[string]interface{}{"percent": "invalid"}, NewTestTemporary(nil, nil, nil, nil, nil), []*service.Error{
					testing.ComposeError(service.ErrorTypeNotFloat("invalid"), "/percent", NewMeta()),
				}),
				Entry("parses object that has multiple valid fields", &map[string]interface{}{"time": "2016-09-06T13:45:58-07:00", "duration": 3600000, "expectedDuration": 7200000, "rate": 1.0, "percent": 0.5},
					NewTestTemporary("2016-09-06T13:45:58-07:00", 3600000, 7200000, 1.0, 0.5), []*service.Error{}),
				Entry("parses object that has multiple invalid fields", &map[string]interface{}{"time": 0, "duration": "invalid", "expectedDuration": "invalid", "rate": "invalid", "percent": "invalid"},
					NewTestTemporary(nil, nil, nil, nil, nil), []*service.Error{
						testing.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
						testing.ComposeError(service.ErrorTypeNotInteger("invalid"), "/duration", NewMeta()),
						testing.ComposeError(service.ErrorTypeNotInteger("invalid"), "/expectedDuration", NewMeta()),
						testing.ComposeError(service.ErrorTypeNotFloat("invalid"), "/rate", NewMeta()),
						testing.ComposeError(service.ErrorTypeNotFloat("invalid"), "/percent", NewMeta()),
					}),
			)

			DescribeTable("Validate",
				func(sourceTemporary *temporary.Temporary, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(log.NewNull())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testValidator, err := validator.NewStandard(testContext)
					Expect(err).ToNot(HaveOccurred())
					Expect(testValidator).ToNot(BeNil())
					Expect(sourceTemporary.Validate(testValidator)).To(Succeed())
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("all valid", NewTestTemporary("2016-09-06T13:45:58-07:00", 3600000, 7200000, 1.0, 0.5), []*service.Error{}),
				Entry("missing time", NewTestTemporary(nil, 3600000, 7200000, 1.0, 0.5), []*service.Error{
					testing.ComposeError(service.ErrorValueNotExists(), "/time", NewMeta()),
				}),
				Entry("missing duration", NewTestTemporary("2016-09-06T13:45:58-07:00", nil, 7200000, 1.0, 0.5), []*service.Error{
					testing.ComposeError(service.ErrorValueNotExists(), "/duration", NewMeta()),
				}),
				Entry("duration out of range (lower)", NewTestTemporary("2016-09-06T13:45:58-07:00", -1, 7200000, 1.0, 0.5), []*service.Error{
					testing.ComposeError(service.ErrorValueNotInRange(-1, 0, 86400000), "/duration", NewMeta()),
				}),
				Entry("duration in range (lower)", NewTestTemporary("2016-09-06T13:45:58-07:00", 0, 7200000, 1.0, 0.5), []*service.Error{}),
				Entry("duration in range (upper)", NewTestTemporary("2016-09-06T13:45:58-07:00", 86400000, nil, 1.0, 0.5), []*service.Error{}),
				Entry("duration out of range (upper)", NewTestTemporary("2016-09-06T13:45:58-07:00", 86400001, nil, 1.0, 0.5), []*service.Error{
					testing.ComposeError(service.ErrorValueNotInRange(86400001, 0, 86400000), "/duration", NewMeta()),
				}),
				Entry("expected duration out of range (lower)", NewTestTemporary("2016-09-06T13:45:58-07:00", 3600000, 3599999, 1.0, 0.5), []*service.Error{
					testing.ComposeError(service.ErrorValueNotInRange(3599999, 3600000, 86400000), "/expectedDuration", NewMeta()),
				}),
				Entry("expected duration in range (lower)", NewTestTemporary("2016-09-06T13:45:58-07:00", 3600000, 3600000, 1.0, 0.5), []*service.Error{}),
				Entry("expected duration in range (upper)", NewTestTemporary("2016-09-06T13:45:58-07:00", 3600000, 86400000, 1.0, 0.5), []*service.Error{}),
				Entry("expected duration out of range (upper)", NewTestTemporary("2016-09-06T13:45:58-07:00", 3600000, 86400001, 1.0, 0.5), []*service.Error{
					testing.ComposeError(service.ErrorValueNotInRange(86400001, 3600000, 86400000), "/expectedDuration", NewMeta()),
				}),
				Entry("missing duration; expected duration out of range (lower)", NewTestTemporary("2016-09-06T13:45:58-07:00", nil, -1, 1.0, 0.5), []*service.Error{
					testing.ComposeError(service.ErrorValueNotExists(), "/duration", NewMeta()),
					testing.ComposeError(service.ErrorValueNotInRange(-1, 0, 86400000), "/expectedDuration", NewMeta()),
				}),
				Entry("missing duration; expected duration in range (lower)", NewTestTemporary("2016-09-06T13:45:58-07:00", nil, 0, 1.0, 0.5), []*service.Error{
					testing.ComposeError(service.ErrorValueNotExists(), "/duration", NewMeta()),
				}),
				Entry("missing duration; expected duration in range (upper)", NewTestTemporary("2016-09-06T13:45:58-07:00", nil, 86400000, 1.0, 0.5), []*service.Error{
					testing.ComposeError(service.ErrorValueNotExists(), "/duration", NewMeta()),
				}),
				Entry("missing duration; expected duration out of range (upper)", NewTestTemporary("2016-09-06T13:45:58-07:00", nil, 86400001, 1.0, 0.5), []*service.Error{
					testing.ComposeError(service.ErrorValueNotExists(), "/duration", NewMeta()),
					testing.ComposeError(service.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", NewMeta()),
				}),

				Entry("missing rate", NewTestTemporary("2016-09-06T13:45:58-07:00", 3600000, 7200000, nil, 0.5), []*service.Error{
					testing.ComposeError(service.ErrorValueNotExists(), "/rate", NewMeta()),
				}),
				Entry("rate out of range (lower)", NewTestTemporary("2016-09-06T13:45:58-07:00", 3600000, 7200000, -0.1, 0.5), []*service.Error{
					testing.ComposeError(service.ErrorValueNotInRange(-0.1, 0.0, 20.0), "/rate", NewMeta()),
				}),
				Entry("rate in range (lower)", NewTestTemporary("2016-09-06T13:45:58-07:00", 3600000, 7200000, 0.0, 0.5), []*service.Error{}),
				Entry("rate in range (upper)", NewTestTemporary("2016-09-06T13:45:58-07:00", 3600000, 7200000, 20.0, 0.5), []*service.Error{}),
				Entry("rate out of range (upper)", NewTestTemporary("2016-09-06T13:45:58-07:00", 3600000, 7200000, 20.1, 0.5), []*service.Error{
					testing.ComposeError(service.ErrorValueNotInRange(20.1, 0.0, 20.0), "/rate", NewMeta()),
				}),
				Entry("percent out of range (lower)", NewTestTemporary("2016-09-06T13:45:58-07:00", 3600000, 7200000, 1.0, -0.1), []*service.Error{
					testing.ComposeError(service.ErrorValueNotInRange(-0.1, 0.0, 10.0), "/percent", NewMeta()),
				}),
				Entry("percent in range (lower)", NewTestTemporary("2016-09-06T13:45:58-07:00", 3600000, 7200000, 1.0, 0.0), []*service.Error{}),
				Entry("percent in range (upper)", NewTestTemporary("2016-09-06T13:45:58-07:00", 3600000, 7200000, 1.0, 10.0), []*service.Error{}),
				Entry("percent out of range (upper)", NewTestTemporary("2016-09-06T13:45:58-07:00", 3600000, 7200000, 1.0, 10.1), []*service.Error{
					testing.ComposeError(service.ErrorValueNotInRange(10.1, 0.0, 10.0), "/percent", NewMeta()),
				}),
			)
		})
	})
})
