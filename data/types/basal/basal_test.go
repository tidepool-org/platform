package basal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/factory"
	"github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/service"
)

func NewMeta(deliveryType string) interface{} {
	return &basal.Meta{
		Type:         "basal",
		DeliveryType: deliveryType,
	}
}

func NewTestBasal(sourceTime interface{}, sourceDeliveryType interface{}) *basal.Basal {
	testBasal := &basal.Basal{}
	testBasal.Init()
	testBasal.DeviceID = pointer.String(id.New())
	if value, ok := sourceTime.(string); ok {
		testBasal.Time = pointer.String(value)
	}
	if value, ok := sourceDeliveryType.(string); ok {
		testBasal.DeliveryType = value
	}
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

			Context("Meta", func() {
				It("returns the meta with no delivery type", func() {
					testBasal.Init()
					Expect(testBasal.Meta()).To(Equal(NewMeta("")))
				})

				It("returns the meta with delivery type", func() {
					testBasal.Init()
					testBasal.DeliveryType = "scheduled"
					Expect(testBasal.Meta()).To(Equal(NewMeta("scheduled")))
				})
			})

			DescribeTable("Parse",
				func(sourceObject *map[string]interface{}, expectedBasal *basal.Basal, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(null.NewLogger())
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
					Expect(testBasal.DeliveryType).To(Equal(expectedBasal.DeliveryType))
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("parses object that is nil",
					nil,
					NewTestBasal(nil, nil),
					[]*service.Error{}),
				Entry("parses object that is empty",
					&map[string]interface{}{},
					NewTestBasal(nil, nil),
					[]*service.Error{}),
				Entry("parses object that has valid time",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00"},
					NewTestBasal("2016-09-06T13:45:58-07:00", nil),
					[]*service.Error{}),
				Entry("parses object that has invalid time",
					&map[string]interface{}{"time": 0},
					NewTestBasal(nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta("")),
					}),
				Entry("does not parse delivery type",
					&map[string]interface{}{"deliveryType": "scheduled"},
					NewTestBasal(nil, nil),
					[]*service.Error{}),
				Entry("parses object that has multiple valid fields",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00", "deliveryType": "scheduled"},
					NewTestBasal("2016-09-06T13:45:58-07:00", nil),
					[]*service.Error{}),
				Entry("parses object that has multiple invalid fields",
					&map[string]interface{}{"time": 0, "deliveryType": 0},
					NewTestBasal(nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta("")),
					}),
			)

			DescribeTable("Validate",
				func(sourceBasal *basal.Basal, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(null.NewLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testValidator, err := validator.NewStandard(testContext)
					Expect(err).ToNot(HaveOccurred())
					Expect(testValidator).ToNot(BeNil())
					Expect(sourceBasal.Validate(testValidator)).To(Succeed())
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("all valid",
					NewTestBasal("2016-09-06T13:45:58-07:00", "scheduled"),
					[]*service.Error{}),
				Entry("missing time",
					NewTestBasal(nil, "scheduled"),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotExists(), "/time", NewMeta("scheduled")),
					}),
				Entry("missing delivery type",
					NewTestBasal("2016-09-06T13:45:58-07:00", nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueEmpty(), "/deliveryType", NewMeta("")),
					}),
				Entry("specified delivery type",
					NewTestBasal("2016-09-06T13:45:58-07:00", "specified"),
					[]*service.Error{}),
				Entry("multiple",
					NewTestBasal(nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotExists(), "/time", NewMeta("")),
						testData.ComposeError(service.ErrorValueEmpty(), "/deliveryType", NewMeta("")),
					}),
			)

			Context("Normalize", func() {
				It("succeeds", func() {
					testContext, err := context.NewStandard(null.NewLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testNormalizer, err := normalizer.NewStandard(testContext)
					Expect(err).ToNot(HaveOccurred())
					Expect(testNormalizer).ToNot(BeNil())
					Expect(testBasal.Normalize(testNormalizer)).To(Succeed())
				})
			})

			Context("IdentityFields", func() {
				var userID string
				var deviceID string

				BeforeEach(func() {
					userID = id.New()
					deviceID = id.New()
					testBasal.UserID = userID
					testBasal.DeviceID = &deviceID
					testBasal.Time = pointer.String("2016-09-06T13:45:58-07:00")
					testBasal.DeliveryType = "scheduled"
				})

				It("returns error if user id is empty", func() {
					testBasal.UserID = ""
					identityFields, err := testBasal.IdentityFields()
					Expect(err).To(MatchError("user id is empty"))
					Expect(identityFields).To(BeEmpty())
				})

				It("returns error if delivery type is empty", func() {
					testBasal.DeliveryType = ""
					identityFields, err := testBasal.IdentityFields()
					Expect(err).To(MatchError("delivery type is empty"))
					Expect(identityFields).To(BeEmpty())
				})

				It("returns the expected identity fields", func() {
					identityFields, err := testBasal.IdentityFields()
					Expect(err).ToNot(HaveOccurred())
					Expect(identityFields).To(Equal([]string{userID, deviceID, "2016-09-06T13:45:58-07:00", "basal", "scheduled"}))
				})
			})
		})
	})
})
