package basal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/factory"
	"github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	"github.com/tidepool-org/platform/data/types/base/basal"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

func NewMeta(deliveryType string) interface{} {
	return &basal.Meta{
		Type:         "basal",
		DeliveryType: deliveryType,
	}
}

func NewTestBasal(sourceTime interface{}, sourceDeliveryType string) *basal.Basal {
	testBasal := &basal.Basal{}
	testBasal.Init()
	testBasal.DeviceID = app.StringAsPointer(app.NewID())
	if value, ok := sourceTime.(string); ok {
		testBasal.Time = app.StringAsPointer(value)
	}
	testBasal.DeliveryType = sourceDeliveryType
	return testBasal
}

var _ = Describe("Basal", func() {
	Context("Type", func() {
		It("returns the expected type", func() {
			Expect(basal.Type()).To(Equal("basal"))
		})
	})

	Context("with new basal", func() {
		var testBasal *basal.Basal

		BeforeEach(func() {
			testBasal = &basal.Basal{}
		})

		Context("Init", func() {
			It("initializes the basal", func() {
				testBasal.Init()
				Expect(testBasal.ID).ToNot(BeEmpty())
				Expect(testBasal.Type).To(Equal("basal"))
				Expect(testBasal.DeliveryType).To(BeEmpty())
			})
		})

		Context("with initialized", func() {
			BeforeEach(func() {
				testBasal.Init()
			})

			DescribeTable("Parse",
				func(sourceObject *map[string]interface{}, expectedBasal *basal.Basal, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(log.NewNull())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testFactory, err := factory.NewStandard()
					Expect(err).ToNot(HaveOccurred())
					Expect(testFactory).ToNot(BeNil())
					testParser, err := parser.NewStandardObject(testContext, testFactory, sourceObject, parser.AppendErrorNotParsed)
					Expect(err).ToNot(HaveOccurred())
					Expect(testParser).ToNot(BeNil())
					Expect(testBasal.Parse(testParser)).To(Succeed())
					Expect(testBasal.Time).To(Equal(expectedBasal.Time))
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("parses object that is nil",
					nil,
					NewTestBasal(nil, ""),
					[]*service.Error{}),
				Entry("parses object that is empty",
					&map[string]interface{}{},
					NewTestBasal(nil, ""),
					[]*service.Error{}),
				Entry("parses object that has valid time",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00"},
					NewTestBasal("2016-09-06T13:45:58-07:00", ""),
					[]*service.Error{}),
				Entry("parses object that has invalid time",
					&map[string]interface{}{"time": 0},
					NewTestBasal(nil, ""),
					[]*service.Error{
						testing.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta("")),
					}),
			)

			DescribeTable("Validate",
				func(sourceBasal *basal.Basal, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(log.NewNull())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testValidator, err := validator.NewStandard(testContext)
					Expect(err).ToNot(HaveOccurred())
					Expect(testValidator).ToNot(BeNil())
					Expect(sourceBasal.Validate(testValidator)).To(Succeed())
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("all valid",
					NewTestBasal("2016-09-06T13:45:58-07:00", "test"),
					[]*service.Error{}),
				Entry("missing time",
					NewTestBasal(nil, "test"),
					[]*service.Error{
						testing.ComposeError(service.ErrorValueNotExists(), "/time", NewMeta("test")),
					}),
				Entry("delivery type empty",
					NewTestBasal("2016-09-06T13:45:58-07:00", ""),
					[]*service.Error{
						testing.ComposeError(service.ErrorValueEmpty(), "/deliveryType", NewMeta("")),
					}),
			)

			Context("Normalize", func() {
				It("succeeds", func() {
					testContext, err := context.NewStandard(log.NewNull())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testNormalizer, err := normalizer.NewStandard(testContext)
					Expect(err).ToNot(HaveOccurred())
					Expect(testNormalizer).ToNot(BeNil())
					Expect(testBasal.Normalize(testNormalizer)).To(Succeed())
				})
			})
		})
	})
})
